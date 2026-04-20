# Track Specification: Implement Concurrent Audio Acquisition (Parallel Downloader)

## Overview
Implement a high-performance audio downloader that utilizes HTTP Range headers to fetch segments of podcast audio files concurrently. This will replace any single-threaded sequential downloading and improve reliability for large media assets.

## Functional Requirements
- **Concurrent Segment Fetching:** Divide large audio files into segments and download them in parallel using 5-10 concurrent connections.
- **Size Discovery:** Use HTTP HEAD requests to determine the `Content-Length` of audio files before downloading.
- **In-Memory Buffering:** Store downloaded segments in memory buffers before merging them into the final file to minimize disk thrashing.
- **Resumability (Basic):** Implement internal tracking of segment completion to allow for resuming partially failed downloads within a single execution session.
- **Clean Merging:** Seamlessly concatenate in-memory buffers into a final audio file on disk.

## Technical Approach
- Use `net/http` with custom `Range` headers.
- Utilize Go channels and `sync.WaitGroup` to orchestrate parallel segment downloads.
- Manage memory buffers carefully to handle large audio files (typically 50MB - 200MB).
- Integrate with the existing `internal/pipeline` to replace the mock/sequential download placeholder.

## Acceptance Criteria
- Successfully downloads 100MB+ audio files significantly faster than sequential methods.
- The resulting merged file is bit-for-bit identical to a sequential download.
- Handles server errors (e.g., servers that don't support Range headers) by falling back to sequential download.
- No progress UI (None).

## Out of Scope
- Visual progress bars.
- Persistent resume-tracking between application restarts.
- Complex retry strategies with exponential backoff.
