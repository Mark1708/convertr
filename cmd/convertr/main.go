package main

import (
	"os"

	"github.com/Mark1708/convertr/internal/cli"
	"github.com/Mark1708/convertr/internal/i18n"
	"github.com/Mark1708/convertr/internal/slogx"

	// Register conversion backends.
	_ "github.com/Mark1708/convertr/internal/backend/backends/asciidoctor"
	_ "github.com/Mark1708/convertr/internal/backend/backends/csvkit"
	_ "github.com/Mark1708/convertr/internal/backend/backends/ffmpeg"
	_ "github.com/Mark1708/convertr/internal/backend/backends/figlet"
	_ "github.com/Mark1708/convertr/internal/backend/backends/imagemagick"
	_ "github.com/Mark1708/convertr/internal/backend/backends/jq"
	_ "github.com/Mark1708/convertr/internal/backend/backends/libreoffice"
	_ "github.com/Mark1708/convertr/internal/backend/backends/pandoc"
	_ "github.com/Mark1708/convertr/internal/backend/backends/plugin"
	_ "github.com/Mark1708/convertr/internal/backend/backends/tesseract"
	_ "github.com/Mark1708/convertr/internal/backend/backends/textutil"
	_ "github.com/Mark1708/convertr/internal/backend/backends/yq"
)

// Version is set by goreleaser via -ldflags.
var Version = "dev"

func main() {
	slogx.Init()
	if err := i18n.Init(""); err != nil {
		// Non-fatal: fall back to key names.
		_ = err
	}

	root := cli.New(Version)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
