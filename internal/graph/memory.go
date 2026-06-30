package graph

import "sync"

// MemoryGraphDB is a pure in-memory graph database using maps and slices.
type MemoryGraphDB struct {
	mu    sync.RWMutex
	nodes map[string]Node
	edges []Edge
}

// NewMemoryGraphDB creates a new empty in-memory graph database.
func NewMemoryGraphDB() *MemoryGraphDB {
	return &MemoryGraphDB{
		nodes: make(map[string]Node),
		edges: make([]Edge, 0),
	}
}

// InsertNode adds a node to the graph. If a node with the same ID already
// exists, it is overwritten.
func (db *MemoryGraphDB) InsertNode(id, label string, properties map[string]string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	props := properties
	if props == nil {
		props = make(map[string]string)
	}

	db.nodes[id] = Node{
		ID:         id,
		Label:      label,
		Properties: props,
	}
	return nil
}

// InsertEdge adds a directed edge between two nodes.
func (db *MemoryGraphDB) InsertEdge(from, to, label string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.edges = append(db.edges, Edge{
		From:  from,
		To:    to,
		Label: label,
	})
	return nil
}

// GetNeighbors returns all nodes connected to the given node by any edge
// (outgoing or incoming).
func (db *MemoryGraphDB) GetNeighbors(id string) ([]Node, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if _, ok := db.nodes[id]; !ok {
		return nil, ErrNotFound
	}

	seen := make(map[string]bool)
	var neighbors []Node

	for _, e := range db.edges {
		if e.From == id {
			if !seen[e.To] {
				if n, ok := db.nodes[e.To]; ok {
					neighbors = append(neighbors, n)
					seen[e.To] = true
				}
			}
		}
		if e.To == id {
			if !seen[e.From] {
				if n, ok := db.nodes[e.From]; ok {
					neighbors = append(neighbors, n)
					seen[e.From] = true
				}
			}
		}
	}

	return neighbors, nil
}

// Snapshot returns a deep copy of the entire graph.
func (db *MemoryGraphDB) Snapshot() (*GraphSnapshot, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	nodes := make([]Node, 0, len(db.nodes))
	for _, n := range db.nodes {
		props := make(map[string]string, len(n.Properties))
		for k, v := range n.Properties {
			props[k] = v
		}
		nodes = append(nodes, Node{
			ID:         n.ID,
			Label:      n.Label,
			Properties: props,
		})
	}

	edges := make([]Edge, len(db.edges))
	copy(edges, db.edges)

	return &GraphSnapshot{
		Nodes: nodes,
		Edges: edges,
	}, nil
}
