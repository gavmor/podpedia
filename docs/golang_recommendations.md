# Go Codebase Recommendations Report

This report provides recommendations for improving the `podpedia` codebase based on Go best practices and the `golang-pro` standard.

## 1. Context Propagation (Critical)
**Observation:** Blocking operations across the pipeline do not accept `context.Context`.
**Impact:** Impossible to cancel long-running operations (e.g., large downloads or LLM extraction) or set timeouts for the entire pipeline.
**Recommendation:**
- Update all interfaces and methods to accept `context.Context` as the first argument.
- Propagate the context from `cmd/run.go` down to the lowest level (HTTP clients, file I/O).
- Example:
  ```go
  type Transcriber interface {
      Transcribe(ctx context.Context, audioURL string) (string, error)
  }
  ```

## 2. Idiomatic Concurrency & Error Handling
**Observation:** The pipeline uses `pond` for worker management but ignores errors from individual tasks in `processEpisode`.
**Impact:** The pipeline might "succeed" even if 90% of episodes failed to process.
**Recommendation:**
- Use an `errgroup.Group` or a similar pattern to collect and respond to errors from goroutines.
- If using `pond`, ensure there is a mechanism to report failures back to the `Run` method.
- Return wrapped errors using `fmt.Errorf("%w", err)` instead of just logging them in sub-functions if they are fatal to the task.

## 3. Modernize Logging
**Observation:** The project uses `code.cloudfoundry.org/lager/v3`.
**Impact:** Lager is less idiomatic than the standard library's `log/slog` introduced in Go 1.21.
**Recommendation:**
- Migrate to `log/slog` for structured logging. It is part of the standard library, highly performant, and has better ecosystem support.

## 4. Fix Deprecated Random Seeding
**Observation:** `rand.Seed(time.Now().UnixNano())` is used in multiple packages.
**Impact:** This is deprecated since Go 1.20 and is unnecessary in Go 1.22+ as the global random generator is automatically seeded.
**Recommendation:**
- Remove `rand.Seed` calls.
- For Go 1.22+, simply use `rand.Intn()`. For versions between 1.20 and 1.22, use `rand.New(rand.NewSource(time.Now().UnixNano()))` if a local generator is needed.

## 5. Testing Strategy
**Observation:** The project uses `ginkgo` and `gomega` for BDD-style testing.
**Impact:** BDD frameworks add complexity and external dependencies that are often unnecessary for Go projects.
**Recommendation:**
- Transition to standard Go table-driven tests using the `testing` package.
- Use `testify/assert` or `testify/require` if a more expressive assertion library is desired, as it is more common in the Go community than `gomega`.

## 6. Project Structure & Encapsulation
**Observation:** `ParseRSSWithGofeed` is located in `internal/pipeline/pipeline.go`.
**Impact:** Mixing parser logic with pipeline orchestration logic reduces maintainability.
**Recommendation:**
- Move RSS parsing logic to a dedicated `internal/ingestion` or `internal/parser` package.
- Move consumer-driven interfaces to a dedicated `internal/interfaces` package or keep them in the packages where they are consumed, but ensure they are minimal.

## 7. Configuration Management
**Observation:** Configuration (like `maxWorkers`) is hardcoded or derived from `runtime.NumCPU()`.
**Impact:** Difficult to tune the application for different environments (e.g., restricted CPU in CI or cloud environments).
**Recommendation:**
- Use functional options for `NewPipeline` to allow configuration of `maxWorkers`, timeouts, and other parameters.
- Load configuration from environment variables or a config file using a library like `spf13/viper` or a simple `envconfig` approach.

## 8. Explicit Error Handling in Storage
**Observation:** `SaveRawData` and `SaveStructuredData` in `Store` return errors, but `processEpisode` only logs them.
**Impact:** Silent failures in storage could lead to data loss.
**Recommendation:**
- Ensure storage errors are propagated back to the caller or handled with a retry mechanism if they are transient.
