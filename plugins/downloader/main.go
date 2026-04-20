//go:build wasip1

// Plugin: downloader
// Delegates audio file download to the host's http_download capability.
// Owns validation and path normalization; the host owns the HTTP socket.
package main

import (
	"encoding/json"
	"fmt"
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

		logMsg("downloading " + req.URL)
		p, _ := json.Marshal(req)

		if hostHTTPDownload(abi.ReturnBytes(p)) == 0 {
			return errBytes("download failed: " + req.URL)
		}

		return []byte(fmt.Sprintf(`{"path":%q}`, req.Dest))
	})
}

func logMsg(s string) { hostLog(abi.ReturnBytes([]byte(s))) }
func errBytes(msg string) []byte {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return b
}
