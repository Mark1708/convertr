package slogx

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var once sync.Once

// Init initialises the global slog logger.
// Level is read from $CONVERTR_LOG (debug|info|warn|error).
// When unset, logging is silenced.
func Init() {
	once.Do(func() {
		raw := strings.ToLower(strings.TrimSpace(os.Getenv("CONVERTR_LOG")))
		if raw == "" {
			slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
				Level: slog.Level(100),
			})))
			return
		}
		level := slog.LevelInfo
		switch raw {
		case "debug":
			level = slog.LevelDebug
		case "warn", "warning":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:     level,
			AddSource: level == slog.LevelDebug,
		})))
	})
}

// SetLevel overrides the log level at runtime (e.g. from --verbose flag).
func SetLevel(level slog.Level) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	})))
}

// SetJSON switches to JSON output (for --json / CI=1).
func SetJSON(level slog.Level) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
}

// LevelFromVerbosity maps -v count to a slog.Level.
func LevelFromVerbosity(v int) slog.Level {
	switch {
	case v >= 3:
		return slog.LevelDebug
	case v == 2:
		return slog.LevelDebug
	case v == 1:
		return slog.LevelInfo
	default:
		return slog.LevelWarn
	}
}

// EnsureDir creates dir if it does not exist.
func EnsureDir(dir string) error {
	return os.MkdirAll(filepath.Dir(dir), 0o750)
}
