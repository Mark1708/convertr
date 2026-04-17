package runner

import (
	"git.mark1708.ru/me/convertr/internal/backend"
	"git.mark1708.ru/me/convertr/internal/router"
	"git.mark1708.ru/me/convertr/internal/sink"
	"git.mark1708.ru/me/convertr/internal/source"
)

// Job is a single conversion unit.
type Job struct {
	Source source.SourceFile
	Route  *router.Route
	Sink   *sink.Sink
	Opts   backend.Options
}
