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
- Integrate **\`github.com/cavaliergopher/grab/v3\`** for robust concurrent audio acquisition.
- Leverage the library's built-in support for \`Range\` headers, chunk merging, and CDN quirk handling.
- Integrate with the existing \`internal/pipeline\` to replace the mock/sequential download placeholder.

## Acceptance Criteria
- Successfully downloads 100MB+ audio files significantly faster than sequential methods.
- The resulting merged file is valid and playable (verified via basic file check).
- Successfully integrates the external FOSS library into the pipeline.
- No progress UI (None).


## Out of Scope
- Visual progress bars.
- Persistent resume-tracking between application restarts.
- Complex retry strategies with exponential backoff.
