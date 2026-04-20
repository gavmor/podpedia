package pipeline

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavmor/podpedia/internal/types"
)

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
