// Package tesseract provides OCR via the tesseract binary.
package tesseract

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.mark1708.ru/me/convertr/internal/backend"
	"git.mark1708.ru/me/convertr/internal/backend/execx"
)

func init() {
	backend.Register(Backend{})
}

// Backend wraps the tesseract binary.
type Backend struct{}

func (Backend) Name() string       { return "tesseract" }
func (Backend) BinaryName() string { return "tesseract" }

func (Backend) Capabilities() []backend.Capability {
	return []backend.Capability{
		{From: "jpg", To: "txt", Cost: 3, Quality: 75},
		{From: "png", To: "txt", Cost: 3, Quality: 80},
		{From: "tiff", To: "txt", Cost: 3, Quality: 80},
	}
}

func (b Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	// Tesseract appends ".txt" to the output stem automatically.
	// We pass the stem (without extension) and rename afterwards.
	stem := strings.TrimSuffix(out, filepath.Ext(out))

	lang := opts.Get("tesseract", "lang")
	if lang == "" {
		lang = "rus+eng"
	}
	oem := opts.Get("tesseract", "oem")
	if oem == "" {
		oem = "1"
	}
	psm := opts.Get("tesseract", "psm")
	if psm == "" {
		psm = "3"
	}

	args := []string{in, stem, "--oem", oem, "--psm", psm, "-l", lang}
	args = append(args, opts.ExtraArgs...)

	if err := execx.Run(ctx, "tesseract", args...); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}

	// Tesseract writes to stem+".txt"; rename to exact out path if different.
	produced := stem + ".txt"
	if produced != out {
		if err := os.Rename(produced, out); err != nil {
			return backend.Wrap(b.Name(), in, out,
				fmt.Errorf("rename %s → %s: %w", produced, out, err))
		}
	}
	return nil
}
