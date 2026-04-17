package sink

import (
	"io"
	"os"
	"path/filepath"
)

// AtomicWrite moves src to dst atomically.
// It tries os.Rename first (same-filesystem fast path).
// On cross-filesystem moves, it copies src to a temp file in dst's directory
// and then renames to avoid partial writes.
func AtomicWrite(dst, src string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// Cross-filesystem fallback: copy to temp, then rename.
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".convertr-tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	in, err := os.Open(src)
	if err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if _, err = io.Copy(tmp, in); err != nil {
		in.Close()
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	in.Close()
	if err = tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err = os.Rename(tmpName, dst); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
