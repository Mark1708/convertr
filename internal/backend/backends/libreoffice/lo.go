// Package libreoffice provides a backend that delegates to soffice (LibreOffice).
package libreoffice

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"git.mark1708.ru/me/convertr/internal/backend"
	"git.mark1708.ru/me/convertr/internal/backend/execx"
)

func init() {
	backend.Register(Backend{})
}

// Backend wraps the soffice binary.
type Backend struct{}

func (Backend) Name() string       { return "libreoffice" }
func (Backend) BinaryName() string { return "soffice" }

func (Backend) Capabilities() []backend.Capability {
	return []backend.Capability{
		// .doc (legacy Word) — only LibreOffice handles this well
		{From: "doc", To: "odt", Cost: 2, Quality: 90},
		{From: "doc", To: "docx", Cost: 2, Quality: 88},
		{From: "doc", To: "pdf", Cost: 3, Quality: 88},
		{From: "doc", To: "txt", Cost: 2, Quality: 75},
		// DOCX (pandoc is preferred for text; LO is better for fidelity)
		{From: "docx", To: "odt", Cost: 2, Quality: 92},
		{From: "docx", To: "pdf", Cost: 4, Quality: 90}, // cost 4 → router prefers pandoc (cost 3)
		{From: "docx", To: "txt", Cost: 2, Quality: 72},
		// ODT
		{From: "odt", To: "docx", Cost: 2, Quality: 90},
		{From: "odt", To: "pdf", Cost: 3, Quality: 90},
		{From: "odt", To: "txt", Cost: 2, Quality: 72},
		// Spreadsheets
		{From: "xlsx", To: "csv", Cost: 1, Quality: 95},
		{From: "xlsx", To: "ods", Cost: 2, Quality: 92},
		{From: "ods", To: "xlsx", Cost: 2, Quality: 90},
		{From: "ods", To: "csv", Cost: 1, Quality: 95},
		// Presentations
		{From: "pptx", To: "odp", Cost: 2, Quality: 85},
		{From: "pptx", To: "pdf", Cost: 3, Quality: 85},
		{From: "odp", To: "pptx", Cost: 2, Quality: 83},
		{From: "odp", To: "pdf", Cost: 3, Quality: 85},
		// RTF
		{From: "rtf", To: "odt", Cost: 2, Quality: 85},
		{From: "rtf", To: "docx", Cost: 2, Quality: 83},
		{From: "rtf", To: "pdf", Cost: 3, Quality: 85},
	}
}

func (b Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	outDir := filepath.Dir(out)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}

	// Extract target extension (without dot) from the output path.
	ext := filepath.Ext(out)
	if len(ext) > 1 {
		ext = ext[1:] // strip leading dot
	}

	// Per-process UserInstallation directory prevents lock conflicts
	// when multiple soffice processes run concurrently.
	userInstall := fmt.Sprintf("file:///tmp/convertr-lo-%d", os.Getpid())

	args := []string{
		"--headless",
		fmt.Sprintf("--env:UserInstallation=%s", userInstall),
		"--convert-to", ext,
		"--outdir", outDir,
		in,
	}
	args = append(args, opts.ExtraArgs...)

	if err := execx.Run(ctx, "soffice", args...); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}

	// soffice writes to outdir/<basename>.<ext> — rename to exact out path if needed.
	expected := filepath.Join(outDir, stripExt(filepath.Base(in))+"."+ext)
	if expected != out {
		if err := os.Rename(expected, out); err != nil {
			return backend.Wrap(b.Name(), in, out, err)
		}
	}
	return nil
}

func stripExt(name string) string {
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)]
}
