// Package jq provides JSON formatting/transformation via the jq binary.
package jq

import (
	"context"
	"os"
	"strings"

	"github.com/Mark1708/convertr/internal/backend"
	"github.com/Mark1708/convertr/internal/backend/execx"
)

func init() {
	backend.Register(Backend{})
}

// Backend wraps the jq binary.
type Backend struct{}

func (Backend) Name() string       { return "jq" }
func (Backend) BinaryName() string { return "jq" }

func (Backend) Capabilities() []backend.Capability {
	return []backend.Capability{
		// json → json: format/minify/transform
		{From: "json", To: "json", Cost: 1, Quality: 100},
	}
}

func (b Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	// Default filter is identity (pretty-print).
	filter := "."

	// Support minification via named option "jq.minify=true".
	minify := strings.EqualFold(opts.Get("jq", "minify"), "true")

	// Support custom transform via named option "jq.filter=EXPR" or ExtraArgs.
	if f := opts.Get("jq", "filter"); f != "" {
		filter = f
	}

	args := []string{}
	if minify {
		args = append(args, "-c")
	}
	args = append(args, filter, in)
	args = append(args, opts.ExtraArgs...)

	outBytes, err := execx.Output(ctx, "jq", args...)
	if err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	if err := os.WriteFile(out, outBytes, 0o644); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	return nil
}
