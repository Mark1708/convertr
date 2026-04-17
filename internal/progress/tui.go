package progress

import (
	"fmt"
	"io"
	"strings"
)

// TUI renders a simple in-place progress bar for TTY output.
// It falls back to plain-text behaviour for non-TTY destinations.
type TUI struct {
	w     io.Writer
	total int
	done  int
	isTTY bool
}

// NewTUI creates a TUI reporter. Pass isTTY=true when w is a terminal.
func NewTUI(w io.Writer, isTTY bool) *TUI {
	return &TUI{w: w, isTTY: isTTY}
}

func (t *TUI) Start(n int) {
	t.total = n
	t.done = 0
	if t.isTTY {
		t.render()
	}
}

func (t *TUI) Update(done, total int, name string, err error) {
	t.done = done
	t.total = total
	if t.isTTY {
		// Clear current line, redraw bar.
		fmt.Fprint(t.w, "\r\033[K")
		t.render()
	} else {
		if err != nil {
			fmt.Fprintf(t.w, "[%d/%d] ERROR  %s: %v\n", done, total, name, err)
		} else {
			fmt.Fprintf(t.w, "[%d/%d] OK     %s\n", done, total, name)
		}
	}
}

func (t *TUI) Done() {
	if t.isTTY {
		fmt.Fprint(t.w, "\r\033[K") // clear bar line
		fmt.Fprintf(t.w, "Done: %d/%d\n", t.done, t.total)
	}
}

func (t *TUI) render() {
	pct := 0
	if t.total > 0 {
		pct = t.done * 100 / t.total
	}
	width := 30
	filled := width * pct / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	fmt.Fprintf(t.w, "[%s] %3d%% (%d/%d)", bar, pct, t.done, t.total)
}
