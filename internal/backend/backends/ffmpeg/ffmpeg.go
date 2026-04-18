// Package ffmpeg provides audio/video conversion via the ffmpeg binary.
package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"git.mark1708.ru/me/convertr/internal/backend"
	"git.mark1708.ru/me/convertr/internal/backend/execx"
)

func init() {
	backend.Register(Backend{})
}

// Backend wraps the ffmpeg binary.
type Backend struct{}

func (Backend) Name() string       { return "ffmpeg" }
func (Backend) BinaryName() string { return "ffmpeg" }

func (Backend) Capabilities() []backend.Capability {
	caps := []backend.Capability{}

	// Video formats — bidirectional
	videoFormats := []string{"mp4", "mkv", "webm", "mov", "avi"}
	for _, from := range videoFormats {
		for _, to := range videoFormats {
			if from != to {
				caps = append(caps, backend.Capability{From: from, To: to, Cost: 3, Quality: 85})
			}
		}
	}

	// Audio formats — bidirectional
	audioFormats := []string{"mp3", "flac", "aac", "ogg", "wav", "m4a", "opus"}
	for _, from := range audioFormats {
		for _, to := range audioFormats {
			if from != to {
				caps = append(caps, backend.Capability{From: from, To: to, Cost: 2, Quality: 88})
			}
		}
	}

	// Video → animated GIF (special palette filter)
	for _, v := range []string{"mp4", "mkv", "webm", "mov"} {
		caps = append(caps, backend.Capability{From: v, To: "gif", Cost: 3, Quality: 75, Lossy: true})
	}

	// Video → audio extraction
	for _, v := range videoFormats {
		for _, a := range audioFormats {
			caps = append(caps, backend.Capability{From: v, To: a, Cost: 2, Quality: 90})
		}
	}

	return caps
}

func (b Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	ext := filepath.Ext(out)
	if ext == ".gif" {
		return convertToGIF(ctx, b, in, out, opts)
	}
	args := buildArgs(in, out, opts)
	if err := execx.Run(ctx, "ffmpeg", args...); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	return nil
}

func buildArgs(in, out string, opts backend.Options) []string {
	args := []string{"-y", "-hide_banner", "-loglevel", "error", "-i", in}

	if opts.Workers > 0 {
		args = append(args, "-threads", strconv.Itoa(opts.Workers))
	}

	if opts.Quality > 0 {
		ext := filepath.Ext(out)
		switch ext {
		case ".mp4", ".mkv", ".webm", ".mov":
			args = append(args, "-crf", fmt.Sprintf("%d", 51-opts.Quality/2))
		case ".mp3", ".aac", ".ogg", ".opus", ".m4a":
			args = append(args, "-b:a", fmt.Sprintf("%dk", opts.Quality*3))
		}
	}

	if codec := opts.Get("ffmpeg", "video_codec"); codec != "" {
		args = append(args, "-c:v", codec)
	}
	if codec := opts.Get("ffmpeg", "audio_codec"); codec != "" {
		args = append(args, "-c:a", codec)
	}
	if fps := opts.Get("ffmpeg", "fps"); fps != "" {
		args = append(args, "-r", fps)
	}
	if rate := opts.Get("ffmpeg", "audio_rate"); rate != "" {
		args = append(args, "-ar", rate)
	}

	if opts.StripMeta {
		args = append(args, "-map_metadata", "-1")
	}

	args = append(args, opts.ExtraArgs...)
	args = append(args, out)
	return args
}

// convertToGIF uses a two-pass palette filter for high-quality GIF output.
func convertToGIF(ctx context.Context, b Backend, in, out string, opts backend.Options) error {
	palette := filepath.Join(os.TempDir(), fmt.Sprintf("convertr-palette-%d.png", os.Getpid()))
	defer os.Remove(palette)

	// Build palette filter: optional fps and scale.
	paletteFilter := buildGIFFilter(opts)

	pass1 := []string{
		"-y", "-hide_banner", "-loglevel", "error",
		"-i", in,
		"-vf", paletteFilter + ",palettegen",
		palette,
	}
	if err := execx.Run(ctx, "ffmpeg", pass1...); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}

	lavfi := paletteFilter + "[x];[x][1:v]paletteuse"
	pass2 := []string{
		"-y", "-hide_banner", "-loglevel", "error",
		"-i", in, "-i", palette,
		"-lavfi", lavfi,
	}

	if opts.StripMeta {
		pass2 = append(pass2, "-map_metadata", "-1")
	}
	pass2 = append(pass2, opts.ExtraArgs...)
	pass2 = append(pass2, out)
	if err := execx.Run(ctx, "ffmpeg", pass2...); err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	return nil
}

// buildGIFFilter constructs the vf filter chain for the palette pass.
func buildGIFFilter(opts backend.Options) string {
	var parts []string
	if fps := opts.Get("ffmpeg", "gif_fps"); fps != "" {
		parts = append(parts, "fps="+fps)
	}
	if scale := opts.Get("ffmpeg", "gif_scale"); scale != "" {
		parts = append(parts, "scale="+scale+":-1:flags=lanczos")
	}
	if len(parts) == 0 {
		return "fps=10,scale=320:-1:flags=lanczos"
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ","
		}
		result += p
	}
	return result
}
