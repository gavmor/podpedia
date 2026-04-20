//go:build wasip1

// Plugin: downloader
// Delegates audio file download to the host's http_download capability.
// Owns validation and path normalization; the host owns the HTTP socket.
package main

import (
	"encoding/json"
	"strings"

	"github.com/gavmor/wasm-microkernel/abi"
)

func main() {}

//go:wasmimport podpedia_host http_download
func hostHTTPDownload(fatPtr uint64) uint32

//go:wasmimport podpedia_host log
func hostLog(fatPtr uint64)

//go:wasmexport allocate
func allocate(size uint32) uint32 { return abi.GuestAllocate(size) }

//go:wasmexport Execute
func Execute(offset, length uint32) uint64 {
	return abi.Delegate(offset, length, func(input []byte) []byte {
		var req struct {
			URL  string `json:"url"`
			Dest string `json:"dest"`
		}
		if err := json.Unmarshal(input, &req); err != nil {
			return errBytes("bad request: " + err.Error())
		}
		if req.URL == "" {
			return errBytes("url required")
		}
		if req.Dest == "" {
			return errBytes("dest required")
		}
		if !strings.HasPrefix(req.URL, "http") {
			return errBytes("url must be http(s)")
		}

		logMsg("downloading " + req.URL)
		payload, _ := json.Marshal(map[string]string{"url": req.URL, "dest": req.Dest})
		if hostHTTPDownload(abi.ReturnBytes(payload)) == 0 {
			return errBytes("download failed: " + req.URL)
		}

		out, _ := json.Marshal(map[string]string{"path": req.Dest})
		return out
	})
}

func logMsg(s string) { hostLog(abi.ReturnBytes([]byte(s))) }
func errBytes(msg string) []byte {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return b
}
