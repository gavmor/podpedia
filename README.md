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
| `--output` | `-o` | `output` | Directory to write transcripts and entries |
| `--plugins` | `-p` | `dist/plugins` | Directory containing compiled `.wasm` plugins |
| `--limit` | `-n` | `0` (all) | Maximum number of episodes to process |
| `--ollama` | | `http://localhost:11434` | Ollama base URL for LLM inference |
| `--transcribe-url` | | *(none)* | ASR endpoint for transcription (Whisper.cpp, Deepgram, etc.) |

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
