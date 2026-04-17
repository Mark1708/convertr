package router

import (
	"context"
	"testing"

	"git.mark1708.ru/me/convertr/internal/backend"
)

type mockBackend struct {
	name string
	caps []backend.Capability
}

func (m mockBackend) Name() string                                                       { return m.name }
func (m mockBackend) BinaryName() string                                                 { return m.name }
func (m mockBackend) Capabilities() []backend.Capability                                 { return m.caps }
func (m mockBackend) Convert(_ context.Context, _, _ string, _ backend.Options) error    { return nil }

func buildTestGraph(backends ...backend.Backend) *Graph {
	g := &Graph{adj: make(map[string][]edge), backends: backends}
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

func TestBFS_sameFormat(t *testing.T) {
	g := buildTestGraph()
	route, err := g.Find("md", "md")
	if err != nil {
		t.Fatal(err)
	}
	if len(route.Steps) != 0 {
		t.Fatalf("expected 0 steps for same format, got %d", len(route.Steps))
	}
}

func TestBFS_singleHop(t *testing.T) {
	b := mockBackend{name: "pandoc", caps: []backend.Capability{
		{From: "md", To: "html", Cost: 1},
	}}
	g := buildTestGraph(b)
	route, err := g.Find("md", "html")
	if err != nil {
		t.Fatal(err)
	}
	if len(route.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(route.Steps))
	}
	s := route.Steps[0]
	if s.From != "md" || s.To != "html" {
		t.Errorf("step: got %s→%s, want md→html", s.From, s.To)
	}
	if s.Backend.Name() != "pandoc" {
		t.Errorf("backend: got %q, want pandoc", s.Backend.Name())
	}
}

func TestBFS_multiHop(t *testing.T) {
	b := mockBackend{name: "pandoc", caps: []backend.Capability{
		{From: "docx", To: "md", Cost: 2},
		{From: "md", To: "html", Cost: 1},
	}}
	g := buildTestGraph(b)
	route, err := g.Find("docx", "html")
	if err != nil {
		t.Fatal(err)
	}
	if len(route.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(route.Steps))
	}
	if route.Steps[0].From != "docx" || route.Steps[0].To != "md" {
		t.Errorf("step 0: got %s→%s, want docx→md", route.Steps[0].From, route.Steps[0].To)
	}
	if route.Steps[1].From != "md" || route.Steps[1].To != "html" {
		t.Errorf("step 1: got %s→%s, want md→html", route.Steps[1].From, route.Steps[1].To)
	}
}

func TestBFS_noRoute(t *testing.T) {
	b := mockBackend{name: "pandoc", caps: []backend.Capability{
		{From: "md", To: "html", Cost: 1},
	}}
	g := buildTestGraph(b)
	_, err := g.Find("mp4", "mp3")
	if err == nil {
		t.Fatal("expected ErrNoRoute, got nil")
	}
}

func TestBFS_prefersLowerCost(t *testing.T) {
	b := mockBackend{name: "pandoc", caps: []backend.Capability{
		{From: "md", To: "html", Cost: 1},
		{From: "md", To: "pdf", Cost: 3},
		{From: "html", To: "pdf", Cost: 1},
	}}
	g := buildTestGraph(b)
	route, err := g.Find("md", "pdf")
	if err != nil {
		t.Fatal(err)
	}
	// md→pdf directly (cost 3) vs md→html→pdf (cost 2) — prefer cheaper
	if len(route.Steps) != 2 {
		t.Errorf("expected 2-step route (cheaper), got %d steps", len(route.Steps))
	}
}

func TestBFS_maxHopsRespected(t *testing.T) {
	// Chain longer than maxHops=4
	b := mockBackend{name: "x", caps: []backend.Capability{
		{From: "a", To: "b", Cost: 1},
		{From: "b", To: "c", Cost: 1},
		{From: "c", To: "d", Cost: 1},
		{From: "d", To: "e", Cost: 1},
		{From: "e", To: "z", Cost: 1},
	}}
	g := buildTestGraph(b)
	_, err := g.Find("a", "z")
	if err == nil {
		t.Fatal("expected ErrNoRoute for chain longer than maxHops, got nil")
	}
}
