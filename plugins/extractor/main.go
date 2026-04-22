//go:build wasip1
package main

import (
	"encoding/json"

	"github.com/gavmor/wasm-microkernel/guest"
	"github.com/samber/lo"
)

func init() {
	guest.Register(func(reqJSON string) (string, error) {
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
			return "", err
		}

		ep := req.Episode
		guest.LogMsg("extracting: " + ep.Title)

		content, _ := lo.Coalesce(ep.Transcript, ep.Description)
		reqBody, _ := json.Marshal(map[string]any{
			"model":  "qwen3.5:27b",
			"prompt": buildPrompt(ep.Title, content),
			"stream": false,
		})

		rawRes, err := guest.HttpPost(req.OllamaURL+"/api/generate", string(reqBody))
		if err != nil {
			return "", err
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
			guest.LogMsg("parse failed, returning empty entry: " + err.Error())
			entry = map[string]any{"episode_id": ep.ID, "guests": []any{}, "companies": []any{}}
		}

		out, _ := json.Marshal(map[string]any{"entry": entry})
		return string(out), nil
	})
}

func main() {}
