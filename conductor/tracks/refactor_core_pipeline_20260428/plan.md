# Implementation Plan: Refactor Core Pipeline into a Reusable Library

## Phase 1: Setup and Workspace Abstraction [checkpoint: 77c1389]
- [x] Task: Update dependencies to include `spf13/afero`. 5b165a5
- [x] Task: Write tests for the `Workspace` interface abstraction using `afero.Fs` in `internal/pipeline`. dd00e9f
- [x] Task: Implement `Workspace` abstraction in `internal/pipeline/pipeline.go` by replacing direct `os` calls with `afero.Fs`. dd00e9f
- [x] Task: Update existing unit tests to use `afero.MemMapFs` where applicable. dd00e9f
- [x] Task: Conductor - User Manual Verification 'Phase 1: Setup and Workspace Abstraction' (Protocol in workflow.md)

## Phase 2: Context Propagation and Hard Abort [checkpoint: 6c65831]
- [x] Task: Write tests to ensure `context.Context` cancellation halts workers and HTTP downloads. 6c65831
- [x] Task: Update `Pipeline.Run` and internal workers to use `pond`'s `pool.GroupContext(ctx)`. 6c65831
- [x] Task: Plumb `context.Context` down into `http.NewRequestWithContext` in the Downloader plugin. 6c65831
- [x] Task: Plumb `context.Context` into WASM `CallWithContext` executions. 6c65831
- [~] Task: Conductor - User Manual Verification 'Phase 2: Context Propagation and Hard Abort' (Protocol in workflow.md)

## Phase 3: Error Aggregation and Progress Hooks [checkpoint: 003641e]
- [x] Task: Define and export structured error types (e.g., `ErrTranscriptionTimeout`) in `pkg/podpedia/errors.go`. 003641e
- [x] Task: Write tests for `errors.Join` aggregation within the worker pool. 003641e
- [x] Task: Implement error aggregation using `errors.Join` in `Pipeline.Run`. 003641e
- [x] Task: Write tests for progress hooks (`OnEpisodeComplete`). 003641e
- [x] Task: Implement progress hook callbacks in the pipeline configuration and trigger them from the worker pool. 003641e
- [~] Task: Conductor - User Manual Verification 'Phase 3: Error Aggregation and Progress Hooks' (Protocol in workflow.md)

## Phase 4: Public API and CLI Refactor [checkpoint: 5da6d77]
- [x] Task: Write tests for the new `pkg/podpedia` functional options API. 5da6d77
- [x] Task: Create `pkg/podpedia/app.go` and define the `App` structure, `Option` pattern, and `NewApp` function. 5da6d77
- [x] Task: Refactor `cmd/run.go` to construct and execute the pipeline via the new `pkg/podpedia` API. 5da6d77
- [~] Task: Conductor - User Manual Verification 'Phase 4: Public API and CLI Refactor' (Protocol in workflow.md)
