package runner

import (
	"github.com/Mark1708/convertr/internal/backend"
	"github.com/Mark1708/convertr/internal/router"
	"github.com/Mark1708/convertr/internal/sink"
	"github.com/Mark1708/convertr/internal/source"
)

// Job is a single conversion unit.
type Job struct {
	Source source.SourceFile
	Route  *router.Route
	Sink   *sink.Sink
	Opts   backend.Options
}
