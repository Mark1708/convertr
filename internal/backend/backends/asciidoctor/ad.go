// Package asciidoctor provides AsciiDoc conversion via the asciidoctor binary.
package asciidoctor

import (
	"context"
	"os/exec"

	"git.mark1708.ru/me/convertr/internal/backend"
	"git.mark1708.ru/me/convertr/internal/backend/execx"
)

func init() {
	backend.Register(Backend{})
}

// Backend wraps the asciidoctor binary.
type Backend struct{}

func (Backend) Name() string       { return "asciidoctor" }
func (Backend) BinaryName() string { return "asciidoctor" }

func (Backend) Capabilities() []backend.Capability {
	caps := []backend.Capability{
		{From: "adoc", To: "html", Cost: 1, Quality: 95},
	}
	// PDF output requires the asciidoctor-pdf gem.
	if _, err := exec.LookPath("asciidoctor-pdf"); err == nil {
		caps = append(caps, backend.Capability{From: "adoc", To: "pdf", Cost: 3, Quality: 88})
	}
	return caps
}

func (b Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	var args []string

	if isPDF(out) {
		args = append(args, "-r", "asciidoctor-pdf", "-b", "pdf")
	}
	args = append(args, "-o", out)
	args = append(args, opts.ExtraArgs...)
	args = append(args, in)

	if err := execx.Run(ctx, "asciidoctor", args...); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	return nil
}

func isPDF(path string) bool {
	return len(path) > 4 && path[len(path)-4:] == ".pdf"
}
