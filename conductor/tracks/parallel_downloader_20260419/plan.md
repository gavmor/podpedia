# Implementation Plan: Implement Concurrent Audio Acquisition (Parallel Downloader)

## Phase 1: Size Discovery & Content Validation
- [x] Task: Implement \`GetAudioMetadata\` in \`internal/pipeline\` b3c847b
    - [x] Write Tests: Create failing tests for fetching \`Content-Length\` and verifying \`Accept-Ranges: bytes\` support via HTTP HEAD requests. b3c847b
    - [x] Implement: Add logic to determine file size and server capability. b3c847b

- [ ] Task: Conductor - User Manual Verification 'Phase 1: Size Discovery & Content Validation' (Protocol in workflow.md)

## Phase 2: Concurrent Segment Fetching (In-Memory)
- [ ] Task: Implement core `DownloadSegment` logic
    - [ ] Write Tests: Define tests for fetching a specific byte range into a buffer.
    - [ ] Implement: Use `Range` headers to download a chunk.
- [ ] Task: Implement `ParallelDownloader` orchestrator
    - [ ] Write Tests: Verify that multiple segments are downloaded concurrently and held in memory.
    - [ ] Implement: Manage goroutines and `sync.WaitGroup` to coordinate 5-10 connections.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Concurrent Segment Fetching (In-Memory)' (Protocol in workflow.md)

## Phase 3: Reassembly & Pipeline Integration
- [ ] Task: Implement reassembly and file persistence
    - [ ] Write Tests: Verify that multiple buffers are correctly ordered and written to a single file.
    - [ ] Implement: Concatenate segments and save the final audio file.
- [ ] Task: Integrate with main pipeline `Run` loop
    - [ ] Write Tests: Ensure the pipeline uses the new parallel downloader for episodes.
    - [ ] Implement: Replace the mock/placeholder download logic.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Reassembly & Pipeline Integration' (Protocol in workflow.md)
