package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/Mark1708/convertr/internal/config"
	"github.com/Mark1708/convertr/internal/i18n"
	"github.com/Mark1708/convertr/internal/xdg"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: i18n.T("cli.config.short"),
	}
	cmd.AddCommand(
		newConfigPrintCmd(),
		newConfigInitCmd(),
		newConfigValidateCmd(),
	)
	return cmd
}

func newConfigPrintCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "print",
		Short: i18n.T("cli.config.print.short"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			loaded, err := config.Load(rootFlags.Config)
			if err != nil {
				return err
			}
			if rootFlags.Profile != "" {
				loaded.Config = config.MergeProfile(loaded.Config, rootFlags.Profile)
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, i18n.T("cli.config.table_header"))
			fmt.Fprintf(tw, "quality\t%d\t%s\n", loaded.Defaults.Quality, loaded.Sources.Quality)
			fmt.Fprintf(tw, "workers\t%d\t%s\n", loaded.Defaults.Workers, loaded.Sources.Workers)
			fmt.Fprintf(tw, "on_error\t%s\t%s\n", loaded.Defaults.OnError, loaded.Sources.OnError)
			fmt.Fprintf(tw, "on_conflict\t%s\t%s\n", loaded.Defaults.OnConflict, loaded.Sources.OnConflict)
			return tw.Flush()
		},
	}
}

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: i18n.T("cli.config.init.short"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := rootFlags.Config
			if path == "" {
				path = xdg.ConfigPath()
			}
			if _, err := os.Stat(path); err == nil {
				return errors.New(i18n.Tf("error.config_exists", map[string]any{"Path": path}))
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return fmt.Errorf("%s: %w", i18n.T("error.create_config_dir"), err)
			}
			const template = `# convertr configuration
# https://github.com/Mark1708/convertr

[defaults]
quality     = 85
workers     = 0        # 0 = GOMAXPROCS
on_error    = "skip"   # skip | stop | retry
on_conflict = "overwrite" # overwrite | skip | rename | error

# [backend.pandoc]
# extra_args = ["--wrap=none"]

# [profile.hi-res]
# quality = 100
`
			if err := os.WriteFile(path, []byte(template), 0o644); err != nil {
				return fmt.Errorf("%s: %w", i18n.T("error.write_config"), err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), i18n.Tf("output.config_created", map[string]any{"Path": path}))
			return nil
		},
	}
}

func newConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: i18n.T("cli.config.validate.short"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := rootFlags.Config
			if path == "" {
				return errors.New(i18n.T("error.config_no_path"))
			}
			if _, err := os.Stat(path); err != nil {
				return errors.New(i18n.Tf("error.config_not_found", map[string]any{"Path": path}))
			}
			if _, err := config.Load(path); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), i18n.Tf("output.config_ok", map[string]any{"Path": path}))
			return nil
		},
	}
}
