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
//
// Edges are filtered by backend.IsAvailable so that capabilities whose
// underlying binary is missing from $PATH never appear in the graph.
// This makes Dijkstra prefer cheap routes that can actually run instead
// of failing at exec time on a backend with the same nominal Cost.
func Build() *Graph {
	return BuildFromBackends(backend.All())
}

// BuildFromBackends constructs the format graph from the supplied
// backends. Exposed for testing and for callers that need to scope a
// graph to a subset of backends (e.g. plugin sandboxing).
func BuildFromBackends(backends []backend.Backend) *Graph {
	g := &Graph{
		adj:      make(map[string][]edge),
		backends: backends,
	}
	for bi, b := range backends {
		for ci, cap := range b.Capabilities() {
			if !backend.IsAvailable(b, cap.From, cap.To) {
				continue
			}
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
