//go:build wasip1
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/extism/go-pdk"
	"github.com/gavmor/wasm-microkernel/guest"
)

func init() {
	guest.Register(func(reqJSON string) (string, error) {
		var req struct {
			URL  string `json:"url"`
			Dest string `json:"dest"`
		}
		if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
			return "", err
		}
		if err := validateDownloadRequest(req.URL, req.Dest); err != nil {
			return "", err
		}

		guest.LogMsg("downloading " + req.URL)

		httpReq := pdk.NewHTTPRequest(pdk.MethodGet, req.URL)
		resp := httpReq.Send()
		if resp.Status() >= 400 {
			return "", fmt.Errorf("download failed: HTTP %d", resp.Status())
		}

		if err := os.MkdirAll(filepath.Dir(req.Dest), 0755); err != nil {
			return "", fmt.Errorf("mkdir failed: %v", err)
		}
		if err := os.WriteFile(req.Dest, resp.Body(), 0644); err != nil {
			return "", fmt.Errorf("write file failed: %v", err)
		}

		return fmt.Sprintf(`{"path":%q}`, req.Dest), nil
	})
}

func main() {}
