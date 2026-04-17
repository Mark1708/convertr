package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"git.mark1708.ru/me/convertr/internal/backend"
	"git.mark1708.ru/me/convertr/internal/formats"
	"git.mark1708.ru/me/convertr/internal/router"
	"git.mark1708.ru/me/convertr/internal/runner"
	"git.mark1708.ru/me/convertr/internal/sink"
	"git.mark1708.ru/me/convertr/internal/source"
	"git.mark1708.ru/me/convertr/internal/watch"
)

type watchFlags struct {
	output     string
	toFormat   string
	fromFormat string
	debounce   time.Duration
	onDelete   string
}

func newWatchCmd() *cobra.Command {
	var f watchFlags

	cmd := &cobra.Command{
		Use:   "watch SRC -o DST",
		Short: "Watch a directory and convert new/changed files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWatch(cmd, args[0], f)
		},
	}

	fl := cmd.Flags()
	fl.StringVarP(&f.output, "output", "o", "", "output directory (required)")
	fl.StringVar(&f.toFormat, "to", "", "target format ID (e.g. md, html)")
	fl.StringVar(&f.fromFormat, "from", "", "source format override")
	fl.DurationVar(&f.debounce, "debounce", 300*time.Millisecond, "debounce interval")
	fl.StringVar(&f.onDelete, "on-delete", "keep", "delete policy: keep|remove|archive")
	_ = cmd.MarkFlagRequired("output")

	return cmd
}

func runWatch(cmd *cobra.Command, src string, f watchFlags) error {
	if f.toFormat == "" {
		return fmt.Errorf("watch requires --to FORMAT")
	}

	deletePolicy, err := watch.ParseDeletePolicy(f.onDelete)
	if err != nil {
		return err
	}

	cfg := watch.Config{
		Debounce:     f.debounce,
		DeletePolicy: deletePolicy,
	}

	watcher, events, err := watch.New(cfg)
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer watcher.Close()

	if err := watcher.Add(src); err != nil {
		return fmt.Errorf("watch %s: %w", src, err)
	}

	sk := &sink.Sink{
		Type:   sink.SinkTypeDir,
		Path:   f.output,
		Format: f.toFormat,
		Policy: sink.ConflictOverwrite,
	}
	g := router.Build()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Fprintf(cmd.OutOrStdout(), "Watching %s → %s (Ctrl+C to stop)\n", src, f.output)

	for {
		select {
		case <-ctx.Done():
			fmt.Fprintln(cmd.OutOrStdout(), "\nStopped.")
			return nil
		case ev, ok := <-events:
			if !ok {
				return nil
			}
			if err := handleWatchEvent(ctx, ev.Path, f.toFormat, f.fromFormat, sk, g); err != nil {
				slog.Warn("watch: conversion failed", "file", ev.Path, "err", err)
			}
		}
	}
}

func handleWatchEvent(ctx context.Context, path, toFormat, fromFormat string, sk *sink.Sink, g *router.Graph) error {
	var sf source.SourceFile
	var sfErr error
	if fromFormat != "" {
		for s, e := range source.FileSourceWithFormat(path, fromFormat) {
			sf, sfErr = s, e
		}
	} else {
		for s, e := range source.FileSource(path) {
			sf, sfErr = s, e
		}
	}
	if sfErr != nil {
		return sfErr
	}
	if sf.Format == "" {
		return nil // unknown format; skip silently
	}

	route, err := g.Find(sf.Format, toFormat)
	if err != nil {
		if errors.Is(err, backend.ErrNoRoute) {
			return nil // no route for this format; skip silently
		}
		return err
	}

	outExt := watchFirstExt(toFormat)
	outName := watchStem(filepath.Base(path)) + outExt
	outPath := filepath.Join(sk.Path, outName)

	fileSink := &sink.Sink{
		Type:   sink.SinkTypeFile,
		Path:   outPath,
		Format: toFormat,
		Policy: sink.ConflictOverwrite,
	}

	opts := runner.RunOpts{
		Workers:    1,
		OnError:    runner.ErrorPolicyStop,
		OnConflict: sink.ConflictOverwrite,
	}

	_, err = runner.Execute(ctx, []runner.Job{{Source: sf, Route: route, Sink: fileSink}}, opts)
	if err == nil {
		slog.Info("watch: converted", "file", path, "out", outPath)
	}
	return err
}

func watchFirstExt(formatID string) string {
	f := formats.ByID(formatID)
	if f == nil || len(f.Extensions) == 0 {
		return "." + formatID
	}
	return f.Extensions[0]
}

func watchStem(name string) string {
	ext := filepath.Ext(name)
	return name[:len(name)-len(ext)]
}
