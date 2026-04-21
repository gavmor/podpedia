// Package kernel is the Podpedia host kernel: it boots wazero, registers the
// WIT-defined capabilities under the canonical module name, and calls the
// `execute` export of each loaded WASM plugin using the Component Model ABI.
package kernel

import (
        "context"
        "encoding/json"
        "fmt"
        "io"
        "os"
        "sync"

        "code.cloudfoundry.org/lager/v3"
        "github.com/tetratelabs/wazero"
        "github.com/tetratelabs/wazero/api"
        "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// capabilitiesModule is the WIT module name every plugin imports capabilities from.
const capabilitiesModule = "podpedia:kernel/host-capabilities@0.3.0"

// Kernel manages a wazero runtime, the registered capability host module, and
// the set of loaded plugin module instances.
type Kernel struct {
        ctx           context.Context
        runtime       wazero.Runtime
        plugins       map[string]api.Module
        mu            sync.Mutex
        ollamaURL     string
        transcribeURL string
}

// New creates a wazero runtime, mounts WASI preview-1, and registers all
// host capabilities under the canonical WIT module name.
func New(ctx context.Context, logger lager.Logger, ollamaURL, transcribeURL string) (*Kernel, error) {
        k := &Kernel{
                ctx:           ctx,
                runtime:       wazero.NewRuntime(ctx),
                plugins:       make(map[string]api.Module),
                ollamaURL:     ollamaURL,
                transcribeURL: transcribeURL,
        }
        wasi_snapshot_preview1.MustInstantiate(ctx, k.runtime)
        if err := k.registerCapabilities(lagerWriter{logger}); err != nil {
                _ = k.runtime.Close(ctx)
                return nil, fmt.Errorf("register capabilities: %w", err)
        }
        return k, nil
}

// Close shuts down the wazero runtime and all loaded modules.
func (k *Kernel) Close() error { return k.runtime.Close(k.ctx) }

// Load compiles and instantiates a WASM plugin from disk. The module name
// is used as the plugin identifier for subsequent Call invocations.
func (k *Kernel) Load(name, path string) error {
        wasmBytes, err := os.ReadFile(path)
        if err != nil {
                return fmt.Errorf("read plugin %s: %w", name, err)
        }
        compiled, err := k.runtime.CompileModule(k.ctx, wasmBytes)
        if err != nil {
                return fmt.Errorf("compile plugin %s: %w", name, err)
        }
        cfg := wazero.NewModuleConfig().
                WithName(name).
                WithStartFunctions("_initialize")
        mod, err := k.runtime.InstantiateModule(k.ctx, compiled, cfg)
        if err != nil {
                return fmt.Errorf("instantiate plugin %s: %w", name, err)
        }
        // Fail fast: every plugin must implement the canonical Component-Model
        // guest ABI — an `execute` entry point and a `cabi_realloc` allocator.
        // A plugin missing these exports (e.g. the current no-op stubs awaiting
        // wasm-microkernel v0.6.0) would otherwise fail silently at request time.
        for _, fn := range []string{"execute", "cabi_realloc"} {
                if mod.ExportedFunction(fn) == nil {
                        _ = mod.Close(k.ctx)
                        return fmt.Errorf("plugin %s: missing required export %q (guest ABI not wired up)", name, fn)
                }
        }
        k.mu.Lock()
        k.plugins[name] = mod
        k.mu.Unlock()
        return nil
}

// Call JSON-encodes req, passes it to the named plugin's `execute` export,
// and returns the OK payload bytes, or an error when the plugin returns an
// error result or the call itself fails.
func (k *Kernel) Call(pluginName string, req interface{}) ([]byte, error) {
        k.mu.Lock()
        mod, ok := k.plugins[pluginName]
        k.mu.Unlock()
        if !ok {
                return nil, fmt.Errorf("plugin not loaded: %s", pluginName)
        }

        reqJSON, err := json.Marshal(req)
        if err != nil {
                return nil, fmt.Errorf("marshal request: %w", err)
        }

        reqLen := uint32(len(reqJSON))
        reqPtr, err := k.cabiRealloc(mod, 0, 0, 1, reqLen)
        if err != nil {
                return nil, fmt.Errorf("alloc guest memory for request: %w", err)
        }
        if !mod.Memory().Write(reqPtr, reqJSON) {
                return nil, fmt.Errorf("write request to guest memory")
        }

        executeFn := mod.ExportedFunction("execute")
        if executeFn == nil {
                return nil, fmt.Errorf("plugin %s: missing 'execute' export", pluginName)
        }
        results, err := executeFn.Call(k.ctx, uint64(reqPtr), uint64(reqLen))
        if err != nil {
                return nil, fmt.Errorf("call execute: %w", err)
        }
        if len(results) == 0 {
                return nil, fmt.Errorf("execute: no return value")
        }

        return k.readStringResult(mod, uint32(results[0]))
}

// readStringResult reads a cm.Result[string, string, string] from guest
// memory at resultPtr.
//
// Memory layout (wasm32 / Component Model ABI):
//
//      offset 0   – isErr uint8   (0 = OK, 1 = Err)
//      offset 1-3 – 3 bytes padding (alignment of string = 4)
//      offset 4   – data.ptr uint32 (pointer to string bytes in guest memory)
//      offset 8   – data.len uint32 (byte length of the string)
//      total: 12 bytes
func (k *Kernel) readStringResult(mod api.Module, resultPtr uint32) ([]byte, error) {
        tag, ok := mod.Memory().ReadByte(resultPtr)
        if !ok {
                return nil, fmt.Errorf("read result tag at %d", resultPtr)
        }
        dataPtr, ok := mod.Memory().ReadUint32Le(resultPtr + 4)
        if !ok {
                return nil, fmt.Errorf("read result data ptr")
        }
        dataLen, ok := mod.Memory().ReadUint32Le(resultPtr + 8)
        if !ok {
                return nil, fmt.Errorf("read result data len")
        }
        if dataLen == 0 {
                if tag == 1 {
                        return nil, fmt.Errorf("plugin error: (empty)")
                }
                return []byte{}, nil
        }
        data, ok := mod.Memory().Read(dataPtr, dataLen)
        if !ok {
                return nil, fmt.Errorf("read result string from guest memory ptr=%d len=%d", dataPtr, dataLen)
        }
        if tag == 1 {
                return nil, fmt.Errorf("plugin error: %s", string(data))
        }
        return data, nil
}

// cabiRealloc calls the guest's exported `cabi_realloc` to allocate memory
// in the guest's linear memory space and returns the pointer.
func (k *Kernel) cabiRealloc(mod api.Module, ptr, origSize, align, newSize uint32) (uint32, error) {
        fn := mod.ExportedFunction("cabi_realloc")
        if fn == nil {
                return 0, fmt.Errorf("plugin does not export 'cabi_realloc'")
        }
        res, err := fn.Call(k.ctx, uint64(ptr), uint64(origSize), uint64(align), uint64(newSize))
        if err != nil {
                return 0, fmt.Errorf("cabi_realloc: %w", err)
        }
        return uint32(res[0]), nil
}

// lagerWriter bridges io.Writer to lager.Logger for plugin log messages.
type lagerWriter struct{ l lager.Logger }

func (w lagerWriter) Write(p []byte) (int, error) {
        w.l.Info("plugin", lager.Data{"msg": string(p)})
        return len(p), nil
}

var _ io.Writer = lagerWriter{}
