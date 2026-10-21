package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gavmor/podpedia/internal/types"
)

// extractRequest is the wire contract with the host: it must stay
// structurally identical to the anonymous struct that
// internal/kernel.(*extractor).ExtractEntities marshals.
//
// types.Episode is resolved through the `replace github.com/gavmor/podpedia
// => ../..` in go.mod, so it is the same struct definition the host encodes
// with. Without that replace this module compiled against a published zip
// whose Episode carried no json tags, and the host's underscored keys
// (audio_url, pub_date) bound to nothing — Go's default field matching is
// case-insensitive but does not cross underscores.
type extractRequest struct {
	Episode types.Episode   `json:"episode"`
	Scheme  json.RawMessage `json:"scheme"`
}

func buildPrompt(title, content string, scheme json.RawMessage) string {
	schemaStr := string(scheme)
	if schemaStr == "" || schemaStr == "null" {
		schemaStr = `{"guests":[{"name":"","background":"","ideology":""}],"companies":[{"name":"","business_model":"","customers":""}]}`
	}

	return fmt.Sprintf(`Extract structured data from this podcast episode into the JSON format provided.
Return ONLY valid JSON.

Schema/Template:
%s

Episode Title: %s
Episode Content: %s
JSON:`, schemaStr, title, content)
}

func parseCompletion(raw string) (json.RawMessage, error) {
	raw = stripMarkdown(raw)
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("no JSON in completion")
	}
	
	js := raw[start : end+1]
	return json.RawMessage(js), nil
}

func stripMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) > 2 {
			return strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	return s
}
