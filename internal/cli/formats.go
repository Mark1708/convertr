package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"git.mark1708.ru/me/convertr/internal/formats"
	"git.mark1708.ru/me/convertr/internal/i18n"
)

func newFormatsCmd() *cobra.Command {
	var dotOutput bool

	cmd := &cobra.Command{
		Use:   "formats",
		Short: i18n.T("cli.formats.short"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if dotOutput {
				return printDot(cmd)
			}
			return printTable(cmd)
		},
	}
	cmd.Flags().BoolVar(&dotOutput, "dot", false, "output Graphviz DOT format")
	return cmd
}

func printTable(cmd *cobra.Command) error {
	w := cmd.OutOrStdout()
	all := formats.All()
	sort.Slice(all, func(i, j int) bool {
		if all[i].Category != all[j].Category {
			return all[i].Category < all[j].Category
		}
		return all[i].ID < all[j].ID
	})
	curCat := formats.Category("")
	for _, f := range all {
		if f.Category != curCat {
			fmt.Fprintf(w, "\n[%s]\n", strings.ToUpper(string(f.Category)))
			curCat = f.Category
		}
		fmt.Fprintf(w, "  %-8s  %s\n", f.ID, strings.Join(f.Extensions, ", "))
	}
	return nil
}

func printDot(cmd *cobra.Command) error {
	w := cmd.OutOrStdout()
	fmt.Fprintln(w, "digraph convertr {")
	fmt.Fprintln(w, "  rankdir=LR;")
	fmt.Fprintln(w, "  node [shape=box];")
	// Edges will be populated from backend capabilities once backends are registered.
	// For now, emit nodes only.
	all := formats.All()
	for _, f := range all {
		fmt.Fprintf(w, "  %s;\n", f.ID)
	}
	fmt.Fprintln(w, "}")
	return nil
}
