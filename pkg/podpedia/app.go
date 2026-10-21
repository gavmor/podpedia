package podpedia

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"code.cloudfoundry.org/lager/v3"
	"github.com/gavmor/podpedia/internal/cache"
	"github.com/gavmor/podpedia/internal/kernel"
	"github.com/gavmor/podpedia/internal/pipeline"
	"github.com/gavmor/podpedia/internal/types"
	"github.com/spf13/afero"
)

// Cache environment overrides, consulted when the corresponding option is not
// set explicitly.
const (
	// CacheDirEnv overrides where the engine keeps its artifact cache. Set it
	// to an empty value to turn caching off.
	CacheDirEnv = "PODPEDIA_CACHE_DIR"
	// CacheMaxBytesEnv overrides the cache's total size budget, in bytes.
	CacheMaxBytesEnv = "PODPEDIA_CACHE_MAX_BYTES"
)

// DefaultCacheMaxBytes bounds the artifact cache when nothing else says
// otherwise. See cache.DefaultMaxBytes.
const DefaultCacheMaxBytes = cache.DefaultMaxBytes

type App struct {
	kernel   *kernel.PodpediaKernel
	pipeline *pipeline.Pipeline
	logger   lager.Logger
}

type Option func(*config)

type config struct {
	logger            lager.Logger
	workspace         afero.Fs
	ollamaURL         string
	ollamaModel       string
	transcribeURL     string
	episodeLimit      int
	outputScheme      []byte
	schemeID          string
	onEpisodeComplete func(ep types.Episode, err error)

	// nil means "not set by the caller", which is distinct from an explicit
	// empty directory (caching off) or an explicit zero budget.
	cacheDir      *string
	cacheMaxBytes *int64
}

func WithLogger(logger lager.Logger) Option {
	return func(c *config) {
		c.logger = logger
	}
}

func WithWorkspace(fs afero.Fs) Option {
	return func(c *config) {
		if fs != nil {
			c.workspace = fs
		}
	}
}

func WithOllamaURL(url string) Option {
	return func(c *config) {
		c.ollamaURL = url
	}
}

func WithOllamaModel(model string) Option {
	return func(c *config) {
		c.ollamaModel = model
	}
}

func WithTranscribeURL(url string) Option {
	return func(c *config) {
		c.transcribeURL = url
	}
}

func WithLimit(limit int) Option {
	return func(c *config) {
		c.episodeLimit = limit
	}
}

// WithScheme sets the JSON structure the extractor should fill in, and the id
// that names the structured output file for it (see naming.StructuredName —
// id "entry" yields <slug>_entry.json).
//
// A nil or empty scheme with a non-empty id is a supported combination: it
// names the output without dictating a schema, leaving the extractor plugin
// to use its own default. Do not pass a placeholder such as []byte(`{}`) to
// get at the filename — that is a schema as far as the extractor is
// concerned, and it is what the model will be told to produce.
func WithScheme(scheme []byte, id string) Option {
	return func(c *config) {
		c.outputScheme = scheme
		c.schemeID = id
	}
}

func WithOnEpisodeComplete(callback func(ep types.Episode, err error)) Option {
	return func(c *config) {
		c.onEpisodeComplete = callback
	}
}

// WithCacheDir points the engine's persistent artifact cache at dir. An empty
// dir disables caching.
//
// The cache holds the expensive outputs of a run — ASR transcripts and LLM
// extractions — keyed by each episode's enclosure URL. It is deliberately not
// the output directory passed to Run: that one is scratch, and a caller is
// free to hand over an empty one per job and delete it afterwards. dir must
// therefore be somewhere persistent that no caller clears.
func WithCacheDir(dir string) Option {
	return func(c *config) {
		c.cacheDir = &dir
	}
}

// WithCacheMaxBytes bounds the artifact cache's total size. Once the budget
// is exceeded, entries are dropped least-recently-used first. A non-positive
// value selects DefaultCacheMaxBytes.
func WithCacheMaxBytes(n int64) Option {
	return func(c *config) {
		c.cacheMaxBytes = &n
	}
}

// WithoutCache turns the artifact cache off, so every stage runs on every
// pass. Equivalent to WithCacheDir("").
func WithoutCache() Option {
	return WithCacheDir("")
}

