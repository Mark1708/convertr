// Package figlet converts plain text to ASCII art via the figlet binary.
package figlet

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"github.com/Mark1708/convertr/internal/backend"
)

func init() {
	backend.Register(Backend{})
}

// Backend wraps the figlet binary.
type Backend struct{}

func (Backend) Name() string       { return "figlet" }
func (Backend) BinaryName() string { return "figlet" }

func (Backend) Capabilities() []backend.Capability {
	return []backend.Capability{
		{From: "txt", To: "ascii", Cost: 1, Quality: 100},
	}
}

func (b Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	input, err := os.ReadFile(in)
	if err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}

	font := opts.Get("figlet", "font")
	if font == "" {
		font = "standard"
	}
	args := []string{"-f", font}
	args = append(args, opts.ExtraArgs...)

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "figlet", args...) //nolint:gosec
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stderr = &stderr
	result, err := cmd.Output()
	if err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	if err := os.WriteFile(out, result, 0o644); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	return nil
}
