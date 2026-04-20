# Implementation Plan: Implement Concurrent Audio Acquisition (Parallel Downloader)

## Phase 1: Size Discovery & Content Validation [checkpoint: d142ff5]
- [x] Task: Implement `GetAudioMetadata` in `internal/pipeline` b3c847b
    - [x] Write Tests: Create failing tests for fetching `Content-Length` and verifying `Accept-Ranges: bytes` support via HTTP HEAD requests. b3c847b
    - [x] Implement: Add logic to determine file size and server capability. b3c847b
- [x] Task: Conductor - User Manual Verification 'Phase 1: Size Discovery & Content Validation' (Protocol in workflow.md)

## Phase 2: Got Library Integration
- [x] Task: Add \`github.com/melbahja/got\` to \`go.mod\` 40658eb
- [x] Task: Implement \`DownloadAudio\` using \`got.Download\` 8976208
    - [x] Write Tests: Verify the integration by downloading a small test file using the library. 8976208
    - [x] Implement: Wrap the library's download logic in \`internal/pipeline/downloader.go\`. 8976208

- [ ] Task: Integrate with main pipeline `Run` loop
    - [ ] Write Tests: Ensure the pipeline uses the new parallel downloader for episodes.
    - [ ] Implement: Replace the mock/placeholder download logic in `processEpisode`.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Got Library Integration' (Protocol in workflow.md)
