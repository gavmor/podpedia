# Podpedia

An automated pipeline that ingests a podcast RSS feed, processes episodes concurrently, and uses local LLMs to extract structured data — building a searchable encyclopedia of guests, companies, and ideas.

## Features

- Robust RSS ingestion via `gofeed`
- Hardware-aware concurrent processing via `alitto/pond`
- Robust audio downloading via `cavaliergopher/grab`
- Sandboxed WASM plugin architecture — swap any pipeline stage without recompiling the host
- BDD-style testing with `Ginkgo` and `Gomega`
- Structured JSON logging with `lager`

## Prerequisites

| Dependency | Purpose | Required |
|---|---|---|
| [Go 1.26+](https://go.dev/dl/) | Build the binary and WASM plugins | Yes |
| [Ollama](https://ollama.com) | LLM entity extraction | Yes (for extraction) |
| Whisper-compatible ASR endpoint | Audio transcription | No (skipped if omitted) |

Pull a model into Ollama before running:

```bash
ollama pull qwen3.5:27b
```

## Build

```bash
# Build the CLI binary
go build -o podpedia .

# Compile all WASM pipeline plugins
make plugins
```

## Usage

```bash
./podpedia run --url <RSS_FEED_URL> [flags]
```

### Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--url` | `-u` | *(required)* | Podcast RSS feed URL |
| `--output` | `-o` | `output` | Directory to save processed data |
| `--plugins` | `-p` | `dist/plugins` | Directory containing compiled `.wasm` plugins |
| `--limit` | `-n` | `0` (all) | Maximum number of episodes to process |
| `--ollama` | | `http://localhost:11434` | Ollama base URL for LLM inference |
| `--ollama-model` | | `qwen2.5:0.5b` | Ollama model to use for extraction |
| `--output-scheme` | | *(none)* | Path to a JSON file defining the extraction schema |
| `--transcribe-url` | | *(none)* | ASR endpoint for transcription (Whisper.cpp, Deepgram, etc.) |

### Configurable Structured Extraction

Podpedia supports arbitrary extraction schemas via the `--output-scheme` flag. Provide a JSON file as a template, and the LLM will be instructed to populate it.

**Example: Ideology Analysis (`ideology.json`)**
```json
{
  "themes": ["List of core topics"],
  "political_leanings": "Description of leanings",
  "key_arguments": ["List of primary claims"]
}
```

**Example: Market Fit (`market_fit.json`)**
```json
{
  "total_addressable_market": "Estimated size",
  "pain_points": ["Customer problems identified"],
  "competitors_mentioned": ["Other companies in the space"],
  "investment_score": 0.85
}
```

Usage:
```bash
./podpedia run --url <FEED> --output-scheme ideology.json
```

### Examples

Process the latest episode of a feed:

```bash
./podpedia run --url https://changelog.com/gotime/feed --limit 1
```

Process 10 episodes with transcription and a custom output directory:

```bash
./podpedia run \
  --url https://changelog.com/gotime/feed \
  --limit 10 \
  --output ./data \
  --transcribe-url http://localhost:9000
```

Run the full back-catalogue against a remote Ollama instance:

```bash
./podpedia run \
  --url https://changelog.com/gotime/feed \
  --output ./data \
  --ollama http://my-ollama-host:11434
```

## Output

Each processed episode produces two files in `--output`:

| File | Contents |
|---|---|
| `<episode-id>_raw.txt` | Raw transcript (or episode title if transcription is skipped) |
| `<episode-id>_entry.json` | Structured encyclopedia entry with guests and companies |

Example entry:

```json
{
  "episode_id": "gotime-320",
  "guests": [
    { "name": "Alice Smith", "background": "Systems programmer", "ideology": "pragmatic" }
  ],
  "companies": [
    { "name": "Acme Corp", "business_model": "SaaS", "customers": "developers" }
  ]
}
```

## Pipeline Stages

```
RSS fetch → Download → Transcribe → Extract entities → Store
```

Each stage runs as a sandboxed WASM plugin loaded by the host kernel at startup. Transcription is skipped gracefully if `--transcribe-url` is not set. Extraction falls back to an empty entry if Ollama is unavailable.

The host kernel communicates with plugins using the [Extism](https://extism.org/) SDK (pure Go, via `wasm-microkernel` v0.7.0). The plugin guest SDK lives in [`github.com/gavmor/wasm-microkernel`](https://github.com/gavmor/wasm-microkernel).

## Plugin Architecture

All five pipeline stages are fully implemented as WebAssembly plugins compiled for `wasip1`. The `wasm-microkernel` provides a clean `guest.Register(...)` interface for plugins to export their business logic, and a secure `host.Engine` for the host application to run them with fine-grained HTTP access control.

## Development

```bash
# Run host unit tests (no WASM build required)
go test ./internal/...

# Run unit tests for all plugin logic (no WASM build required)
make test-plugins

# Run both
make ci

# Rebuild a single plugin
cd plugins/rss && GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o ../../dist/plugins/rss.wasm .

# Remove compiled plugins
make clean
```

# Request for Contributions

Podpedia is built on a modular, plugin-agnostic architecture. We are looking for contributors to help expand the pipeline with new capabilities:

### Enrichment & Validation
*   **Entity Linker:** Link extracted guests and companies to Wikipedia, LinkedIn, or Crunchbase.
*   **Fact Checker:** Verify "Key Arguments" against external sources using search-enabled LLMs.

### Content Generation
*   **Chapter Generator:** Automatically generate time-coded segments and table of contents.
*   **TL;DR Summarizer:** Generate executive summaries or formatted "Show Notes" in Markdown.
*   **Social Hook Generator:** Create threads/posts for Twitter and LinkedIn based on episode insights.

### Intelligence & Context
*   **Cross-Reference Linker:** Build a "knowledge web" by linking mentions across different episodes.
*   **Sentiment & Tone Analyzer:** Extract the "vibe" of discussions (e.g., Skeptical, Bullish, Academic).

### Infrastructure & Distribution
*   **Vector Embedding Store:** Generate embeddings for semantic search across the entire library.
*   **Translation Stage:** Translate transcripts to English before extraction for multilingual support.
*   **Notification Plugin:** Trigger Webhooks or Slack alerts when specific keywords are detected.

### Alternative Ingestion
*   **YouTube Ingester:** A plugin to pull audio from YouTube channels via `yt-dlp`.

If you're interested in building a plugin, check out the [WASM Microkernel Guest SDK](https://github.com/gavmor/wasm-microkernel).
