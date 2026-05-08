package podpedia

import (
	"context"
	"fmt"
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
func (p *GenericPipeline) Execute(ctx context.Context) error {
	docs, err := p.source.Ingest(ctx)
	if err != nil {
		return fmt.Errorf("source ingest: %w", err)
	}

	for i, doc := range docs {
		// Apply each transform in sequence
		for _, t := range p.transforms {
			if err := t.Process(ctx, doc); err != nil {
				return fmt.Errorf("transform on doc %d (%s): %w", i, doc.ID, err)
			}
		}

		// Emit to each sink
		for _, s := range p.sinks {
			if err := s.Emit(ctx, doc); err != nil {
				return fmt.Errorf("sink emit on doc %d (%s): %w", i, doc.ID, err)
			}
		}
	}

	return nil
}
