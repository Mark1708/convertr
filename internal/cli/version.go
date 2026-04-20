package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Mark1708/convertr/internal/i18n"
)

func newVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: i18n.T("cli.version.short"),
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "convertr %s\n", version)
		},
	}
}
