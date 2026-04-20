//go:build wasip1

// Plugin: extractor
// Receives an Episode and extracts structured entities (guests, companies)
// by calling the host's llm_infer capability with a structured prompt.
// Swap this plugin to change models, prompt strategies, or output schemas.
package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"unsafe"

	"github.com/gavmor/wasm-microkernel/abi"
)

func main() {}

//go:wasmimport podpedia_host llm_infer
func hostLLMInfer(fatPtr uint64) uint64

//go:wasmimport podpedia_host log
func hostLog(fatPtr uint64)

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
		}
		if err := json.Unmarshal(input, &req); err != nil {
			return errBytes("bad request: " + err.Error())
		}

		ep := req.Episode
		logMsg("extracting: " + ep.Title)
		prompt := buildPrompt(ep.Title, ep.Description, ep.Transcript)
		completion := callLLM(prompt)

		entry, err := parseCompletion(ep.ID, completion)
		if err != nil {
			logMsg("parse failed, returning empty entry: " + err.Error())
			entry = map[string]any{
				"episode_id": ep.ID,
				"guests":     []any{},
				"companies":  []any{},
			}
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

func callLLM(prompt string) string {
	promptJSON, _ := json.Marshal(prompt)
	result := hostLLMInfer(abi.ReturnBytes(promptJSON))
	off, ln := abi.DecodeFatPointer(result)
	raw := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(off))), ln)
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return string(raw)
	}
	return s
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

func logMsg(s string) { hostLog(abi.ReturnBytes([]byte(s))) }

func errBytes(msg string) []byte {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return b
}
