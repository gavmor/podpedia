//go:build wasip1

// Plugin: extractor
// Extracts structured entities from podcast episodes by calling Ollama
// via the host's generic http-post capability. Swap this plugin to change
// models, prompts, or output schemas without touching the kernel.
package main

import (
	"encoding/json"

	hostcapabilities "github.com/gavmor/podpedia/gen/podpedia/kernel/host-capabilities"
	pluginworld "github.com/gavmor/podpedia/gen/podpedia/kernel/plugin-world"
	"github.com/samber/lo"
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
			Episode struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Transcript  string `json:"transcript"`
			} `json:"episode"`
			OllamaURL string `json:"ollama_url"`
		}
		if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
			return fail("bad request: " + err.Error())
		}

		ep := req.Episode
		hostcapabilities.LogMsg("extracting: " + ep.Title)

		content, _ := lo.Coalesce(ep.Transcript, ep.Description)
		reqBody, _ := json.Marshal(map[string]any{
			"model":  "qwen3.5:27b",
			"prompt": buildPrompt(ep.Title, content),
			"stream": false,
		})

		r := hostcapabilities.HTTPPost(req.OllamaURL+"/api/generate", string(reqBody))
		if r.IsErr() {
			return fail("host http-post failed: " + *r.Err())
		}
		rawRes := *r.OK()

		var ollamaResp struct {
			Response string `json:"response"`
		}
		completion := ""
		if err := json.Unmarshal([]byte(rawRes), &ollamaResp); err == nil {
			completion = ollamaResp.Response
		}

		entry, err := parseCompletion(ep.ID, completion)
		if err != nil {
			hostcapabilities.LogMsg("parse failed, returning empty entry: " + err.Error())
			entry = map[string]any{"episode_id": ep.ID, "guests": []any{}, "companies": []any{}}
		}

		out, _ := json.Marshal(map[string]any{"entry": entry})
		return ok(string(out))
	}
}
