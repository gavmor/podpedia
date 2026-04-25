package kernel

import (
	"context"

	"github.com/gavmor/podpedia/internal/types"
	"github.com/gavmor/wasm-microkernel/host"
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

func (t *transcriber) Transcribe(ep types.Episode) (string, error) {
	req := struct {
		AudioURL string `json:"audio_url"`
		Title    string `json:"title"`
		Notes    string `json:"notes"`
	}{
		AudioURL: ep.AudioURL,
		Title:    ep.Title,
		Notes:    ep.Description,
	}

	config := map[string]string{
		"transcribe-url": t.pk.TranscribeURL,
	}

	res, err := t.pk.Call(context.Background(), "transcriber", req, config)
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

func (e *extractor) ExtractEntities(ep types.Episode, scheme []byte) ([]byte, error) {
	req := struct {
		Episode types.Episode `json:"episode"`
		Scheme  []byte        `json:"scheme"`
	}{
		Episode: ep,
		Scheme:  scheme,
	}

	config := map[string]string{
		"ollama-url":   e.pk.OllamaURL,
		"ollama-model": e.pk.OllamaModel,
	}

	return e.pk.Call(context.Background(), "extractor", req, config)
}

// Downloader adapter
type downloader struct {
	pk *PodpediaKernel
}

func NewDownloader(pk *PodpediaKernel) *downloader {
	return &downloader{pk: pk}
}

func (d *downloader) DownloadAudio(url, dest string) error {
	req := struct {
		URL  string `json:"url"`
		Dest string `json:"dest"`
	}{
		URL:  url,
		Dest: dest,
	}

	_, err := d.pk.Call(context.Background(), "downloader", req, nil)
	return err
}

// Store adapter
type store struct {
	pk *PodpediaKernel
}

func NewStore(pk *PodpediaKernel) *store {
	return &store{pk: pk}
}

func (s *store) SaveRawData(outputDir string, ep types.Episode) error {
	req := struct {
		OutputDir string        `json:"output_dir"`
		Episode   types.Episode `json:"episode"`
	}{
		OutputDir: outputDir,
		Episode:   ep,
	}

	_, err := s.pk.Call(context.Background(), "store", req, nil)
	return err
}

func (s *store) SaveStructuredData(outputDir string, ep types.Episode, entry []byte, schemeID string) error {
	req := struct {
		OutputDir string        `json:"output_dir"`
		Episode   types.Episode `json:"episode"`
		Entry     []byte        `json:"entry"`
		SchemeID  string        `json:"scheme_id"`
	}{
		OutputDir: outputDir,
		Episode:   ep,
		Entry:     entry,
		SchemeID:  schemeID,
	}

	_, err := s.pk.Call(context.Background(), "store", req, nil)
	return err
}
