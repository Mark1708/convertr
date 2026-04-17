package source

import "time"

// SourceFile represents a resolved input file ready for conversion.
type SourceFile struct {
	Path    string
	Format  string // canonical format ID, e.g. "docx", "mp4"
	Size    int64
	ModTime time.Time
	cleanup func() // optional cleanup after use (e.g. remove stdin temp file)
}

// Close runs the cleanup function if set. Safe to call on zero value.
func (sf SourceFile) Close() {
	if sf.cleanup != nil {
		sf.cleanup()
	}
}
