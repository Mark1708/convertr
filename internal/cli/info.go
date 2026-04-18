package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"git.mark1708.ru/me/convertr/internal/formats"
	"git.mark1708.ru/me/convertr/internal/i18n"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info FILE",
		Short: i18n.T("cli.info.short"),
		Args:  cobra.ExactArgs(1),
		RunE:  runInfo,
	}
}

func runInfo(cmd *cobra.Command, args []string) error {
	path := args[0]
	if _, err := os.Stat(path); err != nil {
		return errors.New(i18n.Tf("error.file_not_found", map[string]any{"Path": path}))
	}

	fmt.Fprintln(cmd.OutOrStdout(), i18n.Tf("cli.info.file_label", map[string]any{"Path": path}))

	if f, _ := formats.DetectFile(path); f != nil {
		fmt.Fprintln(cmd.OutOrStdout(), i18n.Tf("cli.info.format_label", map[string]any{"ID": f.ID, "Category": f.Category}))
	}

	// Try ffprobe first (best for audio/video).
	if out, err := runCommand("ffprobe", "-v", "quiet", "-print_format", "flat",
		"-show_format", "-show_streams", path); err == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "\n--- ffprobe ---")
		fmt.Fprintln(cmd.OutOrStdout(), truncate(out, 2000))
		return nil
	}

	// Try exiftool for image/document metadata.
	if out, err := runCommand("exiftool", path); err == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "\n--- exiftool ---")
		fmt.Fprintln(cmd.OutOrStdout(), truncate(out, 2000))
		return nil
	}

	// Fall back to the POSIX file(1) command.
	if out, err := runCommand("file", path); err == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "\n--- file ---")
		fmt.Fprintln(cmd.OutOrStdout(), strings.TrimSpace(out))
		return nil
	}

	return nil
}

func runCommand(name string, args ...string) (string, error) {
	bin, err := exec.LookPath(name)
	if err != nil {
		return "", err
	}
	out, err := exec.Command(bin, args...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n[...truncated]"
}
