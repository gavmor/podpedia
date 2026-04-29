# Specification: Refactor Core Pipeline into a Reusable Library

## 1. Overview
The core domain logic of Podpedia (WASM plugin execution, LLM transcription, entity extraction) will be refactored to decouple it from the CLI initialization. This will expose Podpedia as a reusable Go library (`pkg/podpedia`), allowing it to be imported and executed in serverless/cloud-native environments (like Google Cloud Run or AWS ECS).

## 2. Functional Requirements
- **Decoupled Initialization:** Create a public API in `pkg/podpedia/app.go` to construct and configure the `PodpediaKernel`, WASM plugin loading, and Pipeline.
- **Context Support:** Update `Pipeline.Run(ctx context.Context, rssURL string, ...)` to accept a context. The workers and HTTP downloads must respect the context and perform a hard abort immediately upon cancellation.
- **Workspace Interface:** Abstract intermediate file I/O operations (transcripts, audio, etc.) behind a new `Workspace` interface.
- **Storage Implementations:** Initially provide `Workspace` implementations for the local filesystem and an in-memory byte buffer storage.
- **Functional Options:** Implement a structured functional options pattern (`podpedia.WithOption`) for pipeline configuration.
- **Custom Errors:** Define and export structured error types (e.g., `podpedia.ErrTranscriptionTimeout`, `podpedia.ErrInvalidFeed`) to enable programmatic error handling via `errors.Is`.
- **Progress Hooks:** Add support for callbacks/hooks (e.g., `OnEpisodeComplete`) to allow calling applications to monitor and report pipeline progress.

## 3. Non-Functional Requirements
- **Performance:** Context cancellation must immediately free up system resources without leaking goroutines.
- **Compatibility:** The existing CLI commands must be refactored to use the new `pkg/podpedia` library without changing their external behavior or flags.

## 4. Acceptance Criteria
- `pkg/podpedia` package is exported and well-documented for public use.
- The pipeline can be fully instantiated without using `cmd/run.go` or `os` dependencies.
- Calling `cancel()` on the provided context immediately stops all active downloads, LLM inferences, and WASM executions.
- All file operations are routed through the `Workspace` interface (using `afero.Fs`).
- Unit tests verify the behavior of the new `Workspace` interface implementations (local and memory via `afero`).
- A calling application can successfully inject a custom callback and receive progress updates.

## 5. Out of Scope
- Implementing Cloud Storage (S3/GCS) `Workspace` integrations (to be handled in a future track, though `afero` supports this).
- Adding new extraction features or changing the core LLM prompts.

## 6. Implementation Recommendations (Tools & Patterns)
To minimize the refactoring burden and align with modern Go idioms, we recommend the following patterns and community libraries for the implementation:

#### A. The `Workspace` Interface (Recommended Library: `spf13/afero`)
Instead of writing a custom `Workspace` interface and building an in-memory byte-buffer implementation from scratch, we highly recommend utilizing **`github.com/spf13/afero`**.
* **Why it succeeds:** `afero` is the gold standard for filesystem abstraction in Go. By updating `pipeline.go` to accept an `afero.Fs` interface, the maintainers instantly get robust, battle-tested implementations of `afero.OsFs` (for CLI backward compatibility) and `afero.MemMapFs` (for our serverless/testing requirements) with zero custom logic required.
* **Future-proofing:** When we eventually need Cloud Storage, `afero` already has community adapters for S3 and GCS, meaning the core Podpedia library won't need to change again.

#### B. Context-Aware Worker Pools (Pattern: `pond` Task Groups)
The current pipeline uses `github.com/alitto/pond` for the worker pool. To implement the "Hard Abort" requirement efficiently without leaking goroutines:
* **Recommendation:** Switch from standard `pool.Submit()` to **`pool.GroupContext(ctx)`**. 
* **Why it succeeds:** This native `pond` pattern automatically handles context propagation. If the parent context is cancelled, the group stops accepting new tasks. The implementers just need to ensure that the context passed to the worker is chained down into the `http.NewRequestWithContext` calls in the Downloader and the LLM inference calls.

#### C. Context Propagation into WASM
Since the core domain logic relies heavily on WASM plugins, a hard abort must be able to interrupt a running WASM execution (especially during heavy extraction tasks).
* **Recommendation:** Ensure the WASM runtime (e.g., if using `tetratelabs/wazero`) is passing the `context.Context` directly into the `CallWithContext` methods. If a WASM plugin gets stuck in an infinite loop, the context cancellation will forcefully terminate the WASM instance and free the CPU thread.

#### D. The Functional Options Pattern
For the new `Config` initialization, utilize the standard Go functional options pattern to keep the API surface clean while allowing deep customization for serverless environments.
```go
// Example of the recommended pattern for pkg/podpedia
type Option func(*App)

func WithWorkspace(fs afero.Fs) Option {
    return func(a *App) {
        a.workspace = fs
    }
}

// Allows simple CLI init: app := podpedia.NewApp()
// Or complex Cloud Run init: app := podpedia.NewApp(podpedia.WithWorkspace(memFs))
func NewApp(opts ...Option) (*App, error)
```

#### E. Error Aggregation (Pattern: `errors.Join`)
In a highly concurrent, idempotent pipeline, one failed episode (e.g., a 404 audio file) should not crash the entire pipeline run, but the calling server still needs to know about the failures.
* **Recommendation:** Use Go 1.20's native **`errors.Join()`** (or `hashicorp/go-multierror`) within the worker pool to aggregate episode-specific failures. This allows the `Run(ctx)` method to return a single multierror that the Cloud Run handler can parse to determine if a partial success occurred.
