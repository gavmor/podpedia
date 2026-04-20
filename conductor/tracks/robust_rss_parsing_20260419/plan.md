# Implementation Plan: Implement Robust RSS Parsing with gofeed

## Phase 1: Dependency & Schema Updates
- [x] Task: Add \`github.com/mmcdole/gofeed\` to \`go.mod\` 484f192
- [x] Task: Extend \`internal/types\` with new metadata fields 484f192
    - [x] Update \`Podcast\` struct to include Author and Categories c9a2cb7
    - [x] Update \`Episode\` struct to include Author, Duration, and Explicit flag c9a2cb7
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Dependency & Schema Updates' (Protocol in workflow.md)

## Phase 2: Gofeed Integration
- [ ] Task: Implement `gofeed` based parsing in `internal/pipeline`
    - [ ] Write Tests: Create failing tests for extended metadata extraction and robust parsing of problematic feeds
    - [ ] Implement: Replace manual XML logic with `gofeed.Parser` and verify against tests
- [ ] Task: Map iTunes and Dublin Core extensions to internal types
- [ ] Task: Clean up and remove legacy `parseRSS` and `fetchFeedContent` functions from `pipeline.go`
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Gofeed Integration' (Protocol in workflow.md)
