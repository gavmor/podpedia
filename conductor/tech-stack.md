# Technology Stack

## Core
- **Language:** Go (1.25.0)
- **CLI Framework:** [spf13/cobra](https://github.com/spf13/cobra) - Industry standard for Go CLI applications.

## Data & Storage
- **RSS Parser:** [gofeed](https://github.com/mmcdole/gofeed) - Universal feed parser for RSS, Atom, and JSON feeds with support for iTunes and Dublin Core extensions.
- **Primary Storage:** Markdown files - Portability and ease of search/editing.
- **Data Exchange:** JSON - Standard format for structured exports.

## Machine Learning / AI
- **LLM Inference:** [Ollama](https://ollama.ai/) - Primary local inference engine.
- **Interface:** OpenAI-compatible API - Portability and support for various backends.

## Infrastructure & Concurrency
- **Concurrency Model:** Hardware-aware concurrent processing using Go native goroutines and channels. Workers scale dynamically based on available CPU cores (`runtime.NumCPU()`).
