package pipeline

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseWithGofeed(t *testing.T) {
	xmlInput := `
		<rss xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd" xmlns:dc="http://purl.org/dc/elements/1.1/" version="2.0">
			<channel>
				<title>Integrated Podcast</title>
				<description>A podcast for testing gofeed</description>
				<itunes:author>Test Author</itunes:author>
				<item>
					<title>Integrated Episode</title>
					<description>Testing metadata extraction</description>
					<pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
					<guid>integrated-guid</guid>
					<enclosure url="http://example.com/audio.mp3" type="audio/mpeg"/>
					<itunes:duration>00:45:00</itunes:duration>
					<itunes:explicit>yes</itunes:explicit>
					<dc:creator>Dublin Core Creator</dc:creator>
				</item>
			</channel>
		</rss>
	`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, xmlInput)
	}))
	defer ts.Close()

	// This is the function we want to implement or update
	podcast, episodes, err := parseRSSWithGofeed(ts.URL)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if podcast.Title != "Integrated Podcast" {
		t.Errorf("Expected podcast title 'Integrated Podcast', got '%s'", podcast.Title)
	}
	if podcast.Author != "Test Author" {
		t.Errorf("Expected podcast author 'Test Author', got '%s'", podcast.Author)
	}

	if len(episodes) != 1 {
		t.Fatalf("Expected 1 episode, got %d", len(episodes))
	}

	ep := episodes[0]
	if ep.Title != "Integrated Episode" {
		t.Errorf("Expected episode title 'Integrated Episode', got '%s'", ep.Title)
	}
	if ep.Duration != "00:45:00" {
		t.Errorf("Expected duration '00:45:00', got '%s'", ep.Duration)
	}
	if !ep.Explicit {
		t.Error("Expected explicit flag to be true")
	}
	if ep.Author != "Dublin Core Creator" {
		t.Errorf("Expected author 'Dublin Core Creator', got '%s'", ep.Author)
	}
}
