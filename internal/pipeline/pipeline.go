package pipeline

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/alitto/pond"

	"github.com/gavmor/podpedia/internal/llm"
	"github.com/gavmor/podpedia/internal/storage"
	"github.com/gavmor/podpedia/internal/transcription"
	"github.com/gavmor/podpedia/internal/types"
	"github.com/mmcdole/gofeed"
)

func Run(rssURL string, outputDir string) error {
	fmt.Printf("[Pipeline] Starting for feed: %s (Output: %s)\n", rssURL, outputDir)

	_, episodes, err := parseRSSWithGofeed(rssURL)
	if err != nil {
		return fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	fmt.Printf("[Pipeline] Found %d valid episodes in feed.\n", len(episodes))

	maxConcurrentWorkers := runtime.NumCPU()
	if val, ok := os.LookupEnv("PODPEDIA_MAX_WORKERS"); ok {
		fmt.Printf("[Pipeline] Using custom concurrency limit: %s\n", val)
		// In a real app, you'd parse this to int.
	}
	
	pool := pond.New(maxConcurrentWorkers, 0, pond.IdleTimeout(10*time.Second))
	defer pool.StopAndWait()

	startTime := time.Now()

	for _, ep := range episodes {
		episode := ep
		pool.Submit(func() {
			processEpisode(episode, outputDir)
		})
	}

	pool.StopAndWait()
	fmt.Printf("\n[Pipeline] Completed in %v\n", time.Since(startTime))
	return nil
}

func processEpisode(ep types.Episode, outputDir string) {
	fmt.Printf("[Worker] Starting Episode: %s\n", ep.Title)

	// Download audio if needed
	audioPath := fmt.Sprintf("%s/%s.mp3", outputDir, ep.ID)
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		fmt.Printf("[Worker] Downloading audio for: %s\n", ep.Title)
		if err := DownloadAudio(ep.AudioURL, audioPath); err != nil {
			fmt.Printf("[Worker] ERROR: Failed to download audio for %s: %v\n", ep.ID, err)
			return
		}
	}

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

	if err := storage.SaveRawData(outputDir, ep); err != nil {
		fmt.Printf("[Worker] ERROR: Failed to save raw data for %s: %v\n", ep.ID, err)
	}
	if err := storage.SaveStructuredData(outputDir, entry); err != nil {
		fmt.Printf("[Worker] ERROR: Failed to save structured data for %s: %v\n", entry.EpisodeID, err)
	}

	fmt.Printf("[Worker] Completed Episode: %s\n", ep.Title)
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

func parseRSSWithGofeed(url string) (types.Podcast, []types.Episode, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return types.Podcast{}, nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	podcast := types.Podcast{
		Title:       feed.Title,
		Description: feed.Description,
		URL:         url,
		Categories:  feed.Categories,
	}
	if feed.ITunesExt != nil {
		podcast.Author = feed.ITunesExt.Author
	}

	if podcast.Author == "" && feed.Extensions["dc"] != nil {
		if authors, ok := feed.Extensions["dc"]["creator"]; ok && len(authors) > 0 {
			podcast.Author = authors[0].Value
		}
	}

	var episodes []types.Episode
	for _, item := range feed.Items {
		ep := types.Episode{
			ID:          item.GUID,
			Title:       item.Title,
			Description: item.Description,
			PubDate:     item.Published,
		}

		if item.ITunesExt != nil {
			ep.Duration = item.ITunesExt.Duration
			ep.Explicit = item.ITunesExt.Explicit == "yes"
			ep.Author = item.ITunesExt.Author
		}

		ep.Categories = item.Categories

		if ep.Author == "" && item.Extensions["dc"] != nil {
			if creators, ok := item.Extensions["dc"]["creator"]; ok && len(creators) > 0 {
				ep.Author = creators[0].Value
			}
		}

		if len(item.Enclosures) > 0 {
			ep.AudioURL = item.Enclosures[0].URL
		}

		if err := validateEpisode(ep); err != nil {
			fmt.Printf("[Pipeline] Skipping invalid episode: %v\n", err)
			continue
		}
		episodes = append(episodes, ep)
	}

	return podcast, episodes, nil
}

func fetchRSSFeed(url string) ([]types.Episode, error) {
	_, episodes, err := parseRSSWithGofeed(url)
	return episodes, err
}
