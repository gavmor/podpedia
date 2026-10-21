package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"code.cloudfoundry.org/lager/v3"
	"github.com/alitto/pond"
	"github.com/gavmor/podpedia/internal/types"
	"github.com/gavmor/podpedia/pkg/naming"
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

// ArtifactCache is a persistent store for the expensive outputs of a run —
// ASR transcripts and LLM extractions. Keys are opaque strings (see
// episodeCacheKey); values are bytes.
//
// It is deliberately NOT the workspace. The workspace is scratch: a caller is
// entitled to hand over an empty directory for every run and delete it
// afterwards, so the presence or absence of a file there says nothing about
// whether the work has been done before. A cache implementation lives
// somewhere the engine owns and that survives runs.
//
// A nil cache disables caching; every lookup then misses and every stage runs.
type ArtifactCache interface {
	// Get returns the value for key; the boolean reports a hit. A failure to
	// read is reported as a miss, since recomputing is the only response.
	Get(key string) ([]byte, bool)
	// Put stores data under key.
	Put(key string, data []byte) error
}

// Pipeline coordinates the end-to-end processing of a podcast feed.
type Pipeline struct {
	logger            lager.Logger
	transcriber       Transcriber
	extractor         Extractor
	downloader        AudioDownloader
	store             Store
	cache             ArtifactCache
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

// WithCache attaches the persistent artifact cache consulted around the
// transcription and extraction stages. A nil cache leaves caching off.
func (p *Pipeline) WithCache(c ArtifactCache) *Pipeline {
	p.cache = c
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
//
// Every expensive stage is guarded by the engine's persistent cache, keyed by
// the episode's enclosure URL. The workspace is treated purely as this run's
// output directory: it is never read to decide whether work can be skipped,
// because the caller owns it and may legitimately supply an empty one each
// time. Consequently a cache hit still materialises the episode's files in
// the workspace, so a caller that harvests the directory sees the same result
// whether the work was computed or recalled.
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

	transcriptKey := episodeCacheKey("transcript", ep.AudioURL)

	// --- Transcription ---
	if ep.Transcript == "" {
		if cached, ok := p.cacheGet(transcriptKey); ok {
			lsess.Info("cache-hit-transcript", lager.Data{"bytes": len(cached)})
			ep.Transcript = string(cached)
		}
	}

	if ep.Transcript == "" {
		// Download audio only if we actually need to transcribe.
		audioPath := fmt.Sprintf("%s/%s", outputDir, naming.AudioName(id))
		lsess.Info("downloading-audio")
		if err := p.downloader.DownloadAudio(ctx, p.Workspace, ep.AudioURL, audioPath); err != nil {
			lsess.Error("failed-download", err)
			return fmt.Errorf("failed to download audio: %w", err)
		}

		lsess.Info("transcribing")
		hint := p.generateTranscriptionHint(ep)
		transcript, err := p.transcriber.Transcribe(ctx, ep.AudioURL, hint)
		if err != nil {
			lsess.Error("failed-transcription", err)
			return fmt.Errorf("failed transcription: %w", err)
		}
		ep.Transcript = transcript
		p.cachePut(lsess, transcriptKey, []byte(transcript))
	}

	// Write the transcript into this run's output directory, whether it was
	// just produced or recalled from the cache.
	if err := p.store.SaveRawData(ctx, p.Workspace, outputDir, ep); err != nil {
		lsess.Error("failed-save-raw", err)
		return fmt.Errorf("failed to save raw data: %w", err)
	}

	// --- Extraction ---
	// The scheme decides the shape of the answer, so it is part of the key:
	// the same episode under a different scheme is a different artifact.
	entryKey := episodeCacheKey("entry", ep.AudioURL, p.schemeID, string(p.outputScheme))

	entry, ok := p.cacheGet(entryKey)
	if ok {
		lsess.Info("cache-hit-entry", lager.Data{"bytes": len(entry)})
	} else {
		lsess.Info("extracting-entities")
		var err error
		entry, err = p.extractor.ExtractEntities(ctx, ep, p.outputScheme)
		if err != nil {
			lsess.Error("failed-extraction", err)
			return fmt.Errorf("failed extraction: %w", err)
		}
		p.cachePut(lsess, entryKey, entry)
	}

	if err := p.store.SaveStructuredData(ctx, p.Workspace, outputDir, ep, entry, p.schemeID); err != nil {
		lsess.Error("failed-save-structured", err)
		return fmt.Errorf("failed to save structured data: %w", err)
	}

	lsess.Info("completed")
	return nil
}

// episodeCacheKey derives a cache key from an episode's enclosure URL.
//
// The URL is the episode's stable identity: a publisher re-titling an episode
// or regenerating its <guid> leaves the audio it points at — and therefore the
// transcript and the extraction — unchanged, so neither should cost a miss.
// This is deliberately not naming.Slug: that is the filename contract for a
// workspace, derived from the identifier, and it is exactly the brittle key
// the cache must avoid. The key is a hash, so it is safe as a filename or an
// object key regardless of what the URL contains, and further discriminators
// (an extraction scheme, say) fold in without needing a quoting scheme.
//
// An episode with no enclosure URL has no stable identity, so it is reported
// as uncacheable and every stage runs for it.
func episodeCacheKey(kind string, audioURL string, extra ...string) string {
	if audioURL == "" {
		return ""
	}
	h := sha256.New()
	for _, part := range append([]string{audioURL}, extra...) {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0}) // unambiguous separator
	}
	return kind + "-" + hex.EncodeToString(h.Sum(nil))
}

// cacheGet reads key, treating a disabled cache or an unusable key as a miss.
func (p *Pipeline) cacheGet(key string) ([]byte, bool) {
	if p.cache == nil || key == "" {
		return nil, false
	}
	return p.cache.Get(key)
}

// cachePut stores data under key. A cache write failure is logged and
// swallowed: the run already has the value in hand, and the only cost is that
// the next run recomputes it.
func (p *Pipeline) cachePut(lsess lager.Logger, key string, data []byte) {
	if p.cache == nil || key == "" {
		return
	}
	if err := p.cache.Put(key, data); err != nil {
		lsess.Error("failed-cache-write", err, lager.Data{"key": key})
	}
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
