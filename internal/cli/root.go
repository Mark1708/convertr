package cli

import (
	"os"

	"github.com/spf13/cobra"

	"git.mark1708.ru/me/convertr/internal/i18n"
	"git.mark1708.ru/me/convertr/internal/slogx"
	"git.mark1708.ru/me/convertr/internal/xdg"
)

// RootFlags holds global persistent flags.
type RootFlags struct {
	Config  string
	Profile string
	Lang    string
	Verbose int
	Quiet   bool
	JSON    bool
	NoColor bool
}

var rootFlags RootFlags

// New builds and returns the root cobra command.
func New(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "convertr",
		Short: i18n.T("cli.root.short"),
		Long:  i18n.T("cli.root.long"),
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Reinitialise i18n now that --lang is parsed.
			if err := i18n.Init(rootFlags.Lang); err != nil {
				return err
			}
			// Configure logging.
			switch {
			case rootFlags.Quiet:
				// silence all
			case rootFlags.JSON || os.Getenv("CI") == "1":
				slogx.SetJSON(slogx.LevelFromVerbosity(rootFlags.Verbose))
			default:
				slogx.SetLevel(slogx.LevelFromVerbosity(rootFlags.Verbose))
			}
			return nil
		},
	}

	pf := root.PersistentFlags()
	pf.StringVar(&rootFlags.Config, "config", xdg.ConfigPath(), i18n.T("cli.flag.config"))
	pf.StringVar(&rootFlags.Profile, "profile", "", i18n.T("cli.flag.profile"))
	pf.StringVar(&rootFlags.Lang, "lang", "", i18n.T("cli.flag.lang"))
	pf.CountVarP(&rootFlags.Verbose, "verbose", "v", i18n.T("cli.flag.verbose"))
	pf.BoolVarP(&rootFlags.Quiet, "quiet", "q", false, i18n.T("cli.flag.quiet"))
	pf.BoolVar(&rootFlags.JSON, "json", false, i18n.T("cli.flag.json"))
	pf.BoolVar(&rootFlags.NoColor, "no-color", false, i18n.T("cli.flag.no_color"))

	root.AddCommand(
		newVersionCmd(version),
		newDoctorCmd(),
		newFormatsCmd(),
	)

	return root
}
