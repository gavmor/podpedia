package main

import (
	"encoding/json"

	"github.com/gavmor/podpedia/internal/types"
	"github.com/gavmor/wasm-microkernel/guest"
	"github.com/samber/lo"
)

func init() {
	guest.Register(func(reqJSON string) (string, error) {
		var req struct {
			Episode     types.Episode   `json:"episode"`
			OllamaURL   string          `json:"ollama_url"`
			OllamaModel string          `json:"ollama_model"`
			Scheme      json.RawMessage `json:"scheme"`
		}
		if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
			return "", err
		}

		ep := req.Episode
		guest.LogMsg("extracting: " + ep.Title)

		content, _ := lo.Coalesce(ep.Transcript, ep.Description)

		model := req.OllamaModel
		if model == "" {
			model = "qwen2.5:0.5b"
		}

		reqBody, _ := json.Marshal(map[string]any{
			"model":  model,
			"prompt": buildPrompt(ep.Title, content, req.Scheme),
			"format": "json",
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

		entry, err := parseCompletion(completion)
		
		// Create the final response object.
		response := make(map[string]any)
		
		if err != nil {
			guest.LogMsg("parse failed: " + err.Error())
			response["error"] = err.Error()
			response["raw_completion"] = completion
		} else {
			// If it's an object, merge its keys.
			var extracted map[string]any
			if err := json.Unmarshal(entry, &extracted); err == nil && extracted != nil {
				for k, v := range extracted {
					response[k] = v
				}
			} else {
				// If it's not an object (e.g. array), put it in a 'data' field.
				var data any
				_ = json.Unmarshal(entry, &data)
				response["data"] = data
			}
		}

		out, _ := json.Marshal(response)
		return string(out), nil
	})
}

func main() {}
