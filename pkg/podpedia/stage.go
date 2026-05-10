package podpedia

import (
	"context"
	"fmt"
	"sync"

	"github.com/alitto/pond"
)

// Source generates the initial Document from an external input
// (RSS feed, text upload, audio URL, etc.).
type Source interface {
	// Ingest produces one or more Documents from the source.
	// Returns a slice because some sources (e.g., RSS feeds) produce
	// multiple documents (one per episode).
	Ingest(ctx context.Context) ([]*Document, error)
}

// Transform mutates a Document in place — for example, running LLM
// extraction to populate Document.Graph, or summarizing Content.
type Transform interface {
	// Process applies the transformation to a single document.
	Process(ctx context.Context, doc *Document) error
}

// Sink writes the Document to an output destination — Graph JSON,
// Markdown file, GCS bucket, Firestore, HTTP response, etc.
type Sink interface {
	// Emit sends the document to the sink's destination.
	Emit(ctx context.Context, doc *Document) error
}

// GenericPipeline connects a Source to one or more Transforms and Sinks.
// It is a simple sequential orchestrator: Ingest → Transform... → Sink...
type GenericPipeline struct {
	source     Source
	transforms []Transform
	sinks      []Sink
}

// NewPipeline creates a pipeline with the given stages.
func NewPipeline(source Source, transforms []Transform, sinks []Sink) *GenericPipeline {
	return &GenericPipeline{
		source:     source,
		transforms: transforms,
		sinks:      sinks,
	}
}

// Execute runs the full pipeline for every document produced by the Source.
// Documents are processed sequentially to preserve ordering guarantees.
// For concurrent processing, use ExecuteConcurrent.
func (p *GenericPipeline) Execute(ctx context.Context) error {
	docs, err := p.source.Ingest(ctx)
	if err != nil {
		return fmt.Errorf("source ingest: %w", err)
	}

	for i, doc := range docs {
		if err := p.processDoc(ctx, doc); err != nil {
			return fmt.Errorf("transform on doc %d (%s): %w", i, doc.ID, err)
		}
	}

	return nil
}

// ExecuteConcurrent fans out document processing across maxConcurrency
// goroutines. Transforms and sinks are applied to each document in its own
// goroutine. Errors from individual documents are collected; the first
// error aborts remaining work.
//
// maxConcurrency: maximum simultaneous document processors (0 → sequential).
func (p *GenericPipeline) ExecuteConcurrent(ctx context.Context, maxConcurrency int) error {
	if maxConcurrency <= 1 {
		return p.Execute(ctx)
	}

	docs, err := p.source.Ingest(ctx)
	if err != nil {
		return fmt.Errorf("source ingest: %w", err)
	}

	if len(docs) <= 1 {
		// Single document — no benefit from fan-out
		for i, doc := range docs {
			if err := p.processDoc(ctx, doc); err != nil {
				return fmt.Errorf("transform on doc %d (%s): %w", i, doc.ID, err)
			}
		}
		return nil
	}

	// Use pond for bounded concurrency
	pool := pond.New(maxConcurrency, len(docs), pond.MinWorkers(maxConcurrency))
	defer pool.StopAndWait()

	var (
		mu       sync.Mutex
		firstErr error
	)

	for i, doc := range docs {
		idx := i
		d := doc
		pool.Submit(func() {
			mu.Lock()
			if firstErr != nil {
				mu.Unlock()
				return
			}
			mu.Unlock()

			if err := p.processDoc(ctx, d); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("doc %d (%s): %w", idx, d.ID, err)
				}
				mu.Unlock()
			}
		})
	}

	pool.StopAndWait()

	if firstErr != nil {
		return firstErr
	}
	return nil
}

// processDoc applies all transforms and sinks to a single document.
func (p *GenericPipeline) processDoc(ctx context.Context, doc *Document) error {
	for _, t := range p.transforms {
		if err := t.Process(ctx, doc); err != nil {
			return err
		}
	}

	for _, s := range p.sinks {
		if err := s.Emit(ctx, doc); err != nil {
			return err
		}
	}

	return nil
}
