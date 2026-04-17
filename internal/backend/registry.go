package backend

import (
	"fmt"
	"sync"
)

var (
	mu       sync.RWMutex
	backends []Backend
)

// Register adds a backend to the global registry.
// Backends should call Register in their init() or via a constructor.
func Register(b Backend) {
	mu.Lock()
	defer mu.Unlock()
	backends = append(backends, b)
}

// All returns a snapshot of all registered backends.
func All() []Backend {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Backend, len(backends))
	copy(out, backends)
	return out
}

// Resolve returns the first backend that can convert from→to.
// Returns ErrNoRoute if none found.
func Resolve(from, to string) (Backend, error) {
	mu.RLock()
	defer mu.RUnlock()
	for _, b := range backends {
		for _, c := range b.Capabilities() {
			if c.From == from && c.To == to {
				return b, nil
			}
		}
	}
	return nil, fmt.Errorf("%w: %s → %s", ErrNoRoute, from, to)
}

// AllCapabilities returns every declared capability across all backends.
func AllCapabilities() []Capability {
	mu.RLock()
	defer mu.RUnlock()
	var caps []Capability
	for _, b := range backends {
		caps = append(caps, b.Capabilities()...)
	}
	return caps
}
