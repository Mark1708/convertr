package execx

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes binary with args, returning a formatted error on non-zero exit.
func Run(ctx context.Context, binary string, args ...string) error {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return fmt.Errorf("%s: %w", binary, err)
		}
		return fmt.Errorf("%s: %s", binary, msg)
	}
	return nil
}

// Output executes binary and returns stdout bytes.
func Output(ctx context.Context, binary string, args ...string) ([]byte, error) {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return nil, fmt.Errorf("%s: %w", binary, err)
		}
		return nil, fmt.Errorf("%s: %s", binary, msg)
	}
	return out, nil
}

// LookPath wraps exec.LookPath with a descriptive error.
func LookPath(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, name)
	}
	return path, nil
}

// Version runs binary with versionArgs and returns first line of output.
// Returns "" on any error.
func Version(ctx context.Context, binary string, args ...string) string {
	out, err := Output(ctx, binary, args...)
	if err != nil {
		return ""
	}
	line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	return line
}
