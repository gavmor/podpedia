// Package kernel loads WASM plugins using wazero and provides host
// capabilities (HTTP, LLM inference, file I/O) via Wasm imports.
package kernel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/lager/v3"
	"github.com/gavmor/wasm-microkernel/abi"
	kernelhost "github.com/gavmor/wasm-microkernel/host"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const HostModule = "podpedia_host"

// Kernel holds the wazero runtime and all loaded plugins.
type Kernel struct {
	ctx           context.Context
	runtime       wazero.Runtime
	logger        lager.Logger
	ollamaURL     string // threaded through ExtractRequest; plugins own the Ollama API logic
	transcribeURL string // threaded through TranscribeRequest; plugins own the ASR API logic
	plugins       map[string]*Plugin
}

// Plugin wraps an instantiated WASM module.
type Plugin struct {
	module api.Module
	name   string
}

// New creates a Kernel, instantiates WASI, and registers all host capabilities.
// ollamaURL and transcribeURL are threaded through plugin requests so plugins
// can own the API call logic; the host only provides generic http_post.
func New(ctx context.Context, logger lager.Logger, ollamaURL, transcribeURL string) (*Kernel, error) {
	r := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	k := &Kernel{
		ctx:           ctx,
		runtime:       r,
		logger:        logger,
		ollamaURL:     ollamaURL,
		transcribeURL: transcribeURL,
		plugins:       make(map[string]*Plugin),
	}

	if err := k.registerCapabilities(); err != nil {
		_ = r.Close(ctx)
		return nil, fmt.Errorf("registering host capabilities: %w", err)
	}
	return k, nil
}

// Load compiles and instantiates a WASM plugin from disk.
func (k *Kernel) Load(name, path string) error {
	wasmBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading plugin %s: %w", name, err)
	}
	compiled, err := k.runtime.CompileModule(k.ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("compiling plugin %s: %w", name, err)
	}
	cfg := wazero.NewModuleConfig().
		WithName(name).
		WithStartFunctions("_initialize")
	mod, err := k.runtime.InstantiateModule(k.ctx, compiled, cfg)
	if err != nil {
		return fmt.Errorf("instantiating plugin %s: %w", name, err)
	}
	k.plugins[name] = &Plugin{module: mod, name: name}
	k.logger.Info("plugin-loaded", lager.Data{"name": name})
	return nil
}

// Call invokes a plugin's Execute export with JSON-encoded input,
// returning JSON-encoded output.
func (k *Kernel) Call(name string, input any) ([]byte, error) {
	p, ok := k.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin %q not loaded", name)
	}

	payload, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshalling input: %w", err)
	}

	alloc := p.module.ExportedFunction("allocate")
	if alloc == nil {
		return nil, fmt.Errorf("plugin %q missing allocate export", name)
	}
	res, err := alloc.Call(k.ctx, uint64(len(payload)))
	if err != nil {
		return nil, fmt.Errorf("guest allocate: %w", err)
	}
	offset := uint32(res[0])

	if !p.module.Memory().Write(offset, payload) {
		return nil, fmt.Errorf("writing payload to guest memory")
	}

	execute := p.module.ExportedFunction("Execute")
	if execute == nil {
		return nil, fmt.Errorf("plugin %q missing Execute export", name)
	}
	out, err := execute.Call(k.ctx, uint64(offset), uint64(len(payload)))
	if err != nil {
		return nil, fmt.Errorf("execute: %w", err)
	}

	ptr, size := abi.DecodeFatPointer(out[0])
	return abi.ReadGuestBuffer(k.ctx, p.module, ptr, size)
}

// Close shuts down the wazero runtime.
func (k *Kernel) Close() { _ = k.runtime.Close(k.ctx) }

// ── Host capabilities ─────────────────────────────────────────────────────────

