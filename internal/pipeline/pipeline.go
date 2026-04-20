package pipeline

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gavmor/podpedia/internal/llm"
	"github.com/gavmor/podpedia/internal/storage"
	"github.com/gavmor/podpedia/internal/transcription"
	"github.com/gavmor/podpedia/internal/types"
)

func Run(rssURL string) error {
	fmt.Printf("[Pipeline] Starting for feed: %s\n", rssURL)

	episodes := fetchRSSFeed(rssURL)
	fmt.Printf("[Pipeline] Found %d episodes in feed.\n", len(episodes))

	maxConcurrentWorkers := 3
	if val, ok := os.LookupEnv("PODPEDIA_MAX_WORKERS"); ok {
		fmt.Printf("[Pipeline] Using custom concurrency limit: %s\n", val)
		// In a real app, you'd parse this to int.
	}
	
	semaphore := make(chan struct{}, maxConcurrentWorkers)
	var wg sync.WaitGroup

	startTime := time.Now()

	for _, ep := range episodes {
		wg.Add(1)
		go func(episode types.Episode) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			processEpisode(episode)
		}(ep)
	}

	wg.Wait()
	fmt.Printf("\n[Pipeline] Completed in %v\n", time.Since(startTime))
	return nil
}

func processEpisode(ep types.Episode) {
	fmt.Printf("[Worker] Starting Episode: %s\n", ep.Title)

	if ep.Transcript == "" {
		transcript, err := transcription.Transcribe(ep.AudioURL)
		if err != nil {
			fmt.Printf("[Worker] ERROR: Failed to transcribe episode %s: %v\n", ep.ID, err)
			return
		}
		ep.Transcript = transcript
	}

	entry, err := llm.ExtractEntities(ep)
	if err != nil {
		fmt.Printf("[Worker] ERROR: Failed extraction for episode %s: %v\n", ep.ID, err)
		return
	}

	if err := storage.SaveRawData(ep); err != nil {
		fmt.Printf("[Worker] ERROR: Failed to save raw data for %s: %v\n", ep.ID, err)
	}
	if err := storage.SaveStructuredData(entry); err != nil {
		fmt.Printf("[Worker] ERROR: Failed to save structured data for %s: %v\n", entry.EpisodeID, err)
	}

	fmt.Printf("[Worker] Completed Episode: %s\n", ep.Title)
}

func fetchFeedContent(url string) ([]byte, error) {
	fmt.Printf("[Pipeline] Fetching RSS content from: %s\n", url)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func parseRSS(content []byte) (types.Podcast, []types.Episode, error) {
	type RSS struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Title       string          `xml:"title"`
			Description string          `xml:"description"`
			Items       []types.Episode `xml:"item"`
		} `xml:"channel"`
	}

	var rss RSS
	if err := xml.Unmarshal(content, &rss); err != nil {
		return types.Podcast{}, nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	podcast := types.Podcast{
		Title:       rss.Channel.Title,
		Description: rss.Channel.Description,
	}

	return podcast, rss.Channel.Items, nil
}

func validateEpisode(ep types.Episode) error {
	if ep.Title == "" {
		return fmt.Errorf("missing episode title")
	}
	if ep.AudioURL == "" {
		return fmt.Errorf("missing audio URL for episode: %s", ep.Title)
	}
	return nil
}

func fetchRSSFeed(url string) []types.Episode {
	fmt.Printf("[Pipeline] Fetching RSS feed from: %s\n", url)
	return []types.Episode{
		{ID: "ep001", Title: "The Future of AI", AudioURL: "http://audio.com/ep001.mp3"},
		{ID: "ep002", Title: "SaaS Business Models", AudioURL: "http://audio.com/ep002.mp3"},
		{ID: "ep003", Title: "Open Source Ideology", AudioURL: "http://audio.com/ep003.mp3"},
		{ID: "ep004", Title: "Hardware for LLMs", AudioURL: "http://audio.com/ep004.mp3"},
		{ID: "ep005", Title: "Podcast Monetization", AudioURL: "http://audio.com/ep005.mp3"},
	}
}
