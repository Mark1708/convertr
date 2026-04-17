package formats

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

// DetectFile detects the format of a file by magic bytes then extension.
func DetectFile(path string) (*Format, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return DetectReader(filepath.Base(path), f)
}

// DetectReader detects the format from a reader (reads first 512 bytes).
func DetectReader(name string, r io.Reader) (*Format, error) {
	mime, err := mimetype.DetectReader(r)
	if err == nil {
		if f := byMime(mime.String()); f != nil {
			return f, nil
		}
	}
	// fallback: extension
	ext := strings.ToLower(filepath.Ext(name))
	if f := ByExtension(ext); f != nil {
		return f, nil
	}
	return nil, nil // unknown, not an error
}

// MustID returns the canonical format ID for a file path.
// Returns "" if the format cannot be determined.
func MustID(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if f := ByExtension(ext); f != nil {
		return f.ID
	}
	return ""
}

func byMime(mime string) *Format {
	// strip parameters: "text/html; charset=utf-8" → "text/html"
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	for i := range known {
		for _, m := range known[i].MimeTypes {
			if m == mime {
				return &known[i]
			}
		}
	}
	return nil
}
