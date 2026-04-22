package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildPrompt_ContainsTitleAndContent(t *testing.T) {
	prompt := buildPrompt("My Episode", "Some transcript content", nil)
	if !strings.Contains(prompt, "My Episode") {
		t.Error("prompt must contain episode title")
	}
	if !strings.Contains(prompt, "Some transcript content") {
		t.Error("prompt must contain content")
	}
}

func TestBuildPrompt_ContainsCustomScheme(t *testing.T) {
	scheme := json.RawMessage(`{"ideology":""}`)
	prompt := buildPrompt("Title", "Content", scheme)
	if !strings.Contains(prompt, `"ideology"`) {
		t.Error("prompt must include custom scheme")
	}
}

func TestParseCompletion_ValidJSON(t *testing.T) {
	raw := `{"guests":[{"name":"Alice","background":"Engineer","ideology":"pragmatic"}],"companies":[{"name":"Acme","business_model":"SaaS","customers":"SMBs"}]}`
	result, err := parseCompletion(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var extracted map[string]any
	if err := json.Unmarshal(result, &extracted); err != nil {
		t.Fatal(err)
	}
	if _, ok := extracted["guests"]; !ok {
		t.Error("missing guests key")
	}
}

func TestParseCompletion_JSONEmbeddedInPreamble(t *testing.T) {
	raw := `Here is the extracted data: {"guests":[],"companies":[{"name":"Beta","business_model":"B2B","customers":"enterprises"}]}`
	result, err := parseCompletion(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var extracted map[string]any
	if err := json.Unmarshal(result, &extracted); err != nil {
		t.Fatal(err)
	}
	if _, ok := extracted["companies"]; !ok {
		t.Error("missing companies key")
	}
}

func TestParseCompletion_NoJSONReturnsError(t *testing.T) {
	_, err := parseCompletion("no json here at all")
	if err == nil {
		t.Error("want error for response with no JSON, got nil")
	}
}

func TestParseCompletion_EmptyStringReturnsError(t *testing.T) {
	_, err := parseCompletion("")
	if err == nil {
		t.Error("want error for empty completion, got nil")
	}
}

func TestParseCompletion_MalformedJSONReturnsError(t *testing.T) {
	_, err := parseCompletion(`{"guests": [broken`)
	if err == nil {
		t.Error("want error for malformed JSON, got nil")
	}
}
