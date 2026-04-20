// Package protocol defines the JSON request/response types shared between
// the host kernel and every WASM plugin. Each plugin receives one Request
// and returns one Response, both encoded as JSON over fat pointers.
package protocol

import "github.com/gavmor/podpedia/internal/types"

// ── RSS plugin ────────────────────────────────────────────────────────────────

// RSSRequest carries raw RSS/Atom XML for the rss plugin to parse.
// The host fetches the feed URL and passes the body here so the plugin
// stays network-free.
type RSSRequest struct {
	XML string `json:"xml"`
}

type RSSResponse struct {
	Podcast  types.Podcast   `json:"podcast"`
	Episodes []types.Episode `json:"episodes"`
}

// ── Downloader plugin ─────────────────────────────────────────────────────────

type DownloadRequest struct {
	URL  string `json:"url"`
	Dest string `json:"dest"`
}

type DownloadResponse struct {
	Path string `json:"path"`
}

// ── Transcriber plugin ────────────────────────────────────────────────────────

// TranscribeRequest tells the transcriber plugin where the audio lives and
// which HTTP endpoint to POST to for transcription. The plugin owns the
// API-specific request format; the host only provides http_post.
type TranscribeRequest struct {
	AudioPath     string `json:"audio_path"`
	AudioURL      string `json:"audio_url"`
	TranscribeURL string `json:"transcribe_url"` // e.g. Deepgram, AssemblyAI, Whisper.cpp
}

type TranscribeResponse struct {
	Transcript string `json:"transcript"`
}

// ── Extractor plugin ──────────────────────────────────────────────────────────

// ExtractRequest carries the episode and the Ollama base URL. The plugin
// builds and sends the LLM request itself; the host only provides http_post.
type ExtractRequest struct {
	Episode   types.Episode `json:"episode"`
	OllamaURL string        `json:"ollama_url"`
}

type ExtractResponse struct {
	Entry types.EncyclopediaEntry `json:"entry"`
}

// ── Store plugin ──────────────────────────────────────────────────────────────

type StoreRawRequest struct {
	OutputDir string        `json:"output_dir"`
	Episode   types.Episode `json:"episode"`
}

type StoreStructuredRequest struct {
	OutputDir string                  `json:"output_dir"`
	Entry     types.EncyclopediaEntry `json:"entry"`
}

type StoreResponse struct {
	Path string `json:"path"`
}
