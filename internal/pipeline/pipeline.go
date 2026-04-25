package pipeline

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"code.cloudfoundry.org/lager/v3"
	"github.com/alitto/pond"
	"github.com/gavmor/podpedia/internal/types"
	"github.com/mmcdole/gofeed"
)

// Consumer-driven interfaces
type Transcriber interface {
	Transcribe(ep types.Episode) (string, error)
}

type Extractor interface {
	ExtractEntities(ep types.Episode, scheme []byte) ([]byte, error)
}

type AudioDownloader interface {
	DownloadAudio(url string, dest string) error
}

type Store interface {
	SaveRawData(outputDir string, ep types.Episode) error
	SaveStructuredData(outputDir string, ep types.Episode, entry []byte, schemeID string) error
}

type Pipeline struct {
	logger       lager.Logger
	transcriber  Transcriber
	extractor    Extractor
	downloader   AudioDownloader
	store        Store
	maxWorkers   int
	limit        int
	outputScheme []byte
	schemeID     string
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

func (p *Pipeline) WithLimit(n int) *Pipeline {
	p.limit = n
	return p
}

func (p *Pipeline) WithScheme(scheme []byte, id string) *Pipeline {
	p.outputScheme = scheme
	p.schemeID = id
	return p
}

func (p *Pipeline) Run(rssURL string, outputDir string) error {
	p.logger.Info("starting", lager.Data{"feed": rssURL, "output": outputDir, "scheme": p.schemeID})

	_, episodes, err := ParseRSSWithGofeed(rssURL)
	if err != nil {
		p.logger.Error("failed-parsing-feed", err)
		return fmt.Errorf("failed to fetch RSS feed: %w", err)
	}

	if p.limit > 0 && len(episodes) > p.limit {
		p.logger.Info("applying-limit", lager.Data{"total": len(episodes), "limit": p.limit})
		episodes = episodes[:p.limit]
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

	// Transcribe if needed
	transcriptPath := fmt.Sprintf("%s/%s_raw.txt", outputDir, slug(ep.ID))
	if _, err := os.Stat(transcriptPath); err == nil {
		lsess.Info("transcript-already-exists-loading")
		content, err := os.ReadFile(transcriptPath)
		if err == nil {
			ep.Transcript = string(content)
		} else {
			lsess.Error("failed-read-existing-transcript", err)
		}
	}

	if ep.Transcript == "" {
		// Download audio only if we actually need to transcribe
		audioPath := fmt.Sprintf("%s/%s.mp3", outputDir, slug(ep.ID))
		if _, err := os.Stat(audioPath); err == nil {
			lsess.Info("audio-already-exists-skipping-download")
		} else if os.IsNotExist(err) {
			lsess.Info("downloading-audio")
			if err := p.downloader.DownloadAudio(ep.AudioURL, audioPath); err != nil {
				lsess.Error("failed-download", err)
				return
			}
		} else {
			lsess.Error("failed-stat-audio", err)
			return
		}

		lsess.Info("transcribing")
		transcript, err := p.transcriber.Transcribe(ep)
		if err != nil {
			lsess.Error("failed-transcription", err)
			return
		}
		ep.Transcript = transcript

		// Save raw data immediately after transcription
		if err := p.store.SaveRawData(outputDir, ep); err != nil {
			lsess.Error("failed-save-raw", err)
		}
	}

	lsess.Info("extracting-entities")
	entry, err := p.extractor.ExtractEntities(ep, p.outputScheme)
	if err != nil {
		lsess.Error("failed-extraction", err)
		// No reliable generic "minimal" entry for arbitrary schemes
		return
	}

	if err := p.store.SaveStructuredData(outputDir, ep, entry, p.schemeID); err != nil {
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

func slug(s string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, s)
}
