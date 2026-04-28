package podpedia

import (
	"context"

	"code.cloudfoundry.org/lager/v3"
	"github.com/gavmor/podpedia/internal/kernel"
	"github.com/gavmor/podpedia/internal/pipeline"
	"github.com/gavmor/podpedia/internal/types"
	"github.com/spf13/afero"
)

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

	if len(c.outputScheme) > 0 {
		p.WithScheme(c.outputScheme, c.schemeID)
	}

	if c.onEpisodeComplete != nil {
		p.OnEpisodeComplete(c.onEpisodeComplete)
	}

	return &App{
		kernel:   k,
		pipeline: p,
		logger:   c.logger,
	}, nil
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
