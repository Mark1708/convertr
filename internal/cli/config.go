package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"git.mark1708.ru/me/convertr/internal/config"
	"git.mark1708.ru/me/convertr/internal/xdg"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage convertr configuration",
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
		Short: "Show active configuration with value sources",
		RunE: func(cmd *cobra.Command, _ []string) error {
			loaded, err := config.Load(rootFlags.Config)
			if err != nil {
				return err
			}
			if rootFlags.Profile != "" {
				loaded.Config = config.MergeProfile(loaded.Config, rootFlags.Profile)
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "FIELD\tVALUE\tSOURCE")
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
		Short: "Create a default config file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := rootFlags.Config
			if path == "" {
				path = xdg.ConfigPath()
			}
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("config file already exists: %s", path)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return fmt.Errorf("create config dir: %w", err)
			}
			const template = `# convertr configuration
# https://git.mark1708.ru/me/convertr

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
				return fmt.Errorf("write config: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", path)
			return nil
		},
	}
}

func newConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate config file syntax",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := rootFlags.Config
			if path == "" {
				return fmt.Errorf("no config file specified")
			}
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("config file not found: %s", path)
			}
			if _, err := config.Load(path); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ok: %s\n", path)
			return nil
		},
	}
}
