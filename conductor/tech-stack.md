# Technology Stack

## Core
- **Language:** Go (1.25.0)
- **CLI Framework:** [spf13/cobra](https://github.com/spf13/cobra) - Industry standard for Go CLI applications.

## Data & Storage
- **RSS Parser:** [gofeed](https://github.com/mmcdole/gofeed) - Universal feed parser for RSS, Atom, and JSON feeds with support for iTunes and Dublin Core extensions.
- **File Downloader:** [grab](https://github.com/cavaliercoder/grab) - Robust, production-ready concurrent downloader with retry support and CDN quirk handling.
- **Filesystem Abstraction:** [afero](https://github.com/spf13/afero) - Filesystem abstraction layer for Go, enabling in-memory storage and Cloud Storage adapters.
- **Primary Storage:** Markdown files - Portability and ease of search/editing.
- **Data Exchange:** JSON - Standard format for structured exports.

## Machine Learning / AI
- **LLM Inference:** [Ollama](https://ollama.ai/) - Primary local inference engine.
- **Structured Output:** [instructor-go](https://github.com/567-labs/instructor-go) - Enforces JSON Schema adherence at the LLM level for deterministic extraction.
- **Interface:** OpenAI-compatible API - Portability and support for various backends.

## Infrastructure & Concurrency
- **Worker Pool:** [alitto/pond](https://github.com/alitto/pond) - High-performance goroutine worker pool with backpressure and dynamic scaling.
- **Concurrency Model:** Hardware-aware concurrent processing using `pond` pools. Workers scale dynamically based on available CPU cores (`runtime.NumCPU()`).
- **Logging:** [lager](https://github.com/cloudfoundry/lager) - Structured JSON logging for high-scale observability.

## Testing & Quality
- **CI/CD:** [GitHub Actions](https://github.com/features/actions) - Automated testing, linting, and security scanning.
- **Release Orchestration:** [GoReleaser](https://goreleaser.com/) - Automated binary compilation and GitHub Release management.
- **BDD Framework:** [Ginkgo V2](https://github.com/onsi/ginkgo) - Behavior-Driven Development for expressive, documentation-like tests.
- **Matcher Library:** [Gomega](https://github.com/onsi/gomega) - Expressive, asynchronous matchers for Go.
- **Mocking:** [counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) - Generates type-safe fakes for consumer-driven interfaces.
- **Paradigm:** Strict black-box testing and consumer-defined interfaces.
