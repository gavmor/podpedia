//go:build wasip1

// Package pluginworld provides the cabi_realloc export required by the
// Component Model canonical ABI. The host calls this function to allocate
// guest linear memory before writing strings into it.
package pluginworld

import "unsafe"

// pins keeps allocations reachable so the GC cannot collect them between
// the WASM call that returns a pointer and the host's subsequent read.
var pins [][]byte

//go:wasmexport cabi_realloc
func cabiRealloc(ptr, origSize, align, newSize uint32) uint32 {
	if newSize == 0 {
		return 0
	}
	buf := make([]byte, newSize)
	if ptr != 0 && origSize > 0 {
		old := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), origSize)
		copy(buf, old)
	}
	pins = append(pins, buf)
	return uint32(uintptr(unsafe.Pointer(&buf[0])))
}
