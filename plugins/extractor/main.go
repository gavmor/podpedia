package main

import (
	"encoding/json"
	"fmt"

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

		// Truncate content to avoid overwhelming tiny models
		if len(content) > 4000 {
			content = content[:4000] + "..."
		}

		model := req.OllamaModel
		if model == "" {
			model = "qwen2.5:0.5b"
		}

		schemaStr := string(req.Scheme)
		if schemaStr == "" || schemaStr == "null" {
			schemaStr = `{"guests":[{"name":"","background":"","ideology":""}],"companies":[{"name":"","business_model":"","customers":""}]}`
		}

		messages := []map[string]string{
			{
				"role":    "system",
				"content": "You are a data extraction assistant. You MUST return ONLY valid JSON. If no information is found for a field, return an empty array or null as appropriate for the schema. Do not hallucinate.",
			},
			{
				"role": "user",
				"content": fmt.Sprintf("Extract data from this podcast episode.\n\nSchema:\n%s\n\nTitle: %s\nContent: %s",
					schemaStr, ep.Title, content),
			},
		}

		reqBody, _ := json.Marshal(map[string]any{
			"model":    model,
			"messages": messages,
			"format":   "json",
			"stream":   false,
			"options": map[string]any{
				"num_predict": 1000,
			},
		})

		rawRes, err := guest.HttpPost(req.OllamaURL+"/api/chat", string(reqBody))
		if err != nil {
			return "", err
		}

		var ollamaResp struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}
		if err := json.Unmarshal([]byte(rawRes), &ollamaResp); err != nil {
			return "", fmt.Errorf("ollama response: %w", err)
		}
		completion := ollamaResp.Message.Content

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
