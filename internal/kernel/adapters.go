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

// AllowedPaths is the host→guest filesystem mount policy every plugin runs
// under. Mounting all of "/" read-write into every WASM stage is deliberately
// permissive — tightening it is a separate hardening question — but it has to
// be unambiguous.
//
// It must stay unambiguous because the Extism SDK applies this map by ranging
// over it (extism/go-sdk plugin.go, `for host, guest := range
// p.manifest.AllowedPaths`) and Go randomises map iteration order. Two
// entries resolving to the same guest path are therefore not redundant: the
// winner is drawn afresh on every Kernel.Call.
//
// This used to read {"/": "/", ".": "."}. wazero normalises the guest path
// "." to "/", so both entries claimed the guest root, and a plugin writing
// /tmp/podpedia_workspace/ep_raw.txt landed there or at
// <cwd>/tmp/podpedia_workspace/ep_raw.txt at random, per call — reporting
// success either way, so the only symptom a caller saw was output that went
// missing about half the time.
//
// A cwd-relative mount, if one is ever wanted, needs a guest path no other
// entry claims (".": "/work", say) — never a second route to the guest root.
func AllowedPaths() map[string]string {
	return map[string]string{"/": "/"}
}

func NewPodpediaKernel(ollamaURL, ollamaModel, transcribeURL string) *PodpediaKernel {
	// Policy: Allow everything for now to ensure stability, can be tightened later.
	k := host.NewKernel(host.Config{
		AllowedHosts: []string{"*"},
		AllowedPaths: AllowedPaths(),
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

// SaveStructuredData hands the extractor's output to plugins/store, which
// writes it verbatim to <slug>_<schemeID>.json.
//
// Entry is json.RawMessage rather than []byte for the same reason
// extractRequest.Scheme is: encoding/json renders a []byte as a base64
// *string*, so an extraction of `{"guests":[]}` left the host as
// `"entry":"eyJndWVzdHMiOltdfQ=="` and the plugin — which has always decoded
// into json.RawMessage, and json.MarshalIndents whatever it finds — wrote
// that quoted base64 blob out as the episode's structured artifact. Every
// consumer of the engine got a JSON string where the object should be.
//
// internal/protocol.StoreStructuredRequest has always declared the field
// correctly; this struct had drifted from it.
func (s *store) SaveStructuredData(ctx context.Context, fs afero.Fs, outputDir string, ep *types.Episode, entry json.RawMessage, schemeID string) error {
	// A zero-length RawMessage is not valid JSON and would fail the request
	// marshal outright; nil renders as `null`, which at least round-trips.
	if len(entry) == 0 {
		entry = nil
	}

	req := struct {
		OutputDir string          `json:"output_dir"`
		Episode   types.Episode   `json:"episode"`
		Entry     json.RawMessage `json:"entry"`
		SchemeID  string          `json:"scheme_id"`
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
