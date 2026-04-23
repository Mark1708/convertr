//go:build darwin

// Package textutil provides a macOS-native backend using the textutil(1) command.
// Available only on macOS where textutil is bundled with the OS.
package textutil

import (
	"context"

	"github.com/Mark1708/convertr/internal/backend"
	"github.com/Mark1708/convertr/internal/backend/execx"
)

func init() {
	backend.Register(Backend{})
}

// Backend wraps the macOS textutil command.
type Backend struct{}

func (Backend) Name() string       { return "textutil" }
func (Backend) BinaryName() string { return "textutil" }

func (Backend) Capabilities() []backend.Capability {
	return []backend.Capability{
		// Legacy Word and RTF handled natively on macOS without external deps.
		// Cost is deliberately higher than the cross-platform pandoc edges so
		// that Dijkstra prefers portable backends; textutil stays as a macOS
		// fallback when pandoc is unavailable.
		{From: "doc", To: "txt", Cost: 3, Quality: 80},
		{From: "doc", To: "html", Cost: 3, Quality: 80},
		{From: "rtf", To: "txt", Cost: 3, Quality: 85},
		{From: "rtf", To: "html", Cost: 3, Quality: 85},
	}
}

func (b Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	// textutil -convert FORMAT -output OUT IN
	// FORMAT is derived from the output extension (txt → txt, html → html).
	ext := outputFormat(out)
	args := []string{"-convert", ext, "-output", out, in}
	args = append(args, opts.ExtraArgs...)

	if err := execx.Run(ctx, "textutil", args...); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	return nil
}

func outputFormat(path string) string {
	if len(path) > 5 && path[len(path)-5:] == ".html" {
		return "html"
	}
	return "txt"
}
