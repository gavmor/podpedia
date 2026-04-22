package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gavmor/podpedia/internal/types"
	"github.com/gavmor/wasm-microkernel/guest"
	"github.com/rozoomcool/go-ollama-sdk"
	"github.com/samber/lo"
)

// extismRoundTripper maps standard http.Client calls to guest.HttpPost
type extismRoundTripper struct{}

func (e *extismRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodPost {
		return nil, fmt.Errorf("extismRoundTripper: only POST is supported")
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	rawRes, err := guest.HttpPost(req.URL.String(), string(body))
	if err != nil {
		return nil, err
	}

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(rawRes)),
		Header:     make(http.Header),
	}
	return resp, nil
}

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

		client := ollama.NewClient(req.OllamaURL)
		client.SetHTTPClient(&http.Client{
			Transport: &extismRoundTripper{},
		})

		messages := []ollama.ChatMessage{
			{
				Role:    "system",
				Content: "You are a data extraction assistant. You MUST return ONLY valid JSON. If no information is found for a field, return an empty array or null as appropriate for the schema. Do not hallucinate.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Extract data from this podcast episode.\n\nSchema:\n%s\n\nTitle: %s\nContent: %s", schemaStr, ep.Title, content),
			},
		}

		completion, err := client.Chat(model, messages)
		if err != nil {
			return "", fmt.Errorf("ollama sdk chat: %w", err)
		}

		entry, err := parseCompletion(completion)

		// Create the final response object.
		response := make(map[string]any)

		if err != nil {
			guest.LogMsg("parse failed: " + err.Error())
			response["error"] = err.Error()
			response["raw_completion"] = completion
		} else {
			var extracted map[string]any
			if err := json.Unmarshal(entry, &extracted); err == nil && extracted != nil {
				for k, v := range extracted {
					response[k] = v
				}
			} else {
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
