// Package pandoc provides a backend that delegates conversions to pandoc.
package pandoc

import (
	"context"
	"os/exec"

	"github.com/Mark1708/convertr/internal/backend"
	"github.com/Mark1708/convertr/internal/backend/execx"
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

	// Always pass explicit --from/--to so pandoc never guesses from file extensions.
	// This matters when the source file extension doesn't match its actual content
	// (e.g. an HTML file saved as .doc).
	if from := opts.Named["step.from"]; from != "" {
		args = append(args, "--from", pandocFormat(from))
	}
	if to := opts.Named["step.to"]; to != "" {
		args = append(args, "--to", pandocFormat(to))
	}

	// Named options: toc, standalone, highlight, template, pdf_engine.
	if opts.Get("pandoc", "toc") == "true" {
		args = append(args, "--toc")
	}
	if opts.Get("pandoc", "standalone") == "true" {
		args = append(args, "-s")
	}
	if style := opts.Get("pandoc", "highlight"); style != "" {
		args = append(args, "--highlight-style", style)
	}
	if tpl := opts.Get("pandoc", "template"); tpl != "" {
		args = append(args, "--template", tpl)
	}

	if opts.StripMeta {
		args = append(args, "--strip-comments")
	}

	// PDF engine: explicit override takes priority, then auto-detect.
	if isPDF(out) {
		if engine := opts.Get("pandoc", "pdf_engine"); engine != "" {
			args = append(args, "--pdf-engine="+engine)
		} else if _, err := exec.LookPath("xelatex"); err == nil {
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

// pandocFormat maps convertr format IDs to pandoc format names.
func pandocFormat(id string) string {
	switch id {
	case "md":
		return "markdown"
	case "tex":
		return "latex"
	case "txt":
		return "plain"
	default:
		return id
	}
}

func isPDF(path string) bool {
	return len(path) > 4 && path[len(path)-4:] == ".pdf"
}
