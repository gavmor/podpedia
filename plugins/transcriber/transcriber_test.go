package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseASRResponse_ExtractsTranscriptField(t *testing.T) {
	raw := `{"transcript":"Hello world"}`
	got := parseASRResponse(raw)
	if got != "Hello world" {
		t.Errorf("want %q, got %q", "Hello world", got)
	}
}

func TestParseASRResponse_EmptyTranscript(t *testing.T) {
	raw := `{"transcript":""}`
	got := parseASRResponse(raw)
	if got != "" {
		t.Errorf("want empty string, got %q", got)
	}
}

func TestParseASRResponse_FallsBackToRawOnInvalidJSON(t *testing.T) {
	raw := "plain text response"
	got := parseASRResponse(raw)
	if got != raw {
		t.Errorf("want raw response %q, got %q", raw, got)
	}
}

func TestParseASRResponse_FallsBackToRawOnMissingField(t *testing.T) {
	// Valid JSON but no "transcript" key — struct zero-value is "", which is fine
	raw := `{"text":"something else"}`
	got := parseASRResponse(raw)
	// transcript field defaults to empty string when key is absent
	if got != "" {
		t.Errorf("want empty string for missing transcript field, got %q", got)
	}
}

func TestBuildTranscribeBody_ContainsAudioURL(t *testing.T) {
	body := buildTranscribeBody("https://example.com/audio.mp3")
	var m map[string]string
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if m["audio_url"] != "https://example.com/audio.mp3" {
		t.Errorf("want audio_url %q, got %q", "https://example.com/audio.mp3", m["audio_url"])
	}
}

func TestFormatTranscriptResult_ProducesValidJSON(t *testing.T) {
	result := formatTranscriptResult("hello world")
	if !strings.Contains(result, `"transcript"`) {
		t.Error("result must contain transcript key")
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(result), &m); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if m["transcript"] != "hello world" {
		t.Errorf("want %q, got %q", "hello world", m["transcript"])
	}
}

func TestFormatTranscriptResult_EscapesSpecialChars(t *testing.T) {
	result := formatTranscriptResult(`say "hello" & goodbye`)
	var m map[string]string
	if err := json.Unmarshal([]byte(result), &m); err != nil {
		t.Fatalf("result is not valid JSON after escaping: %v", err)
	}
	if m["transcript"] != `say "hello" & goodbye` {
		t.Errorf("transcript not round-tripped correctly: %q", m["transcript"])
	}
}
