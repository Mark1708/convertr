package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	plugindiscovery "git.mark1708.ru/me/convertr/internal/backend/backends/plugin"
)

func newPluginsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "Manage external plugin backends",
	}
	cmd.AddCommand(
		newPluginsListCmd(),
		newPluginsTestCmd(),
	)
	return cmd
}

func newPluginsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List convertr-* plugins found in PATH",
		RunE: func(cmd *cobra.Command, _ []string) error {
			plugins := findPluginExecutables()
			if len(plugins) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no plugins found in PATH")
				return nil
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "PLUGIN\tPATH")
			for _, p := range plugins {
				fmt.Fprintf(tw, "%s\t%s\n", p.name, p.path)
			}
			return tw.Flush()
		},
	}
}

func newPluginsTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test each plugin by calling its capabilities sub-command",
		RunE: func(cmd *cobra.Command, _ []string) error {
			plugins := findPluginExecutables()
			if len(plugins) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no plugins found in PATH")
				return nil
			}
			tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "PLUGIN\tSTATUS\tCAPABILITIES")
			for _, p := range plugins {
				caps, err := plugindiscovery.ProbePlugin(p.name)
				if err != nil {
					fmt.Fprintf(tw, "%s\tERROR\t%v\n", p.name, err)
					continue
				}
				fmt.Fprintf(tw, "%s\tOK\t%d capabilities\n", p.name, len(caps))
			}
			return tw.Flush()
		},
	}
}

type pluginEntry struct {
	name string
	path string
}

func findPluginExecutables() []pluginEntry {
	seen := map[string]bool{}
	var found []pluginEntry
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
				found = append(found, pluginEntry{name: name, path: full})
			}
		}
	}
	return found
}
