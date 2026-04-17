package router

import (
	"container/heap"
	"fmt"

	"git.mark1708.ru/me/convertr/internal/backend"
)

const maxHops = 4

// item is a priority queue entry for Dijkstra.
type item struct {
	format string
	cost   int
	path   []edge
	index  int
}

type priorityQueue []*item

func (pq priorityQueue) Len() int            { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool  { return pq[i].cost < pq[j].cost }
func (pq priorityQueue) Swap(i, j int)       { pq[i], pq[j] = pq[j], pq[i]; pq[i].index = i; pq[j].index = j }
func (pq *priorityQueue) Push(x interface{}) { it := x.(*item); it.index = len(*pq); *pq = append(*pq, it) }
func (pq *priorityQueue) Pop() interface{}   { old := *pq; n := len(old); it := old[n-1]; *pq = old[:n-1]; return it }

// Find returns the lowest-cost route from src to dst, or ErrNoRoute.
func (g *Graph) Find(src, dst string) (*Route, error) {
	if src == dst {
		return &Route{}, nil
	}

	dist := map[string]int{src: 0}
	pq := &priorityQueue{{format: src, cost: 0}}
	heap.Init(pq)

	for pq.Len() > 0 {
		cur := heap.Pop(pq).(*item)
		if cur.format == dst {
			return buildRoute(g, cur.path), nil
		}
		if cur.cost > dist[cur.format] {
			continue
		}
		if len(cur.path) >= maxHops {
			continue
		}
		for _, e := range g.adj[cur.format] {
			newCost := cur.cost + e.cost
			if prev, ok := dist[e.to]; ok && newCost >= prev {
				continue
			}
			dist[e.to] = newCost
			newPath := make([]edge, len(cur.path)+1)
			copy(newPath, cur.path)
			newPath[len(cur.path)] = e
			heap.Push(pq, &item{format: e.to, cost: newCost, path: newPath})
		}
	}
	return nil, fmt.Errorf("%w: %s → %s", backend.ErrNoRoute, src, dst)
}

func buildRoute(g *Graph, path []edge) *Route {
	steps := make([]Step, len(path))
	for i, e := range path {
		caps := g.backends[e.backendIdx].Capabilities()
		steps[i] = Step{
			From:    caps[e.capIdx].From,
			To:      caps[e.capIdx].To,
			Backend: g.backends[e.backendIdx],
		}
	}
	return &Route{Steps: steps}
}
