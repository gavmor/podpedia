// Package podpedia provides the common intermediate representation (IR) and
// pipeline interfaces for the Podpedia workflow engine.
//
// All pipeline stages operate on a Document, which flows through three
// abstract phases: Source → Transform → Sink. This decouples input types
// (text upload, RSS feed, audio URL) from output types (graph JSON, markdown,
// cloud storage) so any input can produce any output.
package podpedia

// Document is the common intermediate representation that flows through
// every pipeline stage.
type Document struct {
	// ID uniquely identifies this document within a pipeline run.
	ID string `json:"id"`

	// SourceID identifies the origin (e.g., podcast slug, user upload ID, RSS URL).
	SourceID string `json:"source_id"`

	// Metadata carries source-specific attributes (title, date, URL, author, etc.).
	Metadata map[string]string `json:"metadata,omitempty"`

	// Content is the raw text corpus — either the full transcript (podcast) or
	// user-provided text (upload).
	Content string `json:"content"`

	// Graph is populated by the EntityExtractor transform stage.
	// nil until the transform runs.
	Graph *Encyclopedia `json:"graph,omitempty"`
}

// Encyclopedia represents the entity-extraction graph for a document.
// This is the structured meaning extracted from Content by the LLM.
type Encyclopedia struct {
	Entities      []Entity       `json:"entities"`
	Relationships []Relationship `json:"relationships"`
}

// Entity is a node in the knowledge graph.
type Entity struct {
	ID   string `json:"id"`
	Type string `json:"type"` // Organization, Person, Product, Concept, Location, Event, Unknown
}

// Relationship is a directed edge in the knowledge graph.
type Relationship struct {
	Source  string `json:"source"`
	Target  string `json:"target"`
	Relation string `json:"relation"` // e.g., FOUNDED_BY, INVESTED_IN
	Context string `json:"context,omitempty"` // one-sentence justification
}
