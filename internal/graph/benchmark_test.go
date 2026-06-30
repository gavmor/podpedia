package graph_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/gavmor/podpedia/internal/graph"
)

// setupGraph creates a graph with scale nodes and ~3*scale random edges.
func setupGraph(db graph.GraphDB, scale int) {
	for i := 0; i < scale; i++ {
		id := fmt.Sprintf("node-%d", i)
		label := fmt.Sprintf("Label-%d", i)
		props := map[string]string{
			"index": fmt.Sprintf("%d", i),
			"name":  fmt.Sprintf("Node %d", i),
		}
		_ = db.InsertNode(id, label, props)
	}

	// Insert ~3*scale random edges
	rng := rand.New(rand.NewSource(42))
	targetEdges := scale * 3
	for i := 0; i < targetEdges; i++ {
		from := fmt.Sprintf("node-%d", rng.Intn(scale))
		to := fmt.Sprintf("node-%d", rng.Intn(scale))
		label := fmt.Sprintf("edge-%d", i)
		_ = db.InsertEdge(from, to, label)
	}
}

func newDB(kind string) graph.GraphDB {
	switch kind {
	case "memory":
		return graph.NewMemoryGraphDB()
	case "sqlite":
		return graph.NewSQLiteGraphDB()
	default:
		panic("unknown db kind: " + kind)
	}
}

// bfsTraversal performs a full BFS from the given start node.
func bfsTraversal(db graph.GraphDB, startID string) error {
	visited := make(map[string]bool)
	queue := []string{startID}
	visited[startID] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		neighbors, err := db.GetNeighbors(current)
		if err != nil {
			return err
		}
		for _, n := range neighbors {
			if !visited[n.ID] {
				visited[n.ID] = true
				queue = append(queue, n.ID)
			}
		}
	}
	return nil
}

var scales = []int{100, 1000, 5000}
var dbKinds = []string{"memory", "sqlite"}

func BenchmarkGraphSnapshot(b *testing.B) {
	for _, kind := range dbKinds {
		for _, scale := range scales {
			name := fmt.Sprintf("%s/scale-%d", kind, scale)
			b.Run(name, func(b *testing.B) {
				b.StopTimer()
				db := newDB(kind)
				setupGraph(db, scale)
				b.StartTimer()

				for i := 0; i < b.N; i++ {
					_, _ = db.Snapshot()
				}
			})
		}
	}
}

func BenchmarkGraphTraversal(b *testing.B) {
	for _, kind := range dbKinds {
		for _, scale := range scales {
			name := fmt.Sprintf("%s/scale-%d", kind, scale)
			b.Run(name, func(b *testing.B) {
				b.StopTimer()
				db := newDB(kind)
				setupGraph(db, scale)
				b.StartTimer()

				for i := 0; i < b.N; i++ {
					_ = bfsTraversal(db, "node-0")
				}
			})
		}
	}
}
