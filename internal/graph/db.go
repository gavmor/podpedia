// Package graph provides in-memory knowledge graph backends.
package graph

import "fmt"

// Node represents a vertex in the graph.
type Node struct {
	ID         string            `json:"id"`
	Label      string            `json:"label"`
	Properties map[string]string `json:"properties,omitempty"`
}

// Edge represents a directed edge between two nodes.
type Edge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label"`
}

// GraphSnapshot is a serializable representation of the entire graph.
type GraphSnapshot struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// GraphDB is the interface for graph database backends.
type GraphDB interface {
	// InsertNode adds a node to the graph.
	InsertNode(id, label string, properties map[string]string) error

	// InsertEdge adds a directed edge between two nodes.
	InsertEdge(from, to, label string) error

	// GetNeighbors returns all nodes connected to the given node by any edge.
	GetNeighbors(id string) ([]Node, error)

	// Snapshot returns a serializable representation of the entire graph.
	Snapshot() (*GraphSnapshot, error)
}

// ErrNotFound is returned when a node is not found.
var ErrNotFound = fmt.Errorf("node not found")
