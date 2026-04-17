package watch

import (
	"fmt"
	"os"
	"path/filepath"
)

// HandleDeletion applies the delete policy for srcPath whose output lives at outPath.
func HandleDeletion(outPath string, policy DeletePolicy) error {
	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		return nil // output doesn't exist; nothing to do
	}
	switch policy {
	case DeleteKeep:
		return nil
	case DeleteRemove:
		return os.Remove(outPath)
	case DeleteArchive:
		archiveDir := filepath.Join(filepath.Dir(outPath), ".archive")
		if err := os.MkdirAll(archiveDir, 0o755); err != nil {
			return fmt.Errorf("create archive dir: %w", err)
		}
		dest := filepath.Join(archiveDir, filepath.Base(outPath))
		return os.Rename(outPath, dest)
	default:
		return nil
	}
}
