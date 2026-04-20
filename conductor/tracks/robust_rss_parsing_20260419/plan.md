# Implementation Plan: Implement Robust RSS Parsing with gofeed

## Phase 1: Dependency & Schema Updates [checkpoint: fe59b93]
- [x] Task: Add \`github.com/mmcdole/gofeed\` to \`go.mod\` 484f192
- [x] Task: Extend \`internal/types\` with new metadata fields 484f192
    - [x] Update \`Podcast\` struct to include Author and Categories c9a2cb7
    - [x] Update \`Episode\` struct to include Author, Duration, and Explicit flag c9a2cb7
- [x] Task: Conductor - User Manual Verification 'Phase 1: Dependency & Schema Updates' (Protocol in workflow.md)

## Phase 2: Gofeed Integration
- [x] Task: Implement \`gofeed\` based parsing in \`internal/pipeline\` 474c08d
    - [x] Write Tests: Create failing tests for extended metadata extraction and robust parsing of problematic feeds 474c08d
    - [x] Implement: Replace manual XML logic with \`gofeed.Parser\` and verify against tests 474c08d
- [x] Task: Map iTunes and Dublin Core extensions to internal types 5559643
- [x] Task: Clean up and remove legacy \`parseRSS\` and \`fetchFeedContent\` functions from \`pipeline.go\` 474c08d
- [x] Task: Conductor - User Manual Verification 'Phase 2: Gofeed Integration' (Protocol in workflow.md) [checkpoint: ea82dac]
