// Package csvkit provides CSV/XLSX/JSON conversion via csvkit or xlsx2csv.
package csvkit

import (
	"context"
	"os"
	"os/exec"

	"git.mark1708.ru/me/convertr/internal/backend"
	"git.mark1708.ru/me/convertr/internal/backend/execx"
)

func init() {
	backend.Register(Backend{})
}

// Backend wraps csvkit (in2csv/csvjson) with xlsx2csv as fallback.
type Backend struct{}

func (Backend) Name() string       { return "csvkit" }
func (Backend) BinaryName() string { return "in2csv" }

func (Backend) Capabilities() []backend.Capability {
	return []backend.Capability{
		{From: "xlsx", To: "csv", Cost: 1, Quality: 95},
		{From: "csv", To: "json", Cost: 1, Quality: 95},
		{From: "xlsx", To: "json", Cost: 2, Quality: 90},
	}
}

func (b Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	inExt := extOf(in)
	outExt := extOf(out)

	switch {
	case inExt == ".xlsx" && outExt == ".csv":
		return b.xlsxToCSV(ctx, in, out, opts)
	case inExt == ".csv" && outExt == ".json":
		return b.csvToJSON(ctx, in, out, opts)
	case inExt == ".xlsx" && outExt == ".json":
		return b.xlsxToJSON(ctx, in, out, opts)
	default:
		return backend.Wrap(b.Name(), in, out, &unsupportedError{inExt, outExt})
	}
}

func (b Backend) xlsxToCSV(ctx context.Context, in, out string, opts backend.Options) error {
	// Prefer in2csv (csvkit); fall back to xlsx2csv.
	if _, err := exec.LookPath("in2csv"); err == nil {
		args := append([]string{in}, opts.ExtraArgs...)
		data, err := execx.Output(ctx, "in2csv", args...)
		if err != nil {
			return backend.Wrap(b.Name(), in, out, err)
		}
		return os.WriteFile(out, data, 0o644)
	}
	if _, err := exec.LookPath("xlsx2csv"); err == nil {
		args := append([]string{in, out}, opts.ExtraArgs...)
		if err := execx.Run(ctx, "xlsx2csv", args...); err != nil {
			return backend.Wrap(b.Name(), in, out, err)
		}
		return nil
	}
	return backend.Wrap(b.Name(), in, out, errNoBinary("in2csv or xlsx2csv"))
}

func (b Backend) csvToJSON(ctx context.Context, in, out string, opts backend.Options) error {
	if _, err := exec.LookPath("csvjson"); err != nil {
		return backend.Wrap(b.Name(), in, out, errNoBinary("csvjson"))
	}
	args := append([]string{in}, opts.ExtraArgs...)
	data, err := execx.Output(ctx, "csvjson", args...)
	if err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	return os.WriteFile(out, data, 0o644)
}

func (b Backend) xlsxToJSON(ctx context.Context, in, out string, opts backend.Options) error {
	tmp, err := os.CreateTemp("", "convertr-csvkit-*.csv")
	if err != nil {
		return backend.Wrap(b.Name(), in, out, err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	if err := b.xlsxToCSV(ctx, in, tmp.Name(), opts); err != nil {
		return err
	}
	return b.csvToJSON(ctx, tmp.Name(), out, backend.Options{})
}

func extOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}

type unsupportedError struct{ from, to string }

func (e *unsupportedError) Error() string {
	return "unsupported conversion: " + e.from + " → " + e.to
}

type errNoBinary string

func (e errNoBinary) Error() string { return string(e) + " not found in PATH" }
