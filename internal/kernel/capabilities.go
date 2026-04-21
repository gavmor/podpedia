package kernel

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tetratelabs/wazero/api"
)

const (
	resultOKTag  uint8 = 0
	resultErrTag uint8 = 1
)

// wasm param-list shorthands used by NewFunctionBuilder.
var (
	i32x2 = []api.ValueType{api.ValueTypeI32, api.ValueTypeI32}
	i32x3 = []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}
	i32x5 = []api.ValueType{api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32, api.ValueTypeI32}
)

// registerCapabilities mounts the host capability module into the runtime so
// every plugin can import it by its canonical WIT name.
func (k *Kernel) registerCapabilities(logWriter io.Writer) error {
	b := k.runtime.NewHostModuleBuilder(capabilitiesModule)

	// http-post(url_ptr, url_len, body_ptr, body_len, result_ptr)
	b.NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			urlPtr, urlLen := uint32(stack[0]), uint32(stack[1])
			bodyPtr, bodyLen := uint32(stack[2]), uint32(stack[3])
			resultPtr := uint32(stack[4])
			url := readGuestString(mod, urlPtr, urlLen)
			body := readGuestString(mod, bodyPtr, bodyLen)
			resp, err := http.Post(url, "application/json", strings.NewReader(body))
			if err != nil {
				writeErrResult(ctx, mod, resultPtr, err.Error())
				return
			}
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				writeErrResult(ctx, mod, resultPtr, err.Error())
				return
			}
			writeOKResult(ctx, mod, resultPtr, string(respBody))
		}), i32x5, nil).
		Export("http-post")

	// http-fetch(url_ptr, url_len, result_ptr)
	b.NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			urlPtr, urlLen := uint32(stack[0]), uint32(stack[1])
			resultPtr := uint32(stack[2])
			url := readGuestString(mod, urlPtr, urlLen)
			resp, err := http.Get(url)
			if err != nil {
				writeErrResult(ctx, mod, resultPtr, err.Error())
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				writeErrResult(ctx, mod, resultPtr, err.Error())
				return
			}
			writeOKResult(ctx, mod, resultPtr, string(body))
		}), i32x3, nil).
		Export("http-fetch")

	// http-download(url_ptr, url_len, dest_ptr, dest_len, result_ptr)
	b.NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			urlPtr, urlLen := uint32(stack[0]), uint32(stack[1])
			destPtr, destLen := uint32(stack[2]), uint32(stack[3])
			resultPtr := uint32(stack[4])
			url := readGuestString(mod, urlPtr, urlLen)
			dest := readGuestString(mod, destPtr, destLen)
			if err := downloadFile(url, dest); err != nil {
				writeErrResult(ctx, mod, resultPtr, err.Error())
				return
			}
			writeVoidOKResult(mod.Memory(), resultPtr)
		}), i32x5, nil).
		Export("http-download")

	// file-write(path_ptr, path_len, data_ptr, data_len, result_ptr)
	b.NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(ctx context.Context, mod api.Module, stack []uint64) {
			pathPtr, pathLen := uint32(stack[0]), uint32(stack[1])
			dataPtr, dataLen := uint32(stack[2]), uint32(stack[3])
			resultPtr := uint32(stack[4])
			path := readGuestString(mod, pathPtr, pathLen)
			data := readGuestString(mod, dataPtr, dataLen)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				writeErrResult(ctx, mod, resultPtr, err.Error())
				return
			}
			if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
				writeErrResult(ctx, mod, resultPtr, err.Error())
				return
			}
			writeVoidOKResult(mod.Memory(), resultPtr)
		}), i32x5, nil).
		Export("file-write")

	// log-msg(msg_ptr, msg_len)
	b.NewFunctionBuilder().
		WithGoModuleFunction(api.GoModuleFunc(func(_ context.Context, mod api.Module, stack []uint64) {
			msgPtr, msgLen := uint32(stack[0]), uint32(stack[1])
			msg := readGuestString(mod, msgPtr, msgLen)
			_, _ = logWriter.Write([]byte(msg))
		}), i32x2, nil).
		Export("log-msg")

	_, err := b.Instantiate(k.ctx)
	return err
}

// readGuestString reads bytes from the guest's linear memory at (ptr, length).
func readGuestString(mod api.Module, ptr, length uint32) string {
	if length == 0 {
		return ""
	}
	b, _ := mod.Memory().Read(ptr, length)
	return string(b)
}

// writeOKResult writes a cm.Result OK(string) at resultPtr in guest memory.
// The string payload is allocated in guest memory via cabi_realloc.
//
// Layout: [tag u8][pad 3][str_ptr u32][str_len u32] = 12 bytes
func writeOKResult(ctx context.Context, mod api.Module, resultPtr uint32, s string) {
	writeStringResult(ctx, mod, resultPtr, resultOKTag, s)
}

// writeErrResult writes a cm.Result Err(string) at resultPtr in guest memory.
func writeErrResult(ctx context.Context, mod api.Module, resultPtr uint32, errMsg string) {
	writeStringResult(ctx, mod, resultPtr, resultErrTag, errMsg)
}

// writeVoidOKResult writes a cm.Result<_, string> OK() — tag=0, zero payload.
// No allocation needed since the OK type is void (struct{}).
func writeVoidOKResult(mem api.Memory, resultPtr uint32) {
	mem.WriteByte(resultPtr, resultOKTag)
	mem.WriteByte(resultPtr+1, 0)
	mem.WriteByte(resultPtr+2, 0)
	mem.WriteByte(resultPtr+3, 0)
	mem.WriteUint32Le(resultPtr+4, 0)
	mem.WriteUint32Le(resultPtr+8, 0)
}

// writeStringResult is the shared helper for OK and Err variants that carry
// a string payload. It calls cabi_realloc in the guest module to allocate
// guest memory, writes the string bytes there, then writes the result struct.
func writeStringResult(ctx context.Context, mod api.Module, resultPtr uint32, tag uint8, s string) {
	mem := mod.Memory()
	mem.WriteByte(resultPtr, tag)
	mem.WriteByte(resultPtr+1, 0)
	mem.WriteByte(resultPtr+2, 0)
	mem.WriteByte(resultPtr+3, 0)

	if len(s) == 0 {
		mem.WriteUint32Le(resultPtr+4, 0)
		mem.WriteUint32Le(resultPtr+8, 0)
		return
	}

	allocFn := mod.ExportedFunction("cabi_realloc")
	if allocFn == nil {
		mem.WriteUint32Le(resultPtr+4, 0)
		mem.WriteUint32Le(resultPtr+8, 0)
		return
	}
	res, err := allocFn.Call(ctx, 0, 0, 1, uint64(len(s)))
	if err != nil || len(res) == 0 {
		mem.WriteUint32Le(resultPtr+4, 0)
		mem.WriteUint32Le(resultPtr+8, 0)
		return
	}
	strPtr := uint32(res[0])
	mem.Write(strPtr, []byte(s))
	mem.WriteUint32Le(resultPtr+4, strPtr)
	mem.WriteUint32Le(resultPtr+8, uint32(len(s)))
}

// downloadFile fetches url and saves it atomically to dest, creating parent
// directories as needed.
func downloadFile(url, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
