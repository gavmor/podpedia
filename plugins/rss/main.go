//go:build wasip1

// Plugin: rss
// Receives raw RSS XML from the host (which fetched the feed URL) and returns
// a structured list of podcast episodes as JSON. No host imports needed —
// this plugin is pure computation.
package main

import (
	"encoding/json"

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
			XML string `json:"xml"`
		}
		if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
			return fail("bad request: " + err.Error())
		}
		p, eps, err := parseRSS(req.XML)
		if err != nil {
			return fail("rss parse: " + err.Error())
		}
		out, _ := json.Marshal(map[string]any{"podcast": p, "episodes": eps})
		return ok(string(out))
	}
}