func (k *Kernel) registerCapabilities() error {
	b := kernelhost.NewModuleBuilder(k.runtime, HostModule)

	one64 := []api.ValueType{api.ValueTypeI64}
	one32 := []api.ValueType{api.ValueTypeI32}

	// fetch_url(fat_ptr url) -> fat_ptr body
	b.ExportFunction("fetch_url", one64, one64,
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			off, ln := abi.DecodeFatPointer(stack[0])
			urlBytes, err := abi.ReadGuestBuffer(ctx, mod, off, ln)
			if err != nil {
				stack[0] = hostWriteToGuest(ctx, mod, errJSON(err))
				return
			}
			body, err := httpGet(string(urlBytes))
			if err != nil {
				k.logger.Error("fetch-url", err)
				stack[0] = hostWriteToGuest(ctx, mod, errJSON(err))
				return
			}
			stack[0] = hostWriteToGuest(ctx, mod, body)
		}),
	)

	// http_download(fat_ptr "{url,dest}") -> i32 1=ok 0=err
	b.ExportFunction("http_download", one64, one32,
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			off, ln := abi.DecodeFatPointer(stack[0])
			raw, _ := abi.ReadGuestBuffer(ctx, mod, off, ln)
			var req struct {
				URL  string `json:"url"`
				Dest string `json:"dest"`
			}
			if err := json.Unmarshal(raw, &req); err != nil {
				stack[0] = 0
				return
			}
			if err := downloadFile(req.URL, req.Dest); err != nil {
				k.logger.Error("http-download", err)
				stack[0] = 0
				return
			}
			stack[0] = 1
		}),
	)

	// http_post(fat_ptr "{url,body,content_type?}") -> fat_ptr response-bytes
	// Generic HTTP POST. Plugins use this to call Ollama, Deepgram, or any
	// other API — the host doesn't need to know what the API does.
	b.ExportFunction("http_post", one64, one64,
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			off, ln := abi.DecodeFatPointer(stack[0])
			raw, _ := abi.ReadGuestBuffer(ctx, mod, off, ln)
			var req struct {
				URL         string `json:"url"`
				Body        string `json:"body"`
				ContentType string `json:"content_type"`
			}
			if err := json.Unmarshal(raw, &req); err != nil {
				stack[0] = hostWriteToGuest(ctx, mod, errJSON(err))
				return
			}
			ct := req.ContentType
			if ct == "" {
				ct = "application/json"
			}
			resp, err := http.Post(req.URL, ct, strings.NewReader(req.Body)) //nolint:gosec
			if err != nil {
				k.logger.Error("http-post", err)
				stack[0] = hostWriteToGuest(ctx, mod, errJSON(err))
				return
			}
			defer func() { _ = resp.Body.Close() }()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				stack[0] = hostWriteToGuest(ctx, mod, errJSON(err))
				return
			}
			stack[0] = hostWriteToGuest(ctx, mod, body)
		}),
	)

	// file_write(fat_ptr "{path,data}") -> i32 1=ok 0=err
	b.ExportFunction("file_write", one64, one32,
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			off, ln := abi.DecodeFatPointer(stack[0])
			raw, _ := abi.ReadGuestBuffer(ctx, mod, off, ln)
			var req struct {
				Path string `json:"path"`
				Data string `json:"data"`
			}
			if err := json.Unmarshal(raw, &req); err != nil {
				stack[0] = 0
				return
			}
			if err := os.MkdirAll(filepath.Dir(req.Path), 0o755); err != nil {
				stack[0] = 0
				return
			}
			if err := os.WriteFile(req.Path, []byte(req.Data), 0o644); err != nil {
				k.logger.Error("file-write", err)
				stack[0] = 0
				return
			}
			stack[0] = 1
		}),
	)

	// log(fat_ptr message) — fire and forget
	b.ExportFunction("log", one64, nil,
		api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			off, ln := abi.DecodeFatPointer(stack[0])
			msg, _ := abi.ReadGuestBuffer(ctx, mod, off, ln)
			k.logger.Info("plugin", lager.Data{"msg": string(msg)})
		}),
	)

	return b.Instantiate(k.ctx)
}

// ── Internal helpers ──────────────────────────────────────────────────────────

// hostWriteToGuest allocates memory in the guest (via its "allocate" export)
// and writes data into it, returning a fat pointer. Called from host capability
// implementations to return data back to the plugin.
func hostWriteToGuest(ctx context.Context, mod api.Module, data []byte) uint64 {
	alloc := mod.ExportedFunction("allocate")
	if alloc == nil {
		return 0
	}
	res, err := alloc.Call(ctx, uint64(len(data)))
	if err != nil || len(res) == 0 {
		return 0
	}
	offset := uint32(res[0])
	mod.Memory().Write(offset, data)
	return abi.EncodeFatPointer(offset, uint32(len(data)))
}

func errJSON(err error) []byte {
	b, _ := json.Marshal(map[string]string{"error": err.Error()})
	return b
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(resp.Body)
}

func downloadFile(url, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = io.Copy(f, resp.Body)
	return err
}

