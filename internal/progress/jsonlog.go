package progress

import (
	"encoding/json"
	"io"
	"time"
)

// JSONLog emits one JSON object per event (suitable for --json / CI=1).
type JSONLog struct {
	w io.Writer
}

// NewJSONLog creates a JSON-line reporter writing to w.
func NewJSONLog(w io.Writer) *JSONLog { return &JSONLog{w: w} }

func (j *JSONLog) Start(n int) {
	j.emit(map[string]any{"event": "start", "total": n, "ts": ts()})
}

func (j *JSONLog) Update(done, total int, name string, err error) {
	rec := map[string]any{
		"event": "convert",
		"file":  name,
		"done":  done,
		"total": total,
		"ts":    ts(),
	}
	if err != nil {
		rec["status"] = "error"
		rec["error"] = err.Error()
	} else {
		rec["status"] = "ok"
	}
	j.emit(rec)
}

func (j *JSONLog) Done() {
	j.emit(map[string]any{"event": "done", "ts": ts()})
}

func (j *JSONLog) emit(v any) {
	b, _ := json.Marshal(v)
	b = append(b, '\n')
	j.w.Write(b) //nolint:errcheck
}

func ts() string { return time.Now().UTC().Format(time.RFC3339) }
