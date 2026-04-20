package progress

import (
	"fmt"
	"io"
	"path/filepath"
)

// Plain writes one line per job to w (suitable for non-TTY or --quiet mode).
type Plain struct {
	w     io.Writer
	total int
}

// NewPlain creates a plain-text reporter writing to w.
func NewPlain(w io.Writer) *Plain { return &Plain{w: w} }

func (p *Plain) Start(n int) { p.total = n }
func (p *Plain) Done()       {}

func (p *Plain) Update(done, total int, name string, err error) {
	base := filepath.Base(name)
	if err != nil {
		fmt.Fprintf(p.w, "[%d/%d] ERROR  %s: %v\n", done, total, base, err)
	} else {
		fmt.Fprintf(p.w, "[%d/%d] OK     %s\n", done, total, base)
	}
}
