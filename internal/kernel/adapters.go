package kernel

import (
	"context"
	"encoding/json"

	"github.com/gavmor/podpedia/internal/types"
	"github.com/gavmor/wasm-microkernel/host"
	"github.com/spf13/afero"
)

// PodpediaKernel is a domain-specific wrapper around the generic host.Kernel.
// It holds configuration for specific AI services and provides high-level
// methods for pipeline stages.
type PodpediaKernel struct {
	*host.Kernel
	OllamaURL     string
	OllamaModel   string
	TranscribeURL string
}

func NewPodpediaKernel(ollamaURL, ollamaModel, transcribeURL string) *PodpediaKernel {
	// Policy: Allow everything for now to ensure stability, can be tightened later.
	k := host.NewKernel(host.Config{
		AllowedHosts: []string{"*"},
		AllowedPaths: map[string]string{"/": "/", ".": "."},
	})

	return &PodpediaKernel{
		Kernel:        k,
		OllamaURL:     ollamaURL,
		OllamaModel:   ollamaModel,
		TranscribeURL: transcribeURL,
	}
}

// Transcriber adapter
type transcriber struct {
	pk *PodpediaKernel
}

func NewTranscriber(pk *PodpediaKernel) *transcriber {
	return &transcriber{pk: pk}
}

func (t *transcriber) Transcribe(ctx context.Context, audioURL string, prompt string) (string, error) {
	req := struct {
		AudioURL string `json:"audio_url"`
		Prompt   string `json:"prompt"`
	}{
		AudioURL: audioURL,
		Prompt:   prompt,
	}

	config := map[string]string{
		"transcribe-url": t.pk.TranscribeURL,
	}

	res, err := t.pk.Call(ctx, "transcriber", req, config)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

// Extractor adapter
type extractor struct {
	pk *PodpediaKernel
}

func NewExtractor(pk *PodpediaKernel) *extractor {
	return &extractor{pk: pk}
}

// extractRequest is the wire contract with plugins/extractor; it has to stay
// structurally identical to that module's own extractRequest.
//
// Scheme is json.RawMessage rather than []byte deliberately. encoding/json
// renders a []byte as a base64 *string*, so a scheme of `{}` left the host as
// `"scheme":"e30="` and the plugin — which decodes into json.RawMessage and
// pastes the result straight into the LLM prompt — told the model its output
// schema was the six characters e30=. internal/protocol.ExtractRequest has
// always declared the field correctly; this struct had drifted from it.
type extractRequest struct {
	Episode types.Episode   `json:"episode"`
	Scheme  json.RawMessage `json:"scheme"`
}

func newExtractRequest(ep *types.Episode, scheme json.RawMessage) extractRequest {
	return extractRequest{Episode: *ep, Scheme: scheme}
}

func (e *extractor) ExtractEntities(ctx context.Context, ep *types.Episode, scheme json.RawMessage) ([]byte, error) {
	req := newExtractRequest(ep, scheme)

	config := map[string]string{
		"ollama-url":   e.pk.OllamaURL,
		"ollama-model": e.pk.OllamaModel,
	}

	return e.pk.Call(ctx, "extractor", req, config)
}

// Downloader adapter
type downloader struct {
	pk *PodpediaKernel
}

func NewDownloader(pk *PodpediaKernel) *downloader {
	return &downloader{pk: pk}
}

type downloaderWorkspaceKeyType int

const downloaderWorkspaceKey downloaderWorkspaceKeyType = iota

func (d *downloader) DownloadAudio(ctx context.Context, fs afero.Fs, url, dest string) error {
	req := struct {
		URL  string `json:"url"`
		Dest string `json:"dest"`
	}{
		URL:  url,
		Dest: dest,
	}

	// Propagate the workspace through the context
	ctx = context.WithValue(ctx, downloaderWorkspaceKey, fs)

	_, err := d.pk.Call(ctx, "downloader", req, nil)
	return err
}

// Store adapter
type store struct {
	pk *PodpediaKernel
}

func NewStore(pk *PodpediaKernel) *store {
	return &store{pk: pk}
}

type storeWorkspaceKeyType int

const storeWorkspaceKey storeWorkspaceKeyType = iota

func (s *store) SaveRawData(ctx context.Context, fs afero.Fs, outputDir string, ep *types.Episode) error {
	req := struct {
		OutputDir string        `json:"output_dir"`
		Episode   types.Episode `json:"episode"`
	}{
		OutputDir: outputDir,
		Episode:   *ep,
	}

	// Propagate the workspace through the context
	ctx = context.WithValue(ctx, storeWorkspaceKey, fs)

	_, err := s.pk.Call(ctx, "store", req, nil)
	return err
}

// SaveStructuredData hands the extractor's output to plugins/store.
//
// Entry is still []byte, and so still reaches the plugin base64-encoded as a
// JSON string — the same defect this file's extractRequest just lost, in the
// other direction. It is left alone on purpose: plugins/store writes that
// value verbatim to <slug>_<scheme>.json, so retyping it to json.RawMessage
// turns that file from an unparseable string into a real object and makes
// podpedia-app's PodcastSource.readEntryGraph start populating Document.Graph
// where it has always returned nil. That is a product change, not a wire fix,
// and it is the decision parked on podpedia-app card t_298a9c7e.
func (s *store) SaveStructuredData(ctx context.Context, fs afero.Fs, outputDir string, ep *types.Episode, entry []byte, schemeID string) error {
	req := struct {
		OutputDir string        `json:"output_dir"`
		Episode   types.Episode `json:"episode"`
		Entry     []byte        `json:"entry"`
		SchemeID  string        `json:"scheme_id"`
	}{
		OutputDir: outputDir,
		Episode:   *ep,
		Entry:     entry,
		SchemeID:  schemeID,
	}

	// Propagate the workspace through the context
	ctx = context.WithValue(ctx, storeWorkspaceKey, fs)

	_, err := s.pk.Call(ctx, "store", req, nil)
	return err
}
