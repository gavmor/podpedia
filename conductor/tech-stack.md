# Technology Stack

## Core
- **Language:** Go (1.25.0)
- **CLI Framework:** [spf13/cobra](https://github.com/spf13/cobra) - Industry standard for Go CLI applications.

## Data & Storage
- **RSS Parser:** [gofeed](https://github.com/mmcdole/gofeed) - Universal feed parser for RSS, Atom, and JSON feeds with support for iTunes and Dublin Core extensions.
- **File Downloader:** [got](https://github.com/melbahja/got) - High-speed concurrent and resumable downloader using HTTP Range headers.
- **Primary Storage:** Markdown files - Portability and ease of search/editing.
- **Data Exchange:** JSON - Standard format for structured exports.

## Machine Learning / AI
- **LLM Inference:** [Ollama](https://ollama.ai/) - Primary local inference engine.
- **Structured Output:** [instructor-go](https://github.com/567-labs/instructor-go) - Enforces JSON Schema adherence at the LLM level for deterministic extraction.
- **Interface:** OpenAI-compatible API - Portability and support for various backends.

## Infrastructure & Concurrency
- **Worker Pool:** [alitto/pond](https://github.com/alitto/pond) - High-performance goroutine worker pool with backpressure and dynamic scaling.
- **Concurrency Model:** Hardware-aware concurrent processing using `pond` pools. Workers scale dynamically based on available CPU cores (`runtime.NumCPU()`).
