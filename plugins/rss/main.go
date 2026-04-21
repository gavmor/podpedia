//go:build wasip1

// Plugin: rss
// Receives raw RSS XML from the host (which fetched the feed URL) and returns
// a structured list of podcast episodes as JSON. No host imports needed —
// this plugin is pure computation.
package main

import (
	"encoding/json"

	"github.com/gavmor/wasm-microkernel/guest-bindings/plugin_world"
)

func main() {}

func init() { plugin_world.SetExportsPluginWorld(&RSSPlugin{}) }

type RSSPlugin struct{}

func (r *RSSPlugin) Execute(reqJSON string) (plugin_world.Result[string, string], error) {
	var req struct {
		XML string `json:"xml"`
	}
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		return plugin_world.Err[string, string]("bad request: " + err.Error()), nil
	}
	p, eps, err := parseRSS(req.XML)
	if err != nil {
		return plugin_world.Err[string, string]("rss parse: " + err.Error()), nil
	}
	out, _ := json.Marshal(map[string]any{"podcast": p, "episodes": eps})
	return plugin_world.Ok[string, string](string(out)), nil
}
