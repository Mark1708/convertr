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
	binary      string   // display name / primary lookup
	alternates  []string // additional binaries that satisfy this dependency (OR)
	versionArgs []string // args passed to the found binary for its version
	install     string
}

// deps enumerates the external tooling convertr dispatches to. Entries
// with alternates are satisfied when any of the listed binaries is
// present — e.g. csvkit ships several scripts and users only need one.
var deps = []depSpec{
	{binary: "pandoc", versionArgs: []string{"--version"}, install: "brew install pandoc"},
	{binary: "ffmpeg", versionArgs: []string{"-version"}, install: "brew install ffmpeg"},
	{binary: "soffice", versionArgs: []string{"--version"}, install: "brew install --cask libreoffice"},
	{binary: "magick", alternates: []string{"convert"}, versionArgs: []string{"--version"}, install: "brew install imagemagick"},
	{binary: "jq", versionArgs: []string{"--version"}, install: "brew install jq"},
	{binary: "yq", versionArgs: []string{"--version"}, install: "brew install yq"},
	{binary: "tesseract", versionArgs: []string{"--version"}, install: "brew install tesseract"},
	{binary: "figlet", versionArgs: []string{"--version"}, install: "brew install figlet"},
	{binary: "asciidoctor", alternates: []string{"asciidoctor-pdf"}, versionArgs: []string{"--version"}, install: "brew install asciidoctor"},
	{binary: "in2csv", alternates: []string{"xlsx2csv", "csvjson"}, versionArgs: []string{"--version"}, install: "pipx install csvkit  # or: brew install xlsx2csv"},
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
				path, found := findDep(d)
				if found == "" {
					fmt.Fprintf(w, "  %-14s %-18s %s  (%s)\n",
						d.binary, "",
						i18n.T("doctor.missing"),
						i18n.Tf("doctor.install_hint", map[string]any{"Cmd": d.install}))
					missing++
					continue
				}
				ver := shortVersion(execx.Version(ctx, found, d.versionArgs...))
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

// findDep returns (path, binaryName) for the first binary from d that is
// available on $PATH, or ("","") when none are installed. The primary
// name is tried first, then alternates in declaration order.
func findDep(d depSpec) (path, found string) {
	for _, bin := range append([]string{d.binary}, d.alternates...) {
		if p, err := exec.LookPath(bin); err == nil {
			return p, bin
		}
	}
	return "", ""
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
