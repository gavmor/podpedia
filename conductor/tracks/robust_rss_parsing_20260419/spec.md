# Track Specification: Implement Robust RSS Parsing with gofeed

## Overview
Replace the current manual XML parsing with the `gofeed` library to improve robustness and metadata extraction depth. This will allow the pipeline to handle various RSS standards, extensions (iTunes, Dublin Core), and malformed XML.

## Functional Requirements
- Integrate `github.com/mmcdole/gofeed` for RSS ingestion.
- Extract high-value metadata including:
    - Podcast Title, Description, and Author.
    - Episode Title, GUID, Media URL, and PubDate.
    - iTunes Extensions: Summary, Author, Duration.
    - Dublin Core: Creator.
    - Categories and Explicit content flags.
- Support robust parsing of feeds with idiosyncratic formatting or broken markup.
- Ensure the extraction logic handles iTunes and Dublin Core namespaces seamlessly.

## Technical Approach
- Replace `fetchFeedContent` and `parseRSS` with `gofeed.Parser`.
- Update `internal/types` to include the new metadata fields (Author, Category, etc.).
- Map `gofeed.Feed` and `gofeed.Item` fields to our internal `Podcast` and `Episode` structs.

## Acceptance Criteria
- Successfully parses feeds that were previously problematic or used complex extensions.
- All prioritized metadata fields are correctly populated in the `types.Episode` and `types.Podcast` structs.
- Unit tests verify the extraction of iTunes and Dublin Core extensions.
- The existing manual parser logic is fully removed from `internal/pipeline`.

## Out of Scope
- Implementing the "skip transcription" logic based on summary density.
- Concurrent audio downloading.
