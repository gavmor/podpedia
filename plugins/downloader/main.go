//go:build wasip1

// Plugin: downloader
// Validates the request and delegates audio file download to the host's
// http-download capability. Owns validation; the host owns the HTTP socket.
package main

import (
	"encoding/json"
	"fmt"

	hostcapabilities "github.com/gavmor/podpedia/gen/podpedia/kernel/host-capabilities"
	pluginworld "github.com/gavmor/podpedia/gen/podpedia/kernel/plugin-world"
	"go.bytecodealliance.org/cm"
)

func main() {}

// Result is a convenience alias used throughout this file.
type Result = cm.Result[string, string, string]

func ok(s string) Result  { return cm.OK[Result](s) }
func fail(s string) Result { return cm.Err[Result](s) }

func init() {
	pluginworld.Exports.Execute = func(reqJSON string) Result {
		var req struct {
			URL  string `json:"url"`
			Dest string `json:"dest"`
		}
		if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
			return fail("bad request: " + err.Error())
		}
		if err := validateDownloadRequest(req.URL, req.Dest); err != nil {
			return fail(err.Error())
		}

		hostcapabilities.LogMsg("downloading " + req.URL)

		r := hostcapabilities.HTTPDownload(req.URL, req.Dest)
		if r.IsErr() {
			return fail("download failed: " + *r.Err())
		}

		return ok(fmt.Sprintf(`{"path":%q}`, req.Dest))
	}
}
