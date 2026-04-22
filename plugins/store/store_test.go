package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestSlug_AlphanumericPassThrough(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"Hello123", "Hello123"},
		{"abc-def", "abc_def"},
		{"ABC", "ABC"},
	}
	for _, tc := range cases {
		got := slug(tc.input)
		if got != tc.want {
			t.Errorf("slug(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestHandleStructured_CreatesCorrectFilename(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "store-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	episode := map[string]any{
		"id": "ep-123",
	}
	entry := map[string]any{
		"data": "some data",
	}
	
	req := map[string]any{
		"output_dir": tmpDir,
		"episode":    episode,
		"entry":      entry,
		"scheme_id":  "my-scheme",
	}
	reqJSON, _ := json.Marshal(req)

	resJSON, err := HandleStructured(reqJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var res struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(resJSON), &res); err != nil {
		t.Fatal(err)
	}

	path := res.Path
	expectedSuffix := "ep_123_my-scheme.json"
	if !strings.HasSuffix(path, expectedSuffix) {
		t.Errorf("expected path to end with %q, got %q", expectedSuffix, path)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("file was not created at %q", path)
	}
}

func TestHandleStructured_SlugifiesID(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "store-test-slug")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	episode := map[string]any{
		"id": "http://id/123",
	}
	entry := map[string]any{
		"data": "some data",
	}
	req := map[string]any{
		"output_dir": tmpDir,
		"episode":    episode,
		"entry":      entry,
		"scheme_id":  "test",
	}
	reqJSON, _ := json.Marshal(req)

	resJSON, err := HandleStructured(reqJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var res struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal([]byte(resJSON), &res)
	path := res.Path

	if !strings.Contains(path, "http___id_123_test.json") {
		t.Errorf("expected path to contain slugified ID, got %q", path)
	}
}
