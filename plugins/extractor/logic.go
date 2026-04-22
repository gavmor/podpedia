package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

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
