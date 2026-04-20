// Package imagemagick provides image conversion via ImageMagick (magick or convert).
package imagemagick

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/Mark1708/convertr/internal/backend"
	"github.com/Mark1708/convertr/internal/backend/execx"
)

func init() {
	backend.Register(Backend{})
}

// Backend wraps ImageMagick (prefers IM7 `magick`, falls back to IM6 `convert`).
type Backend struct{}

func (Backend) Name() string { return "imagemagick" }
func (Backend) BinaryName() string {
	if _, err := exec.LookPath("magick"); err == nil {
		return "magick"
	}
	return "convert"
}

func (Backend) Capabilities() []backend.Capability {
	caps := []backend.Capability{}

	// Raster image formats — bidirectional
	raster := []string{"jpg", "png", "webp", "gif", "tiff", "bmp"}
	for _, from := range raster {
		for _, to := range raster {
			if from != to {
				caps = append(caps, backend.Capability{From: from, To: to, Cost: 1, Quality: 90})
			}
		}
	}

	// AVIF (needs newer ImageMagick with HEIC/AVIF support)
	for _, from := range raster {
		caps = append(caps, backend.Capability{From: from, To: "avif", Cost: 2, Quality: 85, Lossy: true})
		caps = append(caps, backend.Capability{From: "avif", To: from, Cost: 2, Quality: 88})
	}

	// SVG → raster (one-way; rasterising SVG is lossy)
	for _, to := range raster {
		caps = append(caps, backend.Capability{From: "svg", To: to, Cost: 2, Quality: 85, Lossy: true})
	}
	caps = append(caps, backend.Capability{From: "svg", To: "avif", Cost: 2, Quality: 82, Lossy: true})

	// HEIC → raster (decode only)
	for _, to := range raster {
		caps = append(caps, backend.Capability{From: "heic", To: to, Cost: 2, Quality: 88})
	}

	return caps
}

func (b Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	binary := b.BinaryName()
	var args []string

	if binary == "magick" {
		args = append(args, "convert")
	}

	// -density must come before input for correct SVG rendering.
	if density := opts.Get("imagemagick", "density"); density != "" {
		args = append(args, "-density", density)
	}

	args = append(args, in)

	if opts.Quality > 0 {
		args = append(args, "-quality", fmt.Sprintf("%d", opts.Quality))
	}
	if resize := opts.Get("imagemagick", "resize"); resize != "" {
		args = append(args, "-resize", resize)
	}
	if depth := opts.Get("imagemagick", "depth"); depth != "" {
		args = append(args, "-depth", depth)
	}
	if opts.StripMeta {
		args = append(args, "-strip")
	}

	args = append(args, opts.ExtraArgs...)
	args = append(args, out)

	if err := execx.Run(ctx, binary, args...); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	return nil
}
