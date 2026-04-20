package pipeline

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavmor/podpedia/internal/types"
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

func TestParseRSS(t *testing.T) {
	xmlInput := `
		<rss version="2.0">
			<channel>
				<title>Test Podcast</title>
				<description>A podcast about testing</description>
				<item>
					<title>Episode 1</title>
					<description>The first episode</description>
					<pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
					<enclosure url="http://example.com/e1.mp3" type="audio/mpeg"/>
					<guid>e1-guid</guid>
				</item>
			</channel>
		</rss>
	`

	podcast, episodes, err := parseRSS([]byte(xmlInput))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if podcast.Title != "Test Podcast" {
		t.Errorf("Expected podcast title 'Test Podcast', got '%s'", podcast.Title)
	}

	if len(episodes) != 1 {
		t.Fatalf("Expected 1 episode, got %d", len(episodes))
	}

	if episodes[0].Title != "Episode 1" {
		t.Errorf("Expected episode title 'Episode 1', got '%s'", episodes[0].Title)
	}
}

func TestValidateEpisode(t *testing.T) {
	tests := []struct {
		name    string
		ep      types.Episode
		wantErr bool
	}{
		{
			name: "valid episode",
			ep: types.Episode{
				Title:    "Valid Title",
				AudioURL: "http://example.com/audio.mp3",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			ep: types.Episode{
				AudioURL: "http://example.com/audio.mp3",
			},
			wantErr: true,
		},
		{
			name: "missing audio url",
			ep: types.Episode{
				Title: "Valid Title",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateEpisode(tt.ep); (err != nil) != tt.wantErr {
				t.Errorf("validateEpisode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFetchRSSFeedIntegration(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
			<rss version="2.0">
				<channel>
					<title>Integrated Podcast</title>
					<item>
						<title>Integrated Episode</title>
						<enclosure url="http://example.com/audio.mp3" type="audio/mpeg"/>
					</item>
				</channel>
			</rss>
		`)
	}))
	defer ts.Close()

	// Update fetchRSSFeed signature or behavior to return error if we decide to
	// For now, let's test if it returns the expected episodes
	episodes, err := fetchRSSFeed(ts.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(episodes) != 1 {
		t.Fatalf("Expected 1 episode, got %d", len(episodes))
	}

	if episodes[0].Title != "Integrated Episode" {
		t.Errorf("Expected 'Integrated Episode', got '%s'", episodes[0].Title)
	}
}
