// Package plugin discovers and registers external convertr-* executables as backends.
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Mark1708/convertr/internal/backend"
	"github.com/Mark1708/convertr/internal/backend/execx"
	pluginpkg "github.com/Mark1708/convertr/pkg/plugin"
)

func init() {
	DiscoverAndRegister()
}

// DiscoverAndRegister finds all convertr-* executables in PATH and registers them.
func DiscoverAndRegister() {
	for _, name := range discover() {
		caps, err := probe(name)
		if err != nil || len(caps) == 0 {
			continue
		}
		backend.Register(&pluginBackend{name: name, caps: caps})
	}
}

// discover returns names of all convertr-* executables found in PATH.
func discover() []string {
	seen := map[string]bool{}
	var found []string
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasPrefix(name, "convertr-") {
				continue
			}
			full := filepath.Join(dir, name)
			if info, err := os.Stat(full); err == nil && info.Mode()&0o111 != 0 && !seen[name] {
				seen[name] = true
				found = append(found, name)
			}
		}
	}
	return found
}

// ProbePlugin calls `convertr-NAME capabilities` and returns parsed capabilities.
func ProbePlugin(execName string) ([]pluginpkg.Capability, error) { return probe(execName) }

func probe(execName string) ([]pluginpkg.Capability, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, execName, "capabilities").Output()
	if err != nil {
		return nil, err
	}
	var caps []pluginpkg.Capability
	if err := json.Unmarshal(out, &caps); err != nil {
		return nil, fmt.Errorf("plugin %s: invalid capabilities JSON: %w", execName, err)
	}
	return caps, nil
}

// pluginBackend wraps an external executable as a backend.Backend.
type pluginBackend struct {
	name string
	caps []pluginpkg.Capability
}

func (p *pluginBackend) Name() string       { return p.name }
func (p *pluginBackend) BinaryName() string { return p.name }

func (p *pluginBackend) Capabilities() []backend.Capability {
	out := make([]backend.Capability, len(p.caps))
	for i, c := range p.caps {
		cost := c.Cost
		if cost <= 0 {
			cost = 5
		}
		out[i] = backend.Capability{From: c.From, To: c.To, Cost: cost}
	}
	return out
}

func (p *pluginBackend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
	fromExt := strings.TrimPrefix(filepath.Ext(in), ".")
	toExt := strings.TrimPrefix(filepath.Ext(out), ".")
	args := []string{"convert", "--from", fromExt, "--to", toExt, "--input", in, "--output", out}
	for k, v := range opts.Named {
		args = append(args, "--opt", k+"="+v)
	}
	return execx.Run(ctx, p.name, args...)
}
