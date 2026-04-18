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
	"git.mark1708.ru/me/convertr/internal/i18n"
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
	named      []string
	stripMeta  bool
}

// addConvertToRoot registers conversion flags on the root command and sets RunE.
// When the user invokes `convertr FILE -o OUT`, cobra falls through to root RunE
// because FILE does not match any registered subcommand.
func addConvertToRoot(root *cobra.Command) {
	var f convertFlags

	fl := root.Flags()
	fl.StringVarP(&f.output, "output", "o", "", i18n.T("cli.flag.output"))
	fl.StringVar(&f.toFormat, "to", "", i18n.T("cli.flag.to"))
	fl.StringVar(&f.fromFormat, "from", "", i18n.T("cli.flag.from"))
	fl.BoolVar(&f.dryRun, "dry-run", false, i18n.T("cli.flag.dry_run"))
	fl.IntVarP(&f.workers, "workers", "j", 1, i18n.T("cli.flag.workers"))
	fl.StringVar(&f.onError, "on-error", "skip", i18n.T("cli.flag.on_error"))
	fl.StringVar(&f.onConflict, "on-conflict", "overwrite", i18n.T("cli.flag.on_conflict"))
	fl.BoolVarP(&f.recursive, "recursive", "r", false, i18n.T("cli.flag.recursive"))
	fl.BoolVar(&f.mkdir, "mkdir", false, i18n.T("cli.flag.mkdir"))
	fl.StringArrayVar(&f.named, "named", nil, i18n.T("cli.flag.named"))
	fl.BoolVar(&f.stripMeta, "strip-meta", false, i18n.T("cli.flag.strip_meta"))

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

	namedMap := parseNamed(f.named)

	// Build source iterators.
	var srcs []iter.Seq2[source.SourceFile, error]
	for _, arg := range args {
		if arg == "-" {
			if f.fromFormat == "" {
				return errors.New(i18n.T("error.stdin_needs_from"))
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
			return errors.New(i18n.Tf("error.unknown_format", map[string]any{"Path": sf.Path}))
		}
		route, err := g.Find(sf.Format, toFormat)
		if err != nil {
			if errors.Is(err, backend.ErrNoRoute) {
				return fmt.Errorf("%w: %s → %s", backend.ErrNoRoute, sf.Format, toFormat)
			}
			return err
		}
		jobs = append(jobs, runner.Job{
			Source: sf,
			Route:  route,
			Sink:   nil,
			Opts: backend.Options{
				Named:     namedMap,
				StripMeta: f.stripMeta,
			},
		})
	}

	if len(jobs) == 0 {
		return errors.New(i18n.T("error.no_input_files"))
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
					return fmt.Errorf("%s: %w", i18n.T("error.create_outdir"), err)
				}
			} else if isInteractive(cmd) {
				fmt.Fprint(cmd.ErrOrStderr(), i18n.Tf("error.outdir_missing_prompt", map[string]any{"Path": sk.Path}))
				var answer string
				fmt.Fscan(cmd.InOrStdin(), &answer)
				if answer != "y" && answer != "Y" {
					return errors.New(i18n.Tf("error.outdir_missing", map[string]any{"Path": sk.Path}))
				}
				if err := os.MkdirAll(sk.Path, 0o755); err != nil {
					return fmt.Errorf("%s: %w", i18n.T("error.create_outdir"), err)
				}
			} else {
				return errors.New(i18n.Tf("error.outdir_missing", map[string]any{"Path": sk.Path}))
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
		fmt.Fprintln(cmd.ErrOrStderr(), i18n.Tf("runner.summary", map[string]any{
			"OK":   summary.OK,
			"Fail": summary.Fail,
			"Skip": summary.Skip,
		}))
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
	return "", errors.New(i18n.T("error.target_format_unknown"))
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

// parseNamed parses a slice of "key=value" strings into a map.
func parseNamed(kvs []string) map[string]string {
	if len(kvs) == 0 {
		return nil
	}
	m := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		if idx := strings.IndexByte(kv, '='); idx >= 0 {
			m[kv[:idx]] = kv[idx+1:]
		}
	}
	return m
}
