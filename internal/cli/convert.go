package cli

import (
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"git.mark1708.ru/me/convertr/internal/backend"
	"git.mark1708.ru/me/convertr/internal/formats"
	"git.mark1708.ru/me/convertr/internal/router"
	"git.mark1708.ru/me/convertr/internal/runner"
	"git.mark1708.ru/me/convertr/internal/sink"
	"git.mark1708.ru/me/convertr/internal/source"
)

type convertFlags struct {
	output     string
	toFormat   string
	fromFormat string
	dryRun     bool
	workers    int
	onError    string
	onConflict string
	recursive  bool
	mkdir      bool
}

// addConvertToRoot registers conversion flags on the root command and sets RunE.
// When the user invokes `convertr FILE -o OUT`, cobra falls through to root RunE
// because FILE does not match any registered subcommand.
func addConvertToRoot(root *cobra.Command) {
	var f convertFlags

	fl := root.Flags()
	fl.StringVarP(&f.output, "output", "o", "", "output file or directory (\"-\" for stdout)")
	fl.StringVar(&f.toFormat, "to", "", "target format ID (e.g. md, pdf, mp3)")
	fl.StringVar(&f.fromFormat, "from", "", "source format override")
	fl.BoolVar(&f.dryRun, "dry-run", false, "print planned conversions without executing")
	fl.IntVarP(&f.workers, "workers", "j", 1, "parallel workers")
	fl.StringVar(&f.onError, "on-error", "skip", "error policy: skip|stop|retry")
	fl.StringVar(&f.onConflict, "on-conflict", "overwrite", "conflict policy: overwrite|skip|rename|error")
	fl.BoolVarP(&f.recursive, "recursive", "r", false, "recurse into directories")
	fl.BoolVar(&f.mkdir, "mkdir", false, "create output directory if it does not exist")

	root.Args = cobra.ArbitraryArgs
	root.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return runConvert(cmd, args, f)
	}
}

func runConvert(cmd *cobra.Command, args []string, f convertFlags) error {
	toFormat, err := resolveTargetFormat(f.toFormat, f.output)
	if err != nil {
		return err
	}

	g := router.Build()

	conflictPolicy, err := sink.ParseConflictPolicy(f.onConflict)
	if err != nil {
		return err
	}

	// Build source iterators.
	var srcs []iter.Seq2[source.SourceFile, error]
	for _, arg := range args {
		if arg == "-" {
			if f.fromFormat == "" {
				return fmt.Errorf("stdin input (\"-\") requires --from FORMAT")
			}
			srcs = append(srcs, source.StdinSource(f.fromFormat))
		} else {
			srcs = append(srcs, resolveInputArg(arg, f.fromFormat, f.recursive))
		}
	}

	// Collect jobs.
	var jobs []runner.Job
	for sf, err := range source.Chain(srcs...) {
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		if sf.Format == "" {
			return fmt.Errorf("cannot determine format of %q: use --from FORMAT", sf.Path)
		}
		route, err := g.Find(sf.Format, toFormat)
		if err != nil {
			if errors.Is(err, backend.ErrNoRoute) {
				return fmt.Errorf("%w: %s → %s", backend.ErrNoRoute, sf.Format, toFormat)
			}
			return err
		}
		jobs = append(jobs, runner.Job{Source: sf, Route: route, Sink: nil})
	}

	if len(jobs) == 0 {
		return fmt.Errorf("no input files found")
	}

	// Resolve sink now that we know how many jobs there are.
	// With multiple jobs, an output path that looks like a file is treated
	// as a directory (same behaviour as cp -t).
	sk := sink.ResolveSink(f.output, toFormat)
	if sk.Type == sink.SinkTypeFile && len(jobs) > 1 {
		sk.Type = sink.SinkTypeDir
	}

	// When the output directory does not yet exist, create it (--mkdir) or
	// ask the user interactively.
	if sk.Type == sink.SinkTypeDir {
		if _, err := os.Stat(sk.Path); os.IsNotExist(err) {
			if f.mkdir {
				if err := os.MkdirAll(sk.Path, 0o755); err != nil {
					return fmt.Errorf("create output directory: %w", err)
				}
			} else if isInteractive(cmd) {
				fmt.Fprintf(cmd.ErrOrStderr(), "Output directory does not exist: %s\nCreate it? [y/N] ", sk.Path)
				var answer string
				fmt.Fscan(cmd.InOrStdin(), &answer)
				if answer != "y" && answer != "Y" {
					return fmt.Errorf("output directory %q does not exist; use --mkdir to create it automatically", sk.Path)
				}
				if err := os.MkdirAll(sk.Path, 0o755); err != nil {
					return fmt.Errorf("create output directory: %w", err)
				}
			} else {
				return fmt.Errorf("output directory %q does not exist; use --mkdir to create it automatically", sk.Path)
			}
		}
	}

	sk.Policy = conflictPolicy
	for i := range jobs {
		jobs[i].Sink = sk
	}

	errPolicy, err := runner.ParseErrorPolicy(f.onError)
	if err != nil {
		return err
	}

	opts := runner.RunOpts{
		Workers:    f.workers,
		OnError:    errPolicy,
		OnConflict: conflictPolicy,
		DryRun:     f.dryRun,
	}
	if errPolicy == runner.ErrorPolicyRetry {
		opts.Retry = runner.DefaultRetry
	}

	summary, err := runner.Execute(cmd.Context(), jobs, opts)
	if summary != nil && (summary.OK+summary.Fail+summary.Skip) > 1 {
		fmt.Fprintf(cmd.ErrOrStderr(), "done: %d ok, %d failed, %d skipped\n",
			summary.OK, summary.Fail, summary.Skip)
	}
	return err
}

func resolveTargetFormat(toFlag, output string) (string, error) {
	if toFlag != "" {
		return toFlag, nil
	}
	if output != "" && output != "-" {
		ext := strings.TrimPrefix(filepath.Ext(output), ".")
		if ext != "" {
			if f := formats.ByExtension(ext); f != nil {
				return f.ID, nil
			}
		}
	}
	return "", fmt.Errorf("cannot determine target format: use --to FORMAT or provide an output with a known extension")
}

func resolveInputArg(arg, fromFormat string, recursive bool) iter.Seq2[source.SourceFile, error] {
	if containsGlobChars(arg) {
		return source.GlobSource(arg)
	}
	fi, err := os.Stat(arg)
	if err != nil {
		// Might be a glob that matched nothing; yield the stat error.
		return source.GlobSource(arg)
	}
	if fi.IsDir() {
		opts := source.DirOpts{}
		if !recursive {
			opts.MaxDepth = 1
		}
		return source.DirSource(arg, opts)
	}
	if fromFormat != "" {
		return source.FileSourceWithFormat(arg, fromFormat)
	}
	return source.FileSource(arg)
}

func containsGlobChars(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// isInteractive returns true when the command's stdin is a terminal.
func isInteractive(cmd *cobra.Command) bool {
	f, ok := cmd.InOrStdin().(*os.File)
	if !ok {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
