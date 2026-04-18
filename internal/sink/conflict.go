package sink

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"git.mark1708.ru/me/convertr/internal/i18n"
)

// ConflictPolicy controls what happens when the output file already exists.
type ConflictPolicy int

const (
	ConflictOverwrite ConflictPolicy = iota
	ConflictSkip
	ConflictRename
	ConflictError
)

// ParseConflictPolicy parses a string policy name.
func ParseConflictPolicy(s string) (ConflictPolicy, error) {
	switch strings.ToLower(s) {
	case "overwrite", "":
		return ConflictOverwrite, nil
	case "skip":
		return ConflictSkip, nil
	case "rename":
		return ConflictRename, nil
	case "error":
		return ConflictError, nil
	default:
		return 0, errors.New(i18n.Tf("sink.unknown_conflict_policy", map[string]any{"Policy": s}))
	}
}

// Action is the resolved action after applying a conflict policy.
type Action int

const (
	ActionWrite Action = iota
	ActionSkip
)

// Resolve returns the final output path and action given a potential conflict.
// If the file does not exist, it always returns (path, ActionWrite, nil).
func Resolve(path string, policy ConflictPolicy) (string, Action, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path, ActionWrite, nil
	}
	switch policy {
	case ConflictOverwrite:
		return path, ActionWrite, nil
	case ConflictSkip:
		return path, ActionSkip, nil
	case ConflictRename:
		return uniquePath(path), ActionWrite, nil
	case ConflictError:
		return "", 0, errors.New(i18n.Tf("sink.file_exists", map[string]any{"Path": path}))
	default:
		return path, ActionWrite, nil
	}
}

func uniquePath(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
