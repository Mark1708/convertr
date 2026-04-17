package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"git.mark1708.ru/me/convertr/internal/backend"
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
	fmt.Fprintln(w, `  node [shape=box fontname="Helvetica"];`)

	// Nodes: all known formats grouped by category colour.
	catColor := map[formats.Category]string{
		formats.CategoryDocument: "#dae8fc",
		formats.CategoryMarkup:   "#d5e8d4",
		formats.CategoryData:     "#fff2cc",
		formats.CategoryImage:    "#f8cecc",
		formats.CategoryAudio:    "#e1d5e7",
		formats.CategoryVideo:    "#f5deb3",
		formats.CategoryTextArt:  "#d3d3d3",
		formats.CategoryArchive:  "#c0c0c0",
	}
	all := formats.All()
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	for _, f := range all {
		color := catColor[f.Category]
		if color == "" {
			color = "#ffffff"
		}
		fmt.Fprintf(w, "  %s [style=filled fillcolor=%q];\n", f.ID, color)
	}

	// Edges from registered backend capabilities.
	caps := backend.AllCapabilities()
	// Deduplicate edges (multiple backends may declare the same from→to).
	type edge struct{ from, to string }
	seen := make(map[edge]struct{})
	for _, c := range caps {
		e := edge{c.From, c.To}
		if _, ok := seen[e]; ok {
			continue
		}
		seen[e] = struct{}{}
		fmt.Fprintf(w, "  %s -> %s;\n", c.From, c.To)
	}

	fmt.Fprintln(w, "}")
	return nil
}
