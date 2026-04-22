package kernel

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"code.cloudfoundry.org/lager/v3"
	"github.com/gavmor/wasm-microkernel/host"
)

// Kernel manages a wasm-microkernel Engine and the set of loaded plugins.
type Kernel struct {
	ctx           context.Context
	engine        *host.Engine
	plugins       map[string][]byte
	mu            sync.Mutex
	ollamaURL     string
	transcribeURL string
}

// New creates a new wasm-microkernel Engine.
func New(ctx context.Context, logger lager.Logger, ollamaURL, transcribeURL string) (*Kernel, error) {
	engine := host.NewEngine()
	engine.AllowedHosts = []string{"*"} // Allow all hosts for simplicity, as plugins download audio, RSS, etc.
	engine.AllowedPaths = map[string]string{"/": "/", ".": "."}

	return &Kernel{
		ctx:           ctx,
		engine:        engine,
		plugins:       make(map[string][]byte),
		ollamaURL:     ollamaURL,
		transcribeURL: transcribeURL,
	}, nil
}

// Close shuts down the engine.
func (k *Kernel) Close() error { 
	return k.engine.Close(k.ctx) 
}

// Load reads a WASM plugin from disk and stores its bytes for execution.
func (k *Kernel) Load(name, path string) error {
	wasmBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read plugin %s: %w", name, err)
	}

	k.mu.Lock()
	k.plugins[name] = wasmBytes
	k.mu.Unlock()
	return nil
}

// Call JSON-encodes req, passes it to the named plugin using the Engine,
// and returns the OK payload bytes, or an error.
func (k *Kernel) Call(pluginName string, req interface{}) ([]byte, error) {
	k.mu.Lock()
	wasmBytes, ok := k.plugins[pluginName]
	k.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("plugin not loaded: %s", pluginName)
	}

	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	out, err := k.engine.Execute(k.ctx, wasmBytes, string(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("execute: %w", err)
	}

	return []byte(out), nil
}
