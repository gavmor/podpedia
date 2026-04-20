package kernel

import (
	"encoding/json"
	"fmt"

	"github.com/gavmor/podpedia/internal/protocol"
	"github.com/gavmor/podpedia/internal/types"
)

// Each adapter implements one of the pipeline's consumer-driven interfaces
// by delegating to a named WASM plugin via Kernel.Call. pipeline.go is unchanged.

// ── Transcriber ───────────────────────────────────────────────────────────────

type WasmTranscriber struct{ k *Kernel }

func NewTranscriber(k *Kernel) *WasmTranscriber { return &WasmTranscriber{k} }

func (t *WasmTranscriber) Transcribe(audioURL string) (string, error) {
	raw, err := t.k.Call("transcriber", protocol.TranscribeRequest{AudioURL: audioURL, TranscribeURL: t.k.transcribeURL})
	if err != nil {
		return "", err
	}
	payload, err := unwrapResult(raw)
	if err != nil {
		return "", err
	}
	var resp protocol.TranscribeResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return "", fmt.Errorf("transcriber response: %w", err)
	}
	return resp.Transcript, nil
}

// ── Extractor ─────────────────────────────────────────────────────────────────

type WasmExtractor struct{ k *Kernel }

func NewExtractor(k *Kernel) *WasmExtractor { return &WasmExtractor{k} }

func (e *WasmExtractor) ExtractEntities(ep types.Episode) (types.EncyclopediaEntry, error) {
	raw, err := e.k.Call("extractor", protocol.ExtractRequest{Episode: ep, OllamaURL: e.k.ollamaURL})
	if err != nil {
		return types.EncyclopediaEntry{}, err
	}
	payload, err := unwrapResult(raw)
	if err != nil {
		return types.EncyclopediaEntry{}, err
	}
	var resp protocol.ExtractResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		return types.EncyclopediaEntry{}, fmt.Errorf("extractor response: %w", err)
	}
	return resp.Entry, nil
}

// unwrapResult decodes the {"ok":"..."} / {"err":"..."} wire envelope written
// by guest-bindings/plugin_world. If neither key is present it returns raw
// unchanged so callers are backwards-compatible with old plugins.
func unwrapResult(raw []byte) ([]byte, error) {
	var wire struct {
		Ok  *string `json:"ok"`
		Err *string `json:"err"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return raw, nil
	}
	if wire.Err != nil {
		return nil, fmt.Errorf("plugin error: %s", *wire.Err)
	}
	if wire.Ok != nil {
		return []byte(*wire.Ok), nil
	}
	return raw, nil
}

// ── AudioDownloader ───────────────────────────────────────────────────────────

type WasmDownloader struct{ k *Kernel }

func NewDownloader(k *Kernel) *WasmDownloader { return &WasmDownloader{k} }

func (d *WasmDownloader) DownloadAudio(url, dest string) error {
	_, err := d.k.Call("downloader", protocol.DownloadRequest{URL: url, Dest: dest})
	return err
}

// ── Store ─────────────────────────────────────────────────────────────────────

type WasmStore struct{ k *Kernel }

func NewStore(k *Kernel) *WasmStore { return &WasmStore{k} }

func (s *WasmStore) SaveRawData(outputDir string, ep types.Episode) error {
	_, err := s.k.Call("store", protocol.StoreRawRequest{OutputDir: outputDir, Episode: ep})
	return err
}

func (s *WasmStore) SaveStructuredData(outputDir string, entry types.EncyclopediaEntry) error {
	_, err := s.k.Call("store", protocol.StoreStructuredRequest{OutputDir: outputDir, Entry: entry})
	return err
}
