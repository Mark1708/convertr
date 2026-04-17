package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"git.mark1708.ru/me/convertr/internal/formats"
)

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info FILE",
		Short: "Show metadata about a file",
		Args:  cobra.ExactArgs(1),
		RunE:  runInfo,
	}
}

func runInfo(cmd *cobra.Command, args []string) error {
	path := args[0]
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("file not found: %s", path)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "file:   %s\n", path)

	if f, _ := formats.DetectFile(path); f != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "format: %s (%s)\n", f.ID, f.Category)
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
