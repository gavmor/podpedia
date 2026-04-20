//go:build wasip1

// Plugin: downloader
// Validates the request and delegates audio file download to the host's
// http-download capability. Owns validation; the host owns the HTTP socket.
package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gavmor/wasm-microkernel/abi"
	"github.com/gavmor/wasm-microkernel/guest"
)

func main() {}

//go:wasmexport allocate
func allocate(size uint32) uint32 { return abi.GuestAllocate(size) }

//go:wasmexport Execute
func Execute(offset, length uint32) uint64 {
	return abi.Delegate(offset, length, func(in []byte) []byte {
		var req struct {
			URL  string `json:"url"`
			Dest string `json:"dest"`
		}
		if err := json.Unmarshal(in, &req); err != nil {
			return errBytes("bad request: " + err.Error())
		}

		switch {
		case req.URL == "":
			return errBytes("url required")
		case req.Dest == "":
			return errBytes("dest required")
		case !strings.HasPrefix(req.URL, "http"):
			return errBytes("url must be http(s)")
		}

		guest.Log("downloading " + req.URL)

		if !guest.HTTPDownload(req.URL, req.Dest) {
			return errBytes("download failed: " + req.URL)
		}

		return []byte(fmt.Sprintf(`{"path":%q}`, req.Dest))
	})
}

func errBytes(msg string) []byte {
	return []byte(fmt.Sprintf(`{"error":%q}`, msg))
}
