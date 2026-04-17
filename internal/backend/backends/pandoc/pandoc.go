// Package pandoc provides a backend that delegates conversions to pandoc.
package pandoc

import (
	"context"
	"os/exec"

	"git.mark1708.ru/me/convertr/internal/backend"
	"git.mark1708.ru/me/convertr/internal/backend/execx"
)

func init() {
	backend.Register(Backend{})
}

// Backend wraps the pandoc binary.
type Backend struct{}

func (Backend) Name() string       { return "pandoc" }
func (Backend) BinaryName() string { return "pandoc" }

func (Backend) Capabilities() []backend.Capability {
	return []backend.Capability{
		// Markdown
		{From: "md", To: "html", Cost: 1, Quality: 95},
		{From: "md", To: "docx", Cost: 2, Quality: 90},
		{From: "md", To: "odt", Cost: 2, Quality: 90},
		{From: "md", To: "pdf", Cost: 3, Quality: 90},
		{From: "md", To: "rst", Cost: 1, Quality: 90},
		{From: "md", To: "epub", Cost: 2, Quality: 85},
		{From: "md", To: "tex", Cost: 2, Quality: 90},
		{From: "md", To: "txt", Cost: 1, Quality: 80},
		// HTML
		{From: "html", To: "md", Cost: 2, Quality: 80},
		{From: "html", To: "docx", Cost: 2, Quality: 85},
		{From: "html", To: "pdf", Cost: 3, Quality: 88},
		{From: "html", To: "txt", Cost: 1, Quality: 75},
		// DOCX
		{From: "docx", To: "md", Cost: 2, Quality: 80},
		{From: "docx", To: "html", Cost: 2, Quality: 85},
		{From: "docx", To: "pdf", Cost: 3, Quality: 88},
		{From: "docx", To: "odt", Cost: 2, Quality: 85},
		{From: "docx", To: "txt", Cost: 1, Quality: 75},
		{From: "docx", To: "rst", Cost: 2, Quality: 75},
		// ODT
		{From: "odt", To: "md", Cost: 2, Quality: 78},
		{From: "odt", To: "docx", Cost: 2, Quality: 85},
		{From: "odt", To: "html", Cost: 2, Quality: 82},
		{From: "odt", To: "txt", Cost: 1, Quality: 72},
		// RST
		{From: "rst", To: "md", Cost: 1, Quality: 85},
		{From: "rst", To: "html", Cost: 1, Quality: 90},
		{From: "rst", To: "docx", Cost: 2, Quality: 80},
		{From: "rst", To: "pdf", Cost: 3, Quality: 85},
		{From: "rst", To: "txt", Cost: 1, Quality: 80},
		// EPUB
		{From: "epub", To: "md", Cost: 2, Quality: 75},
		{From: "epub", To: "html", Cost: 2, Quality: 82},
		{From: "epub", To: "txt", Cost: 1, Quality: 70},
		// TeX/LaTeX
		{From: "tex", To: "md", Cost: 2, Quality: 75},
		{From: "tex", To: "html", Cost: 2, Quality: 80},
		{From: "tex", To: "pdf", Cost: 3, Quality: 95},
		// Org
		{From: "org", To: "md", Cost: 1, Quality: 85},
		{From: "org", To: "html", Cost: 1, Quality: 85},
		{From: "org", To: "pdf", Cost: 3, Quality: 85},
	}
}

func (b Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	args := []string{in, "-o", out}

	// Prefer xelatex for PDF output; fall back to pdflatex.
	if isPDF(out) {
		if _, err := exec.LookPath("xelatex"); err == nil {
			args = append(args, "--pdf-engine=xelatex")
		} else if _, err := exec.LookPath("pdflatex"); err == nil {
			args = append(args, "--pdf-engine=pdflatex")
		}
	}

	args = append(args, opts.ExtraArgs...)

	if err := execx.Run(ctx, "pandoc", args...); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	return nil
}

func isPDF(path string) bool {
	return len(path) > 4 && path[len(path)-4:] == ".pdf"
}
