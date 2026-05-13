package graph_test

import (
	"testing"

	"github.com/gavmor/podpedia/internal/graph"
)

func TestMemoryGraphDB(t *testing.T) {
	db := graph.NewMemoryGraphDB()
	testGraphDB(t, db, "memory")
}

func TestSQLiteGraphDB(t *testing.T) {
	db := graph.NewSQLiteGraphDB()
	testGraphDB(t, db, "sqlite")
}

func testGraphDB(t *testing.T, db graph.GraphDB, kind string) {
	// Insert nodes
	if err := db.InsertNode("a", "NodeA", map[string]string{"key": "val"}); err != nil {
		t.Fatalf("[%s] InsertNode: %v", kind, err)
	}
	if err := db.InsertNode("b", "NodeB", nil); err != nil {
		t.Fatalf("[%s] InsertNode: %v", kind, err)
	}
	if err := db.InsertNode("c", "NodeC", map[string]string{"x": "y"}); err != nil {
		t.Fatalf("[%s] InsertNode: %v", kind, err)
	}

	// Insert edges: a → b, b → c, c → a
	if err := db.InsertEdge("a", "b", "knows"); err != nil {
		t.Fatalf("[%s] InsertEdge: %v", kind, err)
	}
	if err := db.InsertEdge("b", "c", "follows"); err != nil {
		t.Fatalf("[%s] InsertEdge: %v", kind, err)
	}
	if err := db.InsertEdge("c", "a", "links"); err != nil {
		t.Fatalf("[%s] InsertEdge: %v", kind, err)
	}

	// GetNeighbors for "a": should have b (outgoing) and c (incoming)
	neighbors, err := db.GetNeighbors("a")
	if err != nil {
		t.Fatalf("[%s] GetNeighbors(a): %v", kind, err)
	}
	if len(neighbors) != 2 {
		t.Errorf("[%s] GetNeighbors(a): expected 2, got %d: %+v", kind, len(neighbors), neighbors)
	}
	foundB, foundC := false, false
	for _, n := range neighbors {
		if n.ID == "b" {
			foundB = true
		}
		if n.ID == "c" {
			foundC = true
		}
	}
	if !foundB || !foundC {
		t.Errorf("[%s] GetNeighbors(a): expected b and c, got %+v", kind, neighbors)
	}

	// GetNeighbors for nonexistent
	_, err = db.GetNeighbors("nonexistent")
	if err != graph.ErrNotFound {
		t.Errorf("[%s] GetNeighbors(nonexistent): expected ErrNotFound, got %v", kind, err)
	}

	// Snapshot
	snap, err := db.Snapshot()
	if err != nil {
		t.Fatalf("[%s] Snapshot: %v", kind, err)
	}
	if len(snap.Nodes) != 3 {
		t.Errorf("[%s] Snapshot: expected 3 nodes, got %d", kind, len(snap.Nodes))
	}
	if len(snap.Edges) != 3 {
		t.Errorf("[%s] Snapshot: expected 3 edges, got %d", kind, len(snap.Edges))
	}

	// Verify node properties in snapshot
	nodeMap := make(map[string]graph.Node)
	for _, n := range snap.Nodes {
		nodeMap[n.ID] = n
	}
	if v, ok := nodeMap["a"].Properties["key"]; !ok || v != "val" {
		t.Errorf("[%s] Snapshot: node 'a' properties incorrect: %+v", kind, nodeMap["a"].Properties)
	}
}
