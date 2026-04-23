// Package pandoc provides a backend that delegates conversions to pandoc.
package pandoc

import (
	"context"
	"os/exec"
	"strings"

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
		{From: "md", To: "typst", Cost: 2, Quality: 85},
		{From: "md", To: "ipynb", Cost: 2, Quality: 80},
		{From: "md", To: "pptx", Cost: 3, Quality: 75},
		{From: "md", To: "mediawiki", Cost: 2, Quality: 80},
		{From: "md", To: "jira", Cost: 1, Quality: 85},
		{From: "md", To: "opml", Cost: 1, Quality: 85},
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
		// Typst
		{From: "typst", To: "md", Cost: 2, Quality: 80},
		{From: "typst", To: "pdf", Cost: 3, Quality: 90},
		// Jupyter Notebook
		{From: "ipynb", To: "md", Cost: 2, Quality: 85},
		{From: "ipynb", To: "html", Cost: 2, Quality: 85},
		{From: "ipynb", To: "pdf", Cost: 3, Quality: 85},
		// PowerPoint (lossy input — text + structure survive, formatting is approximate).
		{From: "pptx", To: "md", Cost: 3, Quality: 60, Lossy: true},
		// Wiki / lightweight markup
		{From: "mediawiki", To: "md", Cost: 2, Quality: 80},
		{From: "dokuwiki", To: "md", Cost: 2, Quality: 75},
		{From: "jira", To: "md", Cost: 1, Quality: 85},
		{From: "textile", To: "md", Cost: 2, Quality: 80},
		// Bibliography
		{From: "bibtex", To: "csljson", Cost: 1, Quality: 95},
		{From: "csljson", To: "bibtex", Cost: 1, Quality: 95},
		{From: "bibtex", To: "md", Cost: 2, Quality: 80},
		// FictionBook / DocBook / OPML
		{From: "fb2", To: "md", Cost: 2, Quality: 75},
		{From: "fb2", To: "html", Cost: 2, Quality: 78},
		{From: "fb2", To: "epub", Cost: 2, Quality: 80},
		{From: "docbook", To: "md", Cost: 2, Quality: 75},
		{From: "opml", To: "md", Cost: 1, Quality: 85},
		// RTF (cross-platform complement to the macOS-only textutil backend).
		// Cost is set low enough that pandoc's rtf→md→txt/html route wins
		// against the macOS-native textutil edges (Cost=3 each) and keeps
		// cross-platform behaviour the default.
		{From: "rtf", To: "md", Cost: 1, Quality: 75},
	}
}

func (b Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	args := buildArgs(in, out, opts)
	if err := execx.Run(ctx, "pandoc", args...); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	return nil
}

// buildArgs assembles the argv passed to pandoc. It is pure (no process
// lookups other than PDF-engine probing) so the logic can be unit-tested
// without pandoc installed.
func buildArgs(in, out string, opts backend.Options) []string {
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
	var engine string
	if isPDF(out) {
		engine = opts.Get("pandoc", "pdf_engine")
		if engine == "" {
			engine = detectPDFEngine()
		}
		if engine != "" {
			args = append(args, "--pdf-engine="+engine)
		}
	}

	// PDF font defaults: only for fontspec-aware engines and only when the
	// user has not already set the corresponding variable via --named or
	// ExtraArgs. pdflatex does not understand fontspec, so skip it there.
	if isPDF(out) && (engine == "xelatex" || engine == "lualatex") {
		args = appendFontDefault(args, opts, "mainfont", opts.Fonts.Mainfont)
		args = appendFontDefault(args, opts, "monofont", opts.Fonts.Monofont)
		args = appendFontDefault(args, opts, "sansfont", opts.Fonts.Sansfont)
		if !hasVariable(opts.ExtraArgs, "geometry") && opts.Get("pandoc", "geometry") == "" {
			args = append(args, "-V", "geometry:margin=2cm")
		}
	}

	args = append(args, opts.ExtraArgs...)
	return args
}

// appendFontDefault injects "-V name=value" only if the user has not supplied
// the variable via --named pandoc.<name> or an explicit -V flag in ExtraArgs.
// Precedence: --named (highest) > ExtraArgs user -V > fallback value > skip.
func appendFontDefault(args []string, opts backend.Options, name, fallback string) []string {
	if v := opts.Get("pandoc", name); v != "" {
		return append(args, "-V", name+"="+v)
	}
	if hasVariable(opts.ExtraArgs, name) {
		return args
	}
	if fallback == "" {
		return args
	}
	return append(args, "-V", name+"="+fallback)
}

// hasVariable reports whether pandoc-style variable `name` is already
// present in extra. Recognised shapes:
//
//	-V name=value              (name as separate token)
//	-V name:value              (colon-style, e.g. geometry:margin=2cm)
//	-Vname=value               (glued)
//	--variable name=value
//	--variable=name=value
func hasVariable(extra []string, name string) bool {
	if name == "" {
		return false
	}
	prefixes := []string{name + "=", name + ":"}
	for i := range extra {
		a := extra[i]
		switch {
		case a == "-V" || a == "--variable":
			if i+1 < len(extra) && matchesAny(extra[i+1], prefixes) {
				return true
			}
		case strings.HasPrefix(a, "-V"):
			rest := strings.TrimPrefix(a, "-V")
			if matchesAny(rest, prefixes) {
				return true
			}
		case strings.HasPrefix(a, "--variable="):
			rest := strings.TrimPrefix(a, "--variable=")
			if matchesAny(rest, prefixes) {
				return true
			}
		}
	}
	return false
}

func matchesAny(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// detectPDFEngine picks the first available LaTeX engine that supports
// fontspec (xelatex, lualatex) before falling back to pdflatex.
func detectPDFEngine() string {
	for _, bin := range []string{"xelatex", "lualatex", "pdflatex"} {
		if _, err := exec.LookPath(bin); err == nil {
			return bin
		}
	}
	return ""
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

// Compile-time assurance that the backend still satisfies the interface.
var _ backend.Backend = Backend{}
