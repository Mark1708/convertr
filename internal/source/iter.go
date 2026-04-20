package source

import (
	"io"
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"strings"

	"github.com/Mark1708/convertr/internal/formats"
)

// FileSource yields a single SourceFile for path, detecting its format.
func FileSource(path string) iter.Seq2[SourceFile, error] {
	return func(yield func(SourceFile, error) bool) {
		sf, err := statFile(path, "")
		yield(sf, err)
	}
}

// FileSourceWithFormat yields a single SourceFile using the given format ID.
func FileSourceWithFormat(path, formatID string) iter.Seq2[SourceFile, error] {
	return func(yield func(SourceFile, error) bool) {
		sf, err := statFile(path, formatID)
		yield(sf, err)
	}
}

// GlobSource yields SourceFiles for all paths matching pattern.
func GlobSource(pattern string) iter.Seq2[SourceFile, error] {
	return func(yield func(SourceFile, error) bool) {
		paths, err := filepath.Glob(pattern)
		if err != nil {
			yield(SourceFile{}, err)
			return
		}
		for _, p := range paths {
			sf, e := statFile(p, "")
			if !yield(sf, e) {
				return
			}
		}
	}
}

// DirSource yields SourceFiles for all files under root, applying opts.
// If opts.MaxDepth == 1, only direct children are yielded (non-recursive).
func DirSource(root string, opts DirOpts) iter.Seq2[SourceFile, error] {
	return func(yield func(SourceFile, error) bool) {
		stop := false
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error { //nolint:errcheck
			if stop {
				return fs.SkipAll
			}
			if err != nil {
				if !yield(SourceFile{}, err) {
					stop = true
					return fs.SkipAll
				}
				return nil
			}
			if d.IsDir() {
				if path == root {
					return nil
				}
				if opts.MaxDepth > 0 {
					rel, _ := filepath.Rel(root, path)
					if depthOf(rel) >= opts.MaxDepth {
						return fs.SkipDir
					}
				}
				return nil
			}
			base := filepath.Base(path)
			if !matchGlobs(base, opts.IncludeGlobs, true) {
				return nil
			}
			if matchGlobs(base, opts.ExcludeGlobs, false) {
				return nil
			}
			sf, e := statFile(path, "")
			if e != nil {
				if !yield(SourceFile{}, e) {
					stop = true
					return fs.SkipAll
				}
				return nil
			}
			if opts.MinSize > 0 && sf.Size < opts.MinSize {
				return nil
			}
			if opts.MaxSize > 0 && sf.Size > opts.MaxSize {
				return nil
			}
			if !yield(sf, nil) {
				stop = true
				return fs.SkipAll
			}
			return nil
		})
	}
}

// StdinSource reads stdin into a temporary file and yields it with the given format.
// The caller must call sf.Close() after use to remove the temp file.
func StdinSource(formatID string) iter.Seq2[SourceFile, error] {
	return func(yield func(SourceFile, error) bool) {
		tmp, err := os.CreateTemp("", "convertr-stdin-*")
		if err != nil {
			yield(SourceFile{}, err)
			return
		}
		tmpName := tmp.Name()
		cleanup := func() { os.Remove(tmpName) }

		if _, err := io.Copy(tmp, os.Stdin); err != nil {
			tmp.Close()
			cleanup()
			yield(SourceFile{}, err)
			return
		}
		if err := tmp.Close(); err != nil {
			cleanup()
			yield(SourceFile{}, err)
			return
		}
		fi, err := os.Stat(tmpName)
		if err != nil {
			cleanup()
			yield(SourceFile{}, err)
			return
		}
		sf := SourceFile{
			Path:    tmpName,
			Format:  formatID,
			Size:    fi.Size(),
			ModTime: fi.ModTime(),
			cleanup: cleanup,
		}
		yield(sf, nil)
	}
}

// Chain concatenates multiple sources into one sequential sequence.
func Chain(sources ...iter.Seq2[SourceFile, error]) iter.Seq2[SourceFile, error] {
	return func(yield func(SourceFile, error) bool) {
		for _, src := range sources {
			for sf, err := range src {
				if !yield(sf, err) {
					return
				}
			}
		}
	}
}

func statFile(path, fmtID string) (SourceFile, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return SourceFile{}, err
	}
	if fmtID == "" {
		f, e := formats.DetectFile(path)
		if e != nil {
			return SourceFile{}, e
		}
		if f != nil {
			fmtID = f.ID
		}
	}
	return SourceFile{
		Path:    path,
		Format:  fmtID,
		Size:    fi.Size(),
		ModTime: fi.ModTime(),
	}, nil
}

func depthOf(rel string) int {
	return len(strings.Split(filepath.ToSlash(rel), "/"))
}

func matchGlobs(name string, globs []string, defaultResult bool) bool {
	if len(globs) == 0 {
		return defaultResult
	}
	for _, g := range globs {
		if ok, _ := filepath.Match(g, name); ok {
			return true
		}
	}
	return false
}
