package graph

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

// SQLiteGraphDB is a graph database backed by SQLite (in-memory).
type SQLiteGraphDB struct {
	mu   sync.Mutex
	db   *sql.DB
	init sync.Once
	err  error
}

// NewSQLiteGraphDB creates a new SQLite-backed graph database using an
// in-memory SQLite database.
func NewSQLiteGraphDB() *SQLiteGraphDB {
	return &SQLiteGraphDB{}
}

func (g *SQLiteGraphDB) ensureInit() error {
	g.init.Do(func() {
		g.db, g.err = sql.Open("sqlite", ":memory:")
		if g.err != nil {
			return
		}
		// Use a single connection so :memory: is shared across all operations.
		g.db.SetMaxOpenConns(1)

		_, g.err = g.db.Exec(`
			CREATE TABLE IF NOT EXISTS nodes (
				id TEXT PRIMARY KEY,
				label TEXT NOT NULL
			);
			CREATE TABLE IF NOT EXISTS node_properties (
				node_id TEXT NOT NULL,
				key TEXT NOT NULL,
				value TEXT NOT NULL,
				PRIMARY KEY (node_id, key),
				FOREIGN KEY (node_id) REFERENCES nodes(id)
			);
			CREATE TABLE IF NOT EXISTS edges (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				from_id TEXT NOT NULL,
				to_id TEXT NOT NULL,
				label TEXT NOT NULL
			);
			CREATE INDEX IF NOT EXISTS idx_edges_from ON edges(from_id);
			CREATE INDEX IF NOT EXISTS idx_edges_to ON edges(to_id);
		`)
	})
	return g.err
}

// InsertNode adds a node to the graph.
func (g *SQLiteGraphDB) InsertNode(id, label string, properties map[string]string) error {
	if err := g.ensureInit(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	_, err := g.db.Exec("INSERT OR REPLACE INTO nodes (id, label) VALUES (?, ?)", id, label)
	if err != nil {
		return fmt.Errorf("insert node: %w", err)
	}

	// Clear existing properties and re-insert
	_, err = g.db.Exec("DELETE FROM node_properties WHERE node_id = ?", id)
	if err != nil {
		return fmt.Errorf("clear properties: %w", err)
	}

	for k, v := range properties {
		_, err := g.db.Exec("INSERT INTO node_properties (node_id, key, value) VALUES (?, ?, ?)", id, k, v)
		if err != nil {
			return fmt.Errorf("insert property: %w", err)
		}
	}

	return nil
}

// InsertEdge adds a directed edge between two nodes.
func (g *SQLiteGraphDB) InsertEdge(from, to, label string) error {
	if err := g.ensureInit(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	_, err := g.db.Exec("INSERT INTO edges (from_id, to_id, label) VALUES (?, ?, ?)", from, to, label)
	return err
}

// GetNeighbors returns all nodes connected to the given node.
func (g *SQLiteGraphDB) GetNeighbors(id string) ([]Node, error) {
	if err := g.ensureInit(); err != nil {
		return nil, err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Check node exists
	var count int
	err := g.db.QueryRow("SELECT COUNT(*) FROM nodes WHERE id = ?", id).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("check node: %w", err)
	}
	if count == 0 {
		return nil, ErrNotFound
	}

	// Get neighbors from both directions
	query := `
		SELECT DISTINCT n.id, n.label
		FROM nodes n
		INNER JOIN (
			SELECT to_id AS neighbor_id FROM edges WHERE from_id = ?
			UNION
			SELECT from_id AS neighbor_id FROM edges WHERE to_id = ?
		) e ON n.id = e.neighbor_id
		WHERE n.id != ?
	`

	rows, err := g.db.Query(query, id, id, id)
	if err != nil {
		return nil, fmt.Errorf("query neighbors: %w", err)
	}

	// Collect rows first, then close before loading properties
	// to avoid nested-query deadlock with MaxOpenConns(1).
	type neighborRow struct {
		id    string
		label string
	}
	var nr []neighborRow
	for rows.Next() {
		var r neighborRow
		if err := rows.Scan(&r.id, &r.label); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan neighbor: %w", err)
		}
		nr = append(nr, r)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	// Now load properties in a separate pass
	var neighbors []Node
	for _, r := range nr {
		props, err := g.loadProperties(r.id)
		if err != nil {
			return nil, err
		}
		neighbors = append(neighbors, Node{
			ID:         r.id,
			Label:      r.label,
			Properties: props,
		})
	}
	return neighbors, nil
}

func (g *SQLiteGraphDB) loadProperties(nodeID string) (map[string]string, error) {
	rows, err := g.db.Query("SELECT key, value FROM node_properties WHERE node_id = ?", nodeID)
	if err != nil {
		return nil, fmt.Errorf("query properties: %w", err)
	}
	defer func() { _ = rows.Close() }()

	props := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan property: %w", err)
		}
		props[k] = v
	}
	return props, rows.Err()
}

// Snapshot returns a full serializable representation of the graph.
func (g *SQLiteGraphDB) Snapshot() (*GraphSnapshot, error) {
	if err := g.ensureInit(); err != nil {
		return nil, err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Load all nodes (IDs and labels first, then properties)
	nodeRows, err := g.db.Query("SELECT id, label FROM nodes")
	if err != nil {
		return nil, fmt.Errorf("query nodes: %w", err)
	}

	type nodeRow struct {
		id    string
		label string
	}
	var nr []nodeRow
	for nodeRows.Next() {
		var r nodeRow
		if err := nodeRows.Scan(&r.id, &r.label); err != nil {
			_ = nodeRows.Close()
			return nil, fmt.Errorf("scan node: %w", err)
		}
		nr = append(nr, r)
	}
	if err := nodeRows.Err(); err != nil {
		_ = nodeRows.Close()
		return nil, err
	}
	if err := nodeRows.Close(); err != nil {
		return nil, err
	}

	// Load properties for each node in a separate pass
	var nodes []Node
	for _, r := range nr {
		props, err := g.loadProperties(r.id)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, Node{
			ID:         r.id,
			Label:      r.label,
			Properties: props,
		})
	}

	// Load all edges
	edgeRows, err := g.db.Query("SELECT from_id, to_id, label FROM edges ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("query edges: %w", err)
	}

	var edges []Edge
	for edgeRows.Next() {
		var e Edge
		if err := edgeRows.Scan(&e.From, &e.To, &e.Label); err != nil {
			_ = edgeRows.Close()
			return nil, fmt.Errorf("scan edge: %w", err)
		}
		edges = append(edges, e)
	}
	if err := edgeRows.Err(); err != nil {
		_ = edgeRows.Close()
		return nil, err
	}
	if err := edgeRows.Close(); err != nil {
		return nil, err
	}

	if nodes == nil {
		nodes = make([]Node, 0)
	}
	if edges == nil {
		edges = make([]Edge, 0)
	}

	return &GraphSnapshot{
		Nodes: nodes,
		Edges: edges,
	}, nil
}
