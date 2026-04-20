package router

import (
	"github.com/Mark1708/convertr/internal/backend"
)

// edge represents a directed edge in the format graph.
type edge struct {
	to         string
	cost       int
	backendIdx int
	capIdx     int
}

// Graph holds the adjacency list for BFS/Dijkstra.
type Graph struct {
	adj      map[string][]edge
	backends []backend.Backend
}

// Build constructs the format graph from all registered backends.
func Build() *Graph {
	backends := backend.All()
	g := &Graph{
		adj:      make(map[string][]edge),
		backends: backends,
	}
	for bi, b := range backends {
		for ci, cap := range b.Capabilities() {
			g.adj[cap.From] = append(g.adj[cap.From], edge{
				to:         cap.To,
				cost:       cap.Cost,
				backendIdx: bi,
				capIdx:     ci,
			})
		}
	}
	return g
}

// Nodes returns all format IDs that appear in the graph.
func (g *Graph) Nodes() []string {
	seen := make(map[string]struct{})
	for from, edges := range g.adj {
		seen[from] = struct{}{}
		for _, e := range edges {
			seen[e.to] = struct{}{}
		}
	}
	nodes := make([]string, 0, len(seen))
	for n := range seen {
		nodes = append(nodes, n)
	}
	return nodes
}
