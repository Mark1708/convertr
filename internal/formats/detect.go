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
// For generic text/plain MIME type, extension-based detection takes priority
// because many text-based formats (md, rst, toml, yaml, etc.) share the same MIME.
func DetectReader(name string, r io.Reader) (*Format, error) {
	ext := strings.ToLower(filepath.Ext(name))

	mime, err := mimetype.DetectReader(r)
	if err == nil {
		mimeStr := mime.String()
		// For unambiguous binary MIME types, trust the magic bytes.
		if !isGenericTextMIME(mimeStr) {
			if f := byMime(mimeStr); f != nil {
				return f, nil
			}
		}
	}
	// For text-based files (or unknown MIME), prefer extension.
	if f := ByExtension(ext); f != nil {
		return f, nil
	}
	// Final fallback: MIME (catches text/plain without a recognised extension).
	if err == nil {
		if f := byMime(mime.String()); f != nil {
			return f, nil
		}
	}
	return nil, nil // unknown, not an error
}

// isGenericTextMIME returns true for MIME types that convey no format specifics.
func isGenericTextMIME(m string) bool {
	if i := strings.Index(m, ";"); i >= 0 {
		m = strings.TrimSpace(m[:i])
	}
	return m == "text/plain" || m == "application/octet-stream"
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
