package backend

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/Mark1708/convertr/internal/fonts"
)

var (
	ErrNoBinary    = errors.New("binary not found")
	ErrNoRoute     = errors.New("no conversion route")
	ErrConvertFail = errors.New("conversion failed")
)

// Capability declares a single From→To conversion that a backend can perform.
type Capability struct {
	From    string
	To      string
	Cost    int // 1=fast, 5=slow/lossy
	Quality int // 0-100
	Lossy   bool
}

// Options carries conversion parameters passed from CLI / config.
type Options struct {
	Quality   int
	StripMeta bool
	Workers   int
	ExtraArgs []string
	Named     map[string]string // "backend.key" → value
	Fonts     fonts.Config      // font preferences for PDF-producing backends
}

// Get returns the named option for a specific backend ("pandoc.wrap" etc.).
func (o Options) Get(backend, key string) string {
	if o.Named == nil {
		return ""
	}
	return o.Named[backend+"."+key]
}

// Backend is the interface every converter must satisfy.
type Backend interface {
	Name() string
	BinaryName() string
	Capabilities() []Capability
	Convert(ctx context.Context, in, out string, opts Options) error
}

// Availabler is an optional interface a Backend may implement to declare,
// per-capability, whether the underlying tooling is currently installed.
// Backends with multiple edge-specific binaries (e.g. csvkit uses
// in2csv/xlsx2csv for xlsx→csv but csvjson for csv→json) should implement
// this so the router can filter out edges that cannot actually run.
//
// If a backend does not implement Availabler, IsAvailable falls back to
// checking whether BinaryName() is present in $PATH.
type Availabler interface {
	IsAvailable(from, to string) bool
}

// IsAvailable reports whether `b` can currently perform the from→to
// capability given the host's installed binaries. The router uses this to
// exclude unavailable edges at graph-build time so conversions never
// attempt to invoke a missing binary.
func IsAvailable(b Backend, from, to string) bool {
	if a, ok := b.(Availabler); ok {
		return a.IsAvailable(from, to)
	}
	if _, err := exec.LookPath(b.BinaryName()); err == nil {
		return true
	}
	return false
}

// ConvertError wraps a backend error with routing context.
type ConvertError struct {
	BackendName string
	From        string
	To          string
	Err         error
}

func (e *ConvertError) Error() string {
	return fmt.Sprintf("%s: convert %s → %s: %v", e.BackendName, e.From, e.To, e.Err)
}

func (e *ConvertError) Unwrap() error { return e.Err }

// Is makes ConvertError satisfy errors.Is(err, ErrConvertFail) for retry logic.
func (e *ConvertError) Is(target error) bool { return target == ErrConvertFail }

// Wrap creates a ConvertError.
func Wrap(name, from, to string, err error) error {
	return &ConvertError{BackendName: name, From: from, To: to, Err: err}
}
