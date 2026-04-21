package main

import (
	"strings"
	"testing"
)

func TestBuildPrompt_ContainsTitleAndContent(t *testing.T) {
	prompt := buildPrompt("My Episode", "Some transcript content")
	if !strings.Contains(prompt, "My Episode") {
		t.Error("prompt must contain episode title")
	}
	if !strings.Contains(prompt, "Some transcript content") {
		t.Error("prompt must contain content")
	}
}

func TestBuildPrompt_ContainsJSONSchema(t *testing.T) {
	prompt := buildPrompt("Title", "Content")
	if !strings.Contains(prompt, `"guests"`) {
		t.Error("prompt must include guests schema")
	}
	if !strings.Contains(prompt, `"companies"`) {
		t.Error("prompt must include companies schema")
	}
}

func TestParseCompletion_ValidJSON(t *testing.T) {
	raw := `{"guests":[{"name":"Alice","background":"Engineer","ideology":"pragmatic"}],"companies":[{"name":"Acme","business_model":"SaaS","customers":"SMBs"}]}`
	result, err := parseCompletion("ep-1", raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["episode_id"] != "ep-1" {
		t.Errorf("want episode_id %q, got %v", "ep-1", result["episode_id"])
	}
	guests, ok := result["guests"]
	if !ok {
		t.Fatal("missing guests key")
	}
	_ = guests
}

func TestParseCompletion_JSONEmbeddedInPreamble(t *testing.T) {
	raw := `Here is the extracted data: {"guests":[],"companies":[{"name":"Beta","business_model":"B2B","customers":"enterprises"}]}`
	result, err := parseCompletion("ep-2", raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["episode_id"] != "ep-2" {
		t.Errorf("want episode_id %q, got %v", "ep-2", result["episode_id"])
	}
}

func TestParseCompletion_NoJSONReturnsError(t *testing.T) {
	_, err := parseCompletion("ep-3", "no json here at all")
	if err == nil {
		t.Error("want error for response with no JSON, got nil")
	}
}

func TestParseCompletion_EmptyStringReturnsError(t *testing.T) {
	_, err := parseCompletion("ep-4", "")
	if err == nil {
		t.Error("want error for empty completion, got nil")
	}
}

func TestParseCompletion_MalformedJSONReturnsError(t *testing.T) {
	_, err := parseCompletion("ep-5", `{"guests": [broken`)
	if err == nil {
		t.Error("want error for malformed JSON, got nil")
	}
}

func TestParseCompletion_EmptyGuestsAndCompanies(t *testing.T) {
	raw := `{"guests":[],"companies":[]}`
	result, err := parseCompletion("ep-6", raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["episode_id"] != "ep-6" {
		t.Errorf("want episode_id %q, got %v", "ep-6", result["episode_id"])
	}
}
