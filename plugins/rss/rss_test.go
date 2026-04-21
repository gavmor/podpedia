package main

import (
	"encoding/json"
	"strings"
	"testing"
)

const minimalFeed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Podcast</title>
    <description>A test podcast</description>
    <item>
      <title>Episode One</title>
      <guid>ep-001</guid>
      <description>  First episode.  </description>
      <enclosure url="https://example.com/ep1.mp3" type="audio/mpeg"/>
      <pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
    </item>
  </channel>
</rss>`

const multiepisodeFeed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">
  <channel>
    <title>Multi Pod</title>
    <description>Several episodes</description>
    <item>
      <title>Ep A</title>
      <guid>ep-a</guid>
      <enclosure url="https://example.com/a.mp3" type="audio/mpeg"/>
      <itunes:duration>30:00</itunes:duration>
    </item>
    <item>
      <title>Ep B</title>
      <guid>ep-b</guid>
      <enclosure url="https://example.com/b.mp3" type="audio/mpeg"/>
    </item>
    <item>
      <title></title>
      <guid>no-title</guid>
      <enclosure url="https://example.com/c.mp3" type="audio/mpeg"/>
    </item>
    <item>
      <title>No Enclosure</title>
      <guid>no-enc</guid>
    </item>
  </channel>
</rss>`

func TestParseRSS_PodcastMetadata(t *testing.T) {
	pod, _, err := parseRSS(minimalFeed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Title != "Test Podcast" {
		t.Errorf("want title %q, got %q", "Test Podcast", pod.Title)
	}
	if pod.Description != "A test podcast" {
		t.Errorf("want description %q, got %q", "A test podcast", pod.Description)
	}
}

func TestParseRSS_EpisodeFields(t *testing.T) {
	_, eps, err := parseRSS(minimalFeed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eps) != 1 {
		t.Fatalf("want 1 episode, got %d", len(eps))
	}
	ep := eps[0]
	if ep.ID != "ep-001" {
		t.Errorf("want id %q, got %q", "ep-001", ep.ID)
	}
	if ep.Title != "Episode One" {
		t.Errorf("want title %q, got %q", "Episode One", ep.Title)
	}
	if ep.Description != "First episode." {
		t.Errorf("want description trimmed, got %q", ep.Description)
	}
	if ep.AudioURL != "https://example.com/ep1.mp3" {
		t.Errorf("want audio url, got %q", ep.AudioURL)
	}
	if ep.PubDate == "" {
		t.Error("want non-empty pub_date")
	}
}

func TestParseRSS_FiltersIncompleteItems(t *testing.T) {
	_, eps, err := parseRSS(multiepisodeFeed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Items with no title or no enclosure must be skipped
	if len(eps) != 2 {
		t.Fatalf("want 2 valid episodes, got %d", len(eps))
	}
	ids := map[string]bool{}
	for _, ep := range eps {
		ids[ep.ID] = true
	}
	if !ids["ep-a"] || !ids["ep-b"] {
		t.Errorf("unexpected episode ids: %v", ids)
	}
}

func TestParseRSS_ItunesDuration(t *testing.T) {
	_, eps, err := parseRSS(multiepisodeFeed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var epA outEpisode
	for _, ep := range eps {
		if ep.ID == "ep-a" {
			epA = ep
		}
	}
	if epA.Duration != "30:00" {
		t.Errorf("want duration %q, got %q", "30:00", epA.Duration)
	}
}

func TestParseRSS_FallbackIDToEnclosureURL(t *testing.T) {
	feed := `<?xml version="1.0"?>
<rss version="2.0"><channel><title>X</title>
  <item>
    <title>No GUID Episode</title>
    <enclosure url="https://example.com/noguid.mp3" type="audio/mpeg"/>
  </item>
</channel></rss>`
	_, eps, err := parseRSS(feed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eps) != 1 {
		t.Fatalf("want 1 episode, got %d", len(eps))
	}
	if eps[0].ID != "https://example.com/noguid.mp3" {
		t.Errorf("want enclosure URL as ID, got %q", eps[0].ID)
	}
}

func TestParseRSS_InvalidXMLReturnsError(t *testing.T) {
	_, _, err := parseRSS("not xml at all")
	if err == nil {
		t.Error("want error for invalid XML, got nil")
	}
}

func TestParseRSS_EmptyFeed(t *testing.T) {
	feed := `<?xml version="1.0"?><rss version="2.0"><channel><title>Empty</title></channel></rss>`
	pod, eps, err := parseRSS(feed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pod.Title != "Empty" {
		t.Errorf("want title %q, got %q", "Empty", pod.Title)
	}
	if len(eps) != 0 {
		t.Errorf("want 0 episodes, got %d", len(eps))
	}
}

func TestParseRSS_OutputIsJSONSerializable(t *testing.T) {
	pod, eps, err := parseRSS(minimalFeed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out, err := json.Marshal(map[string]any{"podcast": pod, "episodes": eps})
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	if !strings.Contains(string(out), "Test Podcast") {
		t.Error("serialized output does not contain podcast title")
	}
}
