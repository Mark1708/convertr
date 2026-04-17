// Package watch provides file system watching with debouncing for convertr watch mode.
package watch

import (
	"context"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// DeletePolicy controls what happens to the output file when the source is deleted.
type DeletePolicy int

const (
	DeleteKeep    DeletePolicy = iota // leave output file as-is
	DeleteRemove                      // remove corresponding output file
	DeleteArchive                     // move output file to .archive/ subdirectory
)

// ParseDeletePolicy parses a string delete policy name.
func ParseDeletePolicy(s string) (DeletePolicy, error) {
	switch s {
	case "keep", "":
		return DeleteKeep, nil
	case "remove":
		return DeleteRemove, nil
	case "archive":
		return DeleteArchive, nil
	default:
		return 0, &unknownPolicyError{s}
	}
}

type unknownPolicyError struct{ s string }

func (e *unknownPolicyError) Error() string {
	return "unknown delete policy " + e.s + ": use keep|remove|archive"
}

// Event is emitted when a file should be (re)converted.
type Event struct {
	Path string // absolute path of changed file
}

// Config holds watcher options.
type Config struct {
	Debounce     time.Duration // how long to wait after last event before emitting
	DeletePolicy DeletePolicy
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Debounce:     300 * time.Millisecond,
		DeletePolicy: DeleteKeep,
	}
}

// Watcher watches a directory tree and emits debounced Events on changes.
type Watcher struct {
	fw     *fsnotify.Watcher
	cfg    Config
	events chan Event
}

// New creates a Watcher that emits to the returned channel.
// The caller must call Close to release resources.
func New(cfg Config) (*Watcher, <-chan Event, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, err
	}
	ch := make(chan Event, 64)
	w := &Watcher{fw: fw, cfg: cfg, events: ch}
	return w, ch, nil
}

// Add recursively watches root and all subdirectories.
func (w *Watcher) Add(root string) error {
	return addRecursive(w.fw, root)
}

// Run processes fsnotify events until ctx is cancelled.
// It emits debounced Events on the channel returned by New.
func (w *Watcher) Run(ctx context.Context) {
	defer close(w.events)

	pending := map[string]*time.Timer{}

	emit := func(path string) func() {
		return func() {
			delete(pending, path)
			select {
			case w.events <- Event{Path: path}:
			case <-ctx.Done():
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-w.fw.Events:
			if !ok {
				return
			}
			path := filepath.Clean(ev.Name)
			switch {
			case ev.Has(fsnotify.Create):
				// Recursively add new directories.
				if err := addRecursive(w.fw, path); err == nil {
					slog.Debug("watch: added", "path", path)
				}
				fallthrough
			case ev.Has(fsnotify.Write):
				// Debounce: reset timer on every new event for the same file.
				if t, ok := pending[path]; ok {
					t.Stop()
				}
				pending[path] = time.AfterFunc(w.cfg.Debounce, emit(path))
			case ev.Has(fsnotify.Remove), ev.Has(fsnotify.Rename):
				if t, ok := pending[path]; ok {
					t.Stop()
					delete(pending, path)
				}
				// Deletion events are handled by the caller if needed.
			}
		case err, ok := <-w.fw.Errors:
			if !ok {
				return
			}
			slog.Warn("watch: fsnotify error", "err", err)
		}
	}
}

// Close stops the underlying fsnotify watcher.
func (w *Watcher) Close() error {
	return w.fw.Close()
}
