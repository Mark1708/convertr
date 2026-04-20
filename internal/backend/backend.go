package backend

import (
	"context"
	"errors"
	"fmt"
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
