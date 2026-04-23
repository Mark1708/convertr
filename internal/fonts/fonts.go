// Package fonts defines the font preferences consumed by PDF-producing backends.
//
// The package lives outside of both `config` and `backend` to avoid an import
// cycle: backends import fonts to render PDFs with sensible defaults, while
// config embeds fonts.Config to let users override those defaults.
package fonts

import "runtime"

// Config holds font family preferences. Empty fields mean "use backend default".
type Config struct {
	Mainfont string `toml:"mainfont"`
	Monofont string `toml:"monofont"`
	Sansfont string `toml:"sansfont"`
}

// Default returns the platform-appropriate font set.
//
// Values are chosen to be present out-of-the-box on each OS: PT Serif and
// Menlo ship with macOS; DejaVu is installed by default on most Linux
// distributions; Times New Roman and Consolas are standard on Windows.
func Default() Config {
	switch runtime.GOOS {
	case "darwin":
		return Config{
			Mainfont: "PT Serif",
			Monofont: "Menlo",
			Sansfont: "Helvetica Neue",
		}
	case "windows":
		return Config{
			Mainfont: "Times New Roman",
			Monofont: "Consolas",
			Sansfont: "Segoe UI",
		}
	default:
		return Config{
			Mainfont: "DejaVu Serif",
			Monofont: "DejaVu Sans Mono",
			Sansfont: "DejaVu Sans",
		}
	}
}
