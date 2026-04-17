package watch

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// addRecursive adds path and all subdirectories to fw.
// Non-directory paths and stat errors are silently skipped.
func addRecursive(fw *fsnotify.Watcher, root string) error {
	fi, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fw.Add(root)
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			return fw.Add(path)
		}
		return nil
	})
}
