// Package progress provides progress reporting for batch conversions.
package progress

// Reporter is the interface all progress implementations satisfy.
type Reporter interface {
	// Start signals that a batch of n jobs is beginning.
	Start(n int)
	// Update is called after each job completes (successfully or not).
	Update(done, total int, name string, err error)
	// Done signals that all jobs have finished.
	Done()
}

// Noop is a reporter that discards all events (used with --quiet).
type Noop struct{}

func (Noop) Start(_ int)                          {}
func (Noop) Update(_, _ int, _ string, _ error)   {}
func (Noop) Done()                                {}
