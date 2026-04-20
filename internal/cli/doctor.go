package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Mark1708/convertr/internal/backend/execx"
	"github.com/Mark1708/convertr/internal/i18n"
)

type depSpec struct {
	binary      string
	versionArgs []string // args to pass to get version string
	install     string
}

var deps = []depSpec{
	{"pandoc", []string{"--version"}, "brew install pandoc"},
	{"ffmpeg", []string{"-version"}, "brew install ffmpeg"},
	{"soffice", []string{"--version"}, "brew install --cask libreoffice"},
	{"magick", []string{"--version"}, "brew install imagemagick"},
	{"jq", []string{"--version"}, "brew install jq"},
	{"yq", []string{"--version"}, "brew install yq"},
	{"tesseract", []string{"--version"}, "brew install tesseract"},
	{"figlet", []string{"--version"}, "brew install figlet"},
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: i18n.T("cli.doctor.short"),
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "%s\n\n", i18n.T("doctor.header"))

			missing := 0
			for _, d := range deps {
				path, err := exec.LookPath(d.binary)
				if err != nil {
					// Try "convert" as fallback name for ImageMagick 6.x.
					if d.binary == "magick" {
						if p, e := exec.LookPath("convert"); e == nil {
							path = p
							err = nil
						}
					}
				}
				if err != nil {
					fmt.Fprintf(w, "  %-14s %-18s %s  (%s)\n",
						d.binary, "",
						i18n.T("doctor.missing"),
						i18n.Tf("doctor.install_hint", map[string]any{"Cmd": d.install}))
					missing++
					continue
				}
				ver := shortVersion(execx.Version(ctx, d.binary, d.versionArgs...))
				fmt.Fprintf(w, "  %-14s %-18s %s  (%s)\n",
					d.binary, ver, i18n.T("doctor.ok"), path)
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

// shortVersion extracts a short version token from a full version string.
// E.g. "pandoc 3.7.0" → "3.7.0", "ffmpeg version 7.1" → "7.1".
func shortVersion(full string) string {
	if full == "" {
		return ""
	}
	fields := strings.Fields(full)
	// Find first field that looks like a version number (contains a digit).
	for _, f := range fields {
		hasDigit := false
		for _, r := range f {
			if r >= '0' && r <= '9' {
				hasDigit = true
				break
			}
		}
		if hasDigit {
			// Strip trailing punctuation.
			return strings.TrimRight(f, ",;")
		}
	}
	return fields[0]
}
