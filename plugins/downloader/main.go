//go:build wasip1

// Plugin: downloader
// Validates the request and delegates audio file download to the host's
// http-download capability. Owns validation; the host owns the HTTP socket.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/gavmor/wasm-microkernel/guest-bindings/plugin_world"
	host "github.com/gavmor/wasm-microkernel/guest-bindings/podpedia/kernel/host_capabilities"
)

func main() {}

func init() { plugin_world.SetExportsPluginWorld(&DownloaderPlugin{}) }

type DownloaderPlugin struct{}

func (d *DownloaderPlugin) Execute(reqJSON string) (plugin_world.Result[string, string], error) {
	var req struct {
		URL  string `json:"url"`
		Dest string `json:"dest"`
	}
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		return plugin_world.Err[string, string]("bad request: " + err.Error()), nil
	}

	if err := validateDownloadRequest(req.URL, req.Dest); err != nil {
		return plugin_world.Err[string, string](err.Error()), nil
	}

	host.LogMsg("downloading " + req.URL)

	if err := host.HttpDownload(req.URL, req.Dest); err != nil {
		return plugin_world.Err[string, string]("download failed: " + err.Error()), nil
	}

	return plugin_world.Ok[string, string](fmt.Sprintf(`{"path":%q}`, req.Dest)), nil
}
