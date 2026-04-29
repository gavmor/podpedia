package pipeline

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"code.cloudfoundry.org/lager/v3"
	"github.com/alitto/pond"
	"github.com/gavmor/podpedia/internal/types"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/afero"
)

// Consumer-driven interfaces
type Transcriber interface {
	Transcribe(ctx context.Context, audioURL string, prompt string) (string, error)
}

type Extractor interface {
	ExtractEntities(ctx context.Context, ep *types.Episode, scheme []byte) ([]byte, error)
}

type AudioDownloader interface {
	DownloadAudio(ctx context.Context, fs afero.Fs, url string, dest string) error
}

type Store interface {
	SaveRawData(ctx context.Context, fs afero.Fs, outputDir string, ep *types.Episode) error
	SaveStructuredData(ctx context.Context, fs afero.Fs, outputDir string, ep *types.Episode, entry []byte, schemeID string) error
}

// Pipeline coordinates the end-to-end processing of a podcast feed.
type Pipeline struct {
	logger            lager.Logger
	transcriber       Transcriber
	extractor         Extractor
	downloader        AudioDownloader
	store             Store
	Workspace         afero.Fs
	onEpisodeComplete func(ep types.Episode, err error)
	maxWorkers        int
	limit             int
	outputScheme      []byte
	schemeID          string
}

// NewPipeline creates a new instance of the processing pipeline.
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
		Workspace:   afero.NewOsFs(),
		maxWorkers:  runtime.NumCPU(),
	}
}

func (p *Pipeline) OnEpisodeComplete(callback func(ep types.Episode, err error)) *Pipeline {
	p.onEpisodeComplete = callback
	return p
}

func (p *Pipeline) WithWorkspace(fs afero.Fs) *Pipeline {
	if fs != nil {
		p.Workspace = fs
	}
	return p
}

// WithLimit sets a maximum number of episodes to process.
func (p *Pipeline) WithLimit(n int) *Pipeline {
	p.limit = n
	return p
}

// WithScheme sets the extraction scheme for the pipeline.
func (p *Pipeline) WithScheme(scheme []byte, id string) *Pipeline {
	p.outputScheme = scheme
	p.schemeID = id
	return p
}

// Run executes the pipeline for a given RSS URL and output directory.
func (p *Pipeline) Run(ctx context.Context, rssURL string, outputDir string) error {
	p.logger.Info("starting", lager.Data{"feed": rssURL, "output": outputDir, "scheme": p.schemeID})

	_, episodes, err := ParseRSSWithGofeed(ctx, rssURL)
	if err != nil {
		p.logger.Error("failed-parsing-feed", err)
		return fmt.Errorf("failed to fetch RSS feed: %v", err)
	}

	if p.limit > 0 && len(episodes) > p.limit {
		p.logger.Info("applying-limit", lager.Data{"total": len(episodes), "limit": p.limit})
		episodes = episodes[:p.limit]
	}

	p.logger.Info("found-episodes", lager.Data{"count": len(episodes)})

	pool := pond.New(p.maxWorkers, 0, pond.IdleTimeout(10*time.Second))
	defer pool.StopAndWait()

	group, groupCtx := pool.GroupContext(ctx)

	var mu sync.Mutex
	var errs []error

	startTime := time.Now()

	for _, ep := range episodes {
		episode := ep
		group.Submit(func() error {
			err := p.processEpisode(groupCtx, &episode, outputDir)

			if p.onEpisodeComplete != nil {
				p.onEpisodeComplete(episode, err)
			}

			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
			return err
		})
	}

	if err := group.Wait(); err != nil {
		mu.Lock()
		errs = append(errs, err)
		mu.Unlock()
	}

	if len(errs) > 0 {
		combinedErr := errors.Join(errs...)
		p.logger.Error("pipeline-failed", combinedErr)
		return combinedErr
	}

	p.logger.Info("completed", lager.Data{"duration": time.Since(startTime).String()})
	return nil
}

func (p *Pipeline) ProcessEpisode(ctx context.Context, ep *types.Episode, outputDir string) error {
	return p.processEpisode(ctx, ep, outputDir)
}