func NewApp(opts ...Option) (*App, error) {
	c := &config{
		ollamaURL:   "http://localhost:11434",
		ollamaModel: "qwen2.5:0.5b",
		workspace:   afero.NewOsFs(),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.logger == nil {
		c.logger = lager.NewLogger("podpedia")
		// Default to no-op logger if none provided? Or stdout?
		// For now, let's just make it required or use a default.
	}

	k := kernel.NewPodpediaKernel(c.ollamaURL, c.ollamaModel, c.transcribeURL)

	p := pipeline.NewPipeline(
		c.logger,
		kernel.NewTranscriber(k),
		kernel.NewExtractor(k),
		kernel.NewDownloader(k),
		kernel.NewStore(k),
	).WithWorkspace(c.workspace)

	if c.episodeLimit > 0 {
		p.WithLimit(c.episodeLimit)
	}

	// A scheme id with no scheme is meaningful — it names the structured
	// output file and lets the extractor plugin pick its own schema — so
	// the id alone is enough to forward.
	if len(c.outputScheme) > 0 || c.schemeID != "" {
		p.WithScheme(c.outputScheme, c.schemeID)
	}

	if c.onEpisodeComplete != nil {
		p.OnEpisodeComplete(c.onEpisodeComplete)
	}

	if store := openCache(c); store != nil {
		c.logger.Info("artifact-cache-enabled", lager.Data{
			"dir":       store.Dir(),
			"max_bytes": store.MaxBytes(),
		})
		p.WithCache(store)
	} else {
		c.logger.Info("artifact-cache-disabled")
	}

	return &App{
		kernel:   k,
		pipeline: p,
		logger:   c.logger,
	}, nil
}

// openCache resolves the artifact cache's location and budget and opens it.
// It returns nil when caching is switched off or cannot be set up; a cache is
// an optimisation, so failing to open one must never fail the run.
func openCache(c *config) *cache.Store {
	dir, err := cacheDir(c)
	if err != nil {
		c.logger.Error("failed-resolving-cache-dir", err)
		return nil
	}
	if dir == "" {
		return nil
	}

	maxBytes, err := cacheMaxBytes(c)
	if err != nil {
		c.logger.Error("failed-resolving-cache-budget", err)
		maxBytes = DefaultCacheMaxBytes
	}

	store, err := cache.NewStore(dir, maxBytes)
	if err != nil {
		c.logger.Error("failed-opening-artifact-cache", err)
		return nil
	}
	return store
}

// cacheDir resolves where the artifact cache lives: the explicit option wins,
// then CacheDirEnv, then a subdirectory of the user cache directory. An empty
// result means caching is off.
func cacheDir(c *config) (string, error) {
	if c.cacheDir != nil {
		return strings.TrimSpace(*c.cacheDir), nil
	}
	if dir, ok := os.LookupEnv(CacheDirEnv); ok {
		return strings.TrimSpace(dir), nil
	}

	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("no %s set and no user cache directory: %w", CacheDirEnv, err)
	}
	return filepath.Join(base, "podpedia", "artifacts"), nil
}

// cacheMaxBytes resolves the cache's size budget: the explicit option wins,
// then CacheMaxBytesEnv, then DefaultCacheMaxBytes.
func cacheMaxBytes(c *config) (int64, error) {
	if c.cacheMaxBytes != nil {
		return *c.cacheMaxBytes, nil
	}
	raw, ok := os.LookupEnv(CacheMaxBytesEnv)
	if !ok || strings.TrimSpace(raw) == "" {
		return DefaultCacheMaxBytes, nil
	}
	n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil {
		return DefaultCacheMaxBytes, fmt.Errorf("invalid %s=%q: %w", CacheMaxBytesEnv, raw, err)
	}
	return n, nil
}

func (a *App) LoadPlugin(name string, wasmBytes []byte) {
	a.kernel.Load(name, wasmBytes)
}

func (a *App) Run(ctx context.Context, rssURL string, outputDir string) error {
	return a.pipeline.Run(ctx, rssURL, outputDir)
}

func (a *App) Close() error {
	if a.kernel != nil {
		return a.kernel.Close()
	}
	return nil
}
