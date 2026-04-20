//go:build wasip1

// Plugin: extractor
// Extracts structured entities from podcast episodes by calling Ollama
// via the host's generic http-post capability. Swap this plugin to change
// models, prompts, or output schemas without touching the kernel.
package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gavmor/wasm-microkernel/abi"
	"github.com/gavmor/wasm-microkernel/guest"
)

func main() {}

//go:wasmexport allocate
func allocate(size uint32) uint32 { return abi.GuestAllocate(size) }

//go:wasmexport Execute
func Execute(offset, length uint32) uint64 {
	return abi.Delegate(offset, length, func(input []byte) []byte {
		var req struct {
			Episode struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Transcript  string `json:"transcript"`
			} `json:"episode"`
			OllamaURL string `json:"ollama_url"`
		}
		if err := json.Unmarshal(input, &req); err != nil {
			return errBytes("bad request: " + err.Error())
		}

		ep := req.Episode
		guest.Log("extracting: " + ep.Title)

		reqBody, _ := json.Marshal(map[string]any{
			"model":  "qwen3.5:27b",
			"prompt": buildPrompt(ep.Title, ep.Description, ep.Transcript),
			"stream": false,
		})
		raw := guest.HTTPPost(req.OllamaURL+"/api/generate", string(reqBody))

		var ollamaResp struct {
			Response string `json:"response"`
		}
		completion := ""
		if err := json.Unmarshal(raw, &ollamaResp); err == nil {
			completion = ollamaResp.Response
		}

		entry, err := parseCompletion(ep.ID, completion)
		if err != nil {
			guest.Log("parse failed, returning empty entry: " + err.Error())
			entry = map[string]any{"episode_id": ep.ID, "guests": []any{}, "companies": []any{}}
		}

		out, _ := json.Marshal(map[string]any{"entry": entry})
		return out
	})
}

func buildPrompt(title, description, transcript string) string {
	content := description
	if transcript != "" {
		content = transcript
	}
	return fmt.Sprintf(`Extract structured data from this podcast episode. Return ONLY valid JSON:
{"guests":[{"name":"","background":"","ideology":""}],"companies":[{"name":"","business_model":"","customers":""}]}

Episode: %s
Content: %s
JSON:`, title, content)
}

func parseCompletion(episodeID, raw string) (map[string]any, error) {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("no JSON in completion")
	}
	var extracted struct {
		Guests []struct {
			Name       string `json:"name"`
			Background string `json:"background"`
			Ideology   string `json:"ideology"`
		} `json:"guests"`
		Companies []struct {
			Name          string `json:"name"`
			BusinessModel string `json:"business_model"`
			Customers     string `json:"customers"`
		} `json:"companies"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &extracted); err != nil {
		return nil, err
	}
	return map[string]any{
		"episode_id": episodeID,
		"guests":     extracted.Guests,
		"companies":  extracted.Companies,
	}, nil
}

func errBytes(msg string) []byte {
	return []byte(fmt.Sprintf(`{"error":%q}`, msg))
}
