package pipeline

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchFeedContent(t *testing.T) {
	// Mock a successful RSS server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `<rss><channel><title>Test</title></channel></rss>`)
	}))
	defer ts.Close()

	// This is the function we want to implement
	content, err := fetchFeedContent(ts.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected content, got empty slice")
	}

	// Test error case
	_, err = fetchFeedContent("http://invalid-url-that-hopefully-does-not-exist.com")
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}
