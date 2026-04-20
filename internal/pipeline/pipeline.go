package pipeline

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"code.cloudfoundry.org/lager/v3"
	"github.com/alitto/pond"
	"github.com/gavmor/podpedia/internal/types"
	"github.com/mmcdole/gofeed"
)

// Consumer-driven interfaces
type Transcriber interface {
	Transcribe(audioURL string) (string, error)
}

type Extractor interface {
	ExtractEntities(ep types.Episode) (types.EncyclopediaEntry, error)
}

type AudioDownloader interface {
	DownloadAudio(url string, dest string) error
}

type Store interface {
	SaveRawData(outputDir string, ep types.Episode) error
	SaveStructuredData(outputDir string, entry types.EncyclopediaEntry) error
}

type Pipeline struct {
	logger       lager.Logger
	transcriber  Transcriber
	extractor    Extractor
	downloader   AudioDownloader
	store        Store
	maxWorkers   int
}

func NewPipeline(
	logger lager.Logger,
	transcriber Transcriber,
	extractor Extractor,
	downloader AudioDownloader,
	store Store,
) *Pipeline {
	return &Pipeline{
		logger:      logger,
		transcriber: transcriber,
		extractor:   extractor,
		downloader:  downloader,
		store:       store,
		maxWorkers:  runtime.NumCPU(),
	}
}

func (p *Pipeline) Run(rssURL string, outputDir string) error {
	p.logger.Info("starting", lager.Data{"feed": rssURL, "output": outputDir})

	_, episodes, err := ParseRSSWithGofeed(rssURL)
	if err != nil {
		p.logger.Error("failed-parsing-feed", err)
		return fmt.Errorf("failed to fetch RSS feed: %w", err)
	}

	p.logger.Info("found-episodes", lager.Data{"count": len(episodes)})

	pool := pond.New(p.maxWorkers, 0, pond.IdleTimeout(10*time.Second))
	defer pool.StopAndWait()

	startTime := time.Now()

	for _, ep := range episodes {
		episode := ep
		pool.Submit(func() {
			p.processEpisode(episode, outputDir)
		})
	}

	pool.StopAndWait()
	p.logger.Info("completed", lager.Data{"duration": time.Since(startTime).String()})
	return nil
}

func (p *Pipeline) processEpisode(ep types.Episode, outputDir string) {
	lsess := p.logger.Session("process-episode", lager.Data{"episode": ep.ID})
	lsess.Info("starting")

	// Download audio if needed
	audioPath := fmt.Sprintf("%s/%s.mp3", outputDir, ep.ID)
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		lsess.Info("downloading-audio")
		if err := p.downloader.DownloadAudio(ep.AudioURL, audioPath); err != nil {
			lsess.Error("failed-download", err)
			return
		}
	}

	if ep.Transcript == "" {
		lsess.Info("transcribing")
		transcript, err := p.transcriber.Transcribe(ep.AudioURL)
		if err != nil {
			lsess.Error("failed-transcription", err)
			return
		}
		ep.Transcript = transcript
	}

	lsess.Info("extracting-entities")
	entry, err := p.extractor.ExtractEntities(ep)
	if err != nil {
		lsess.Error("failed-extraction", err)
		return
	}

	if err := p.store.SaveRawData(outputDir, ep); err != nil {
		lsess.Error("failed-save-raw", err)
	}
	if err := p.store.SaveStructuredData(outputDir, entry); err != nil {
		lsess.Error("failed-save-structured", err)
	}

	lsess.Info("completed")
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

func ParseRSSWithGofeed(url string) (types.Podcast, []types.Episode, error) {
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

