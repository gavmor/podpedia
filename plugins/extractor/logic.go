package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func buildPrompt(title, content string) string {
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
