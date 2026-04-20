package types

import (
	"encoding/xml"
	"testing"
)

func TestMetadataStructures(t *testing.T) {
	// Test data representing a typical podcast RSS item
	xmlInput := `
		<rss xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd" version="2.0">
			<channel>
				<title>Test Podcast</title>
				<description>A podcast about testing</description>
				<item>
					<title>Episode 1</title>
					<description>The first episode</description>
					<pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
					<enclosure url="http://example.com/e1.mp3" type="audio/mpeg" length="123456"/>
					<guid>e1-guid</guid>
				</item>
			</channel>
		</rss>
	`

	// This test defines what we expect our structures to handle
	type RSS struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Title       string    `xml:"title"`
			Description string    `xml:"description"`
			Episodes    []Episode `xml:"item"`
		} `xml:"channel"`
	}

	var rss RSS
	err := xml.Unmarshal([]byte(xmlInput), &rss)
	if err != nil {
		t.Fatalf("Failed to unmarshal XML: %v", err)
	}

	if rss.Channel.Title != "Test Podcast" {
		t.Errorf("Expected podcast title 'Test Podcast', got '%s'", rss.Channel.Title)
	}

	if len(rss.Channel.Episodes) != 1 {
		t.Fatalf("Expected 1 episode, got %d", len(rss.Channel.Episodes))
	}

	ep := rss.Channel.Episodes[0]
	if ep.Title != "Episode 1" {
		t.Errorf("Expected episode title 'Episode 1', got '%s'", ep.Title)
	}
	if ep.AudioURL != "http://example.com/e1.mp3" {
		t.Errorf("Expected audio URL 'http://example.com/e1.mp3', got '%s'", ep.AudioURL)
	}

	// Verify new fields are present (though not unmarshaled by legacy XML)
	if ep.Author != "" {
		t.Errorf("Expected empty author for legacy parse, got '%s'", ep.Author)
	}
}
