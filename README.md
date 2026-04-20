# Podpedia

An automated pipeline that ingests a podcast RSS feed, processes episodes concurrently, and uses local LLMs to extract structured data.

## Features
- Robust RSS ingestion via `gofeed`.
- Hardware-aware concurrent processing via `alitto/pond`.
- Robust audio downloading via `cavaliergopher/grab`.
- BDD-style testing with `Ginkgo` and `Gomega`.
- Structured JSON logging with `lager`.

## Usage
Run the pipeline against a podcast RSS feed:
```bash
./podpedia run --url <RSS_FEED_URL> --output <OUTPUT_DIR>
```
