//go:build wasip1

// Plugin: extractor
// Extracts structured entities from podcast episodes by calling Ollama
// via the host's generic http-post capability. Swap this plugin to change
// models, prompts, or output schemas without touching the kernel.
package main

import (
	"encoding/json"

	"github.com/gavmor/wasm-microkernel/guest-bindings/plugin_world"
	host "github.com/gavmor/wasm-microkernel/guest-bindings/podpedia/kernel/host_capabilities"
	"github.com/samber/lo"
)

func main() {}

func init() { plugin_world.SetExportsPluginWorld(&ExtractorPlugin{}) }

type ExtractorPlugin struct{}

func (e *ExtractorPlugin) Execute(reqJSON string) (plugin_world.Result[string, string], error) {
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
		return plugin_world.Err[string, string]("bad request: " + err.Error()), nil
	}

	ep := req.Episode
	host.LogMsg("extracting: " + ep.Title)

	content, _ := lo.Coalesce(ep.Transcript, ep.Description)
	reqBody, _ := json.Marshal(map[string]any{
		"model":  "qwen3.5:27b",
		"prompt": buildPrompt(ep.Title, content),
		"stream": false,
	})

	rawRes, err := host.HttpPost(req.OllamaURL+"/api/generate", string(reqBody))
	if err != nil {
		return plugin_world.Err[string, string]("host http-post failed: " + err.Error()), nil
	}

	var ollamaResp struct {
		Response string `json:"response"`
	}
	completion := ""
	if err := json.Unmarshal([]byte(rawRes), &ollamaResp); err == nil {
		completion = ollamaResp.Response
	}

	entry, err := parseCompletion(ep.ID, completion)
	if err != nil {
		host.LogMsg("parse failed, returning empty entry: " + err.Error())
		entry = map[string]any{"episode_id": ep.ID, "guests": []any{}, "companies": []any{}}
	}

	out, _ := json.Marshal(map[string]any{"entry": entry})
	return plugin_world.Ok[string, string](string(out)), nil
}
