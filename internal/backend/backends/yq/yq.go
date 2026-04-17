// Package yq provides YAML/JSON/TOML conversion via the yq binary.
package yq

import (
	"context"
	"os"

	"git.mark1708.ru/me/convertr/internal/backend"
	"git.mark1708.ru/me/convertr/internal/backend/execx"
)

func init() {
	backend.Register(Backend{})
}

// Backend wraps the yq binary.
type Backend struct{}

func (Backend) Name() string       { return "yq" }
func (Backend) BinaryName() string { return "yq" }

// yqFormat maps our canonical format IDs to yq output format names.
var yqFormat = map[string]string{
	"yaml": "yaml",
	"json": "json",
	"toml": "toml",
}

func (Backend) Capabilities() []backend.Capability {
	formats := []string{"yaml", "json", "toml"}
	caps := make([]backend.Capability, 0, len(formats)*len(formats))
	for _, from := range formats {
		for _, to := range formats {
			if from != to {
				caps = append(caps, backend.Capability{From: from, To: to, Cost: 1, Quality: 95})
			}
		}
	}
	return caps
}

func (b Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	toFmt := yqFmtFromExt(out)
	args := []string{"-o", toFmt, ".", in}
	args = append(args, opts.ExtraArgs...)

	outBytes, err := execx.Output(ctx, "yq", args...)
	if err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	if err := os.WriteFile(out, outBytes, 0o644); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	return nil
}

func yqFmtFromExt(path string) string {
	if len(path) > 5 && path[len(path)-5:] == ".yaml" {
		return "yaml"
	}
	if len(path) > 4 && path[len(path)-4:] == ".yml" {
		return "yaml"
	}
	if len(path) > 5 && path[len(path)-5:] == ".toml" {
		return "toml"
	}
	return "json"
}
