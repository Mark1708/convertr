package cli

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"git.mark1708.ru/me/convertr/internal/i18n"
)

type depSpec struct {
	binary  string
	install string
}

var deps = []depSpec{
	{"pandoc", "brew install pandoc"},
	{"ffmpeg", "brew install ffmpeg"},
	{"soffice", "brew install --cask libreoffice"},
	{"convert", "brew install imagemagick"},
	{"jq", "brew install jq"},
	{"yq", "brew install yq"},
	{"tesseract", "brew install tesseract"},
	{"figlet", "brew install figlet"},
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: i18n.T("cli.doctor.short"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "%s\n\n", i18n.T("doctor.header"))

			missing := 0
			for _, d := range deps {
				path, err := exec.LookPath(d.binary)
				if err != nil {
					fmt.Fprintf(w, "  %-14s %s  (%s)\n", d.binary,
						i18n.T("doctor.missing"),
						i18n.Tf("doctor.install_hint", map[string]any{"Cmd": d.install}))
					missing++
				} else {
					fmt.Fprintf(w, "  %-14s %s  (%s)\n", d.binary, i18n.T("doctor.ok"), path)
				}
			}

			fmt.Fprintln(w)
			if missing > 0 {
				fmt.Fprintln(w, i18n.T("doctor.has_missing"))
			} else {
				fmt.Fprintln(w, i18n.T("doctor.all_ok"))
			}
			return nil
		},
	}
}
