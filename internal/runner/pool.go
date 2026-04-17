package runner

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

func executeParallel(ctx context.Context, jobs []Job, opts RunOpts, coll *collector) (*Summary, error) {
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	sem := semaphore.NewWeighted(int64(opts.Workers))
	var g errgroup.Group
	var mu sync.Mutex
	var firstErr error

	for _, j := range jobs {
		if err := sem.Acquire(cancelCtx, 1); err != nil {
			// Context cancelled (stop policy triggered or parent cancelled).
			break
		}
		g.Go(func() error {
			defer sem.Release(1)
			outPath, err := executeJob(cancelCtx, j, opts)
			if err != nil {
				coll.incFail()
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				if opts.OnError == ErrorPolicyStop {
					cancel()
				}
				return nil // don't propagate; we handle errors via firstErr
			}
			if outPath == "" {
				coll.incSkip()
			} else {
				coll.incOK()
			}
			return nil
		})
	}

	g.Wait() //nolint:errcheck // goroutines never return non-nil errors

	sum := coll.summary()
	return &sum, firstErr
}
