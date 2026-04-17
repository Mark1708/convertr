package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"git.mark1708.ru/me/convertr/internal/formats"
	"git.mark1708.ru/me/convertr/internal/progress"
	"git.mark1708.ru/me/convertr/internal/sink"
)

// ErrorPolicy controls what happens when a job fails.
type ErrorPolicy int

const (
	ErrorPolicySkip  ErrorPolicy = iota // record error and continue
	ErrorPolicyStop                     // abort all remaining jobs
	ErrorPolicyRetry                    // retry with backoff
)

// ParseErrorPolicy parses a string error policy name.
func ParseErrorPolicy(s string) (ErrorPolicy, error) {
	switch s {
	case "skip", "":
		return ErrorPolicySkip, nil
	case "stop":
		return ErrorPolicyStop, nil
	case "retry":
		return ErrorPolicyRetry, nil
	default:
		return 0, fmt.Errorf("unknown error policy %q: use skip|stop|retry", s)
	}
}

// RunOpts configures the runner.
type RunOpts struct {
	Workers    int
	OnError    ErrorPolicy
	OnConflict sink.ConflictPolicy
	Retry      RetryOpts
	DryRun     bool
	Reporter   progress.Reporter // nil → Noop
}

// Execute runs all jobs and returns a summary.
func Execute(ctx context.Context, jobs []Job, opts RunOpts) (*Summary, error) {
	coll := &collector{}
	if opts.Workers > 1 {
		return executeParallel(ctx, jobs, opts, coll)
	}
	return executeSerial(ctx, jobs, opts, coll)
}

func executeSerial(ctx context.Context, jobs []Job, opts RunOpts, coll *collector) (*Summary, error) {
	rep := opts.Reporter
	if rep == nil {
		rep = progress.Noop{}
	}
	rep.Start(len(jobs))
	var firstErr error
	for _, j := range jobs {
		if ctx.Err() != nil {
			break
		}
		outPath, err := executeJob(ctx, j, opts)
		if err != nil {
			coll.incFail()
			done := int(coll.ok.Load() + coll.fail.Load() + coll.skip.Load())
			rep.Update(done, len(jobs), j.Source.Path, err)
			if opts.OnError == ErrorPolicyStop {
				rep.Done()
				sum := coll.summary()
				return &sum, err
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if outPath == "" {
			coll.incSkip()
		} else {
			coll.incOK()
		}
		done := int(coll.ok.Load() + coll.fail.Load() + coll.skip.Load())
		rep.Update(done, len(jobs), j.Source.Path, nil)
	}
	rep.Done()
	sum := coll.summary()
	return &sum, firstErr
}

func executeJob(ctx context.Context, j Job, opts RunOpts) (string, error) {
	defer j.Source.Close() // clean up stdin temp files

	finalPath, action, err := resolveFinalPath(j, opts.OnConflict)
	if err != nil {
		return "", err
	}
	if action == sink.ActionSkip {
		return "", nil
	}

	if opts.DryRun {
		fmt.Fprintf(os.Stderr, "dry-run: %s → %s\n", j.Source.Path, finalPath)
		return finalPath, nil
	}

	run := func() error { return convertJob(ctx, j, finalPath) }
	if opts.OnError == ErrorPolicyRetry {
		return finalPath, withRetry(ctx, opts.Retry, run)
	}
	return finalPath, run()
}

func convertJob(ctx context.Context, j Job, finalPath string) error {
	if len(j.Route.Steps) == 0 {
		if finalPath == j.Source.Path {
			return nil
		}
		if finalPath == "-" {
			return copyToStdout(j.Source.Path)
		}
		return sink.AtomicWrite(finalPath, j.Source.Path)
	}

	tmpDir, err := os.MkdirTemp("", "convertr-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	current := j.Source.Path
	for i, step := range j.Route.Steps {
		ext := firstExt(step.To)
		next := filepath.Join(tmpDir, fmt.Sprintf("step%02d%s", i, ext))
		if err := step.Backend.Convert(ctx, current, next, j.Opts); err != nil {
			return err
		}
		current = next
	}

	if finalPath == "-" {
		return copyToStdout(current)
	}
	return sink.AtomicWrite(finalPath, current)
}

func resolveFinalPath(j Job, policy sink.ConflictPolicy) (string, sink.Action, error) {
	switch j.Sink.Type {
	case sink.SinkTypeStdout:
		return "-", sink.ActionWrite, nil
	case sink.SinkTypeFile:
		return sink.Resolve(j.Sink.Path, policy)
	case sink.SinkTypeDir:
		stem := stemOf(filepath.Base(j.Source.Path))
		ext := firstExt(j.Sink.Format)
		name := stem + ext
		if j.Sink.Template != "" {
			name = sink.ExpandTemplate(j.Sink.Template, stem, j.Sink.Format, 0)
		}
		outPath := filepath.Join(j.Sink.Path, name)
		return sink.Resolve(outPath, policy)
	default:
		return "", 0, fmt.Errorf("unknown sink type %d", j.Sink.Type)
	}
}

func firstExt(formatID string) string {
	f := formats.ByID(formatID)
	if f == nil || len(f.Extensions) == 0 {
		return "." + formatID
	}
	return f.Extensions[0]
}

func stemOf(name string) string {
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)]
}

func copyToStdout(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(os.Stdout, f)
	return err
}