// processEpisode handles the full processing lifecycle of a single episode.
func (p *Pipeline) processEpisode(ctx context.Context, ep *types.Episode, outputDir string) error {
	id := ep.ID
	if id == "" {
		id = ep.AudioURL
	}
	lsess := p.logger.Session("process-episode", lager.Data{"episode": id})
	lsess.Info("starting")

	// Check if context is already cancelled
	if ctx.Err() != nil {
		lsess.Info("context-cancelled-aborting")
		return ctx.Err()
	}

	// --- Idempotency Check ---
	// Skip if both the final structured data and metadata sidecar already exist.
	entryID := slug(id)
	structuredPath := fmt.Sprintf("%s/%s_%s.json", outputDir, entryID, p.schemeID)
	metaPath := fmt.Sprintf("%s/%s_meta.json", outputDir, entryID)

	if _, err := p.Workspace.Stat(structuredPath); err == nil {
		if _, err := p.Workspace.Stat(metaPath); err == nil {
			lsess.Info("skipping-already-processed")
			return nil
		}
	}
	// --- End Idempotency Check ---

	// Transcribe if needed
	transcriptPath := fmt.Sprintf("%s/%s_raw.txt", outputDir, entryID)
	if _, err := p.Workspace.Stat(transcriptPath); err == nil {
		lsess.Info("transcript-already-exists-loading")
		content, err := afero.ReadFile(p.Workspace, transcriptPath)
		if err == nil {
			ep.Transcript = string(content)
		} else {
			lsess.Error("failed-read-existing-transcript", err)
			return fmt.Errorf("failed to read existing transcript: %w", err)
		}
	}

	if ep.Transcript == "" {
		// Download audio only if we actually need to transcribe
		audioPath := fmt.Sprintf("%s/%s.mp3", outputDir, entryID)
		exists, _ := afero.Exists(p.Workspace, audioPath)
		if exists {
			lsess.Info("audio-already-exists-skipping-download")
		} else {
			lsess.Info("downloading-audio")
			if err := p.downloader.DownloadAudio(ctx, p.Workspace, ep.AudioURL, audioPath); err != nil {
				lsess.Error("failed-download", err)
				return fmt.Errorf("failed to download audio: %w", err)
			}
		}

		lsess.Info("transcribing")
		hint := p.generateTranscriptionHint(ep)
		transcript, err := p.transcriber.Transcribe(ctx, ep.AudioURL, hint)
		if err != nil {
			lsess.Error("failed-transcription", err)
			return fmt.Errorf("failed transcription: %w", err)
		}
		ep.Transcript = transcript

		// Save raw data immediately after transcription
		if err := p.store.SaveRawData(ctx, p.Workspace, outputDir, ep); err != nil {
			lsess.Error("failed-save-raw", err)
			return fmt.Errorf("failed to save raw data: %w", err)
		}
	}

	lsess.Info("extracting-entities")
	entry, err := p.extractor.ExtractEntities(ctx, ep, p.outputScheme)
	if err != nil {
		lsess.Error("failed-extraction", err)
		return fmt.Errorf("failed extraction: %w", err)
	}

	if err := p.store.SaveStructuredData(ctx, p.Workspace, outputDir, ep, entry, p.schemeID); err != nil {
		lsess.Error("failed-save-structured", err)
		return fmt.Errorf("failed to save structured data: %w", err)
	}

	lsess.Info("completed")
	return nil
}

// generateTranscriptionHint builds an ASR initial prompt from podcast episode metadata.
func (p *Pipeline) generateTranscriptionHint(ep *types.Episode) string {
	switch {
	case ep.Title != "" && ep.Description != "":
		return ep.Title + "\n\n" + ep.Description
	case ep.Title != "":
		return ep.Title
	default:
		return ep.Description
	}
}

// validateEpisode checks if an episode has all required fields.
func validateEpisode(ep *types.Episode) error {
	if ep.ID == "" {
		return fmt.Errorf("missing episode ID")
	}
	if ep.Title == "" {
		return fmt.Errorf("missing episode title")
	}
	if ep.AudioURL == "" {
		return fmt.Errorf("missing audio URL for episode: %s", ep.Title)
	}
	return nil
}

// ParseRSSWithGofeed fetches and parses a podcast RSS feed into internal types.
func ParseRSSWithGofeed(ctx context.Context, url string) (types.Podcast, []types.Episode, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(url, ctx)
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

		if err := validateEpisode(&ep); err != nil {
			fmt.Printf("[Pipeline] Skipping invalid episode: %v\n", err)
			continue
		}
		episodes = append(episodes, ep)
	}

	return podcast, episodes, nil
}

// slug normalizes a string for use in filenames by keeping only alphanumeric characters.
func slug(s string) string {
	if s == "" {
		return "unknown"
	}
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, s)
}

// Slug normalizes a string for use in filenames by keeping only alphanumeric characters.
func Slug(s string) string {
	return slug(s)
}
