# convertr

> Universal file-format converter CLI — convert 50+ formats with one binary.

[![CI](https://github.com/Mark1708/convertr/actions/workflows/ci.yml/badge.svg)](https://github.com/Mark1708/convertr/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Mark1708/convertr.svg)](https://pkg.go.dev/github.com/Mark1708/convertr)
[![Go Report Card](https://goreportcard.com/badge/github.com/Mark1708/convertr)](https://goreportcard.com/report/github.com/Mark1708/convertr)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

convertr wraps pandoc, ffmpeg, LibreOffice, ImageMagick, jq, yq, Tesseract and a dozen more tools behind a single `convertr FILE -o OUT` interface. It finds the shortest conversion path automatically — you never need to remember which binary handles which format.

**Documents:** MD · DOCX · PDF · ODT · HTML · EPUB · RST · TeX · PPTX  
**Images:** JPG · PNG · WebP · SVG · AVIF · HEIC · GIF  
**Video:** MP4 · MKV · WebM · AVI · MOV  
**Audio:** MP3 · FLAC · AAC · WAV · OGG  
**Data:** JSON · YAML · TOML · CSV · XLSX  
**OCR:** image → text via Tesseract

---

## Why convertr?

Most conversions need a different tool: `pandoc` for documents, `ffmpeg` for
video, `convert` for images. convertr wraps them all — just specify source and
target format, and it picks the right backend automatically, even chaining
multiple tools when no direct route exists.

- Zero format memorisation — just specify source and target
- Batch conversion with parallel workers and retry policies
- Watch mode: auto-convert on file save
- Plugin protocol for custom backends

---

## Table of contents

- [Install](#install)
- [Quick start](#quick-start)
- [Config — config.toml](#config--configtoml)
- [CLI reference](#cli-reference)
- [Backends](#backends)
- [Watch mode](#watch-mode)
- [Plugins](#plugins)
- [Progress reporting](#progress-reporting)
- [Architecture](#architecture)
- [Troubleshooting](#troubleshooting)
- [License](#license)

---

## Install

### `go install`

```sh
go install github.com/Mark1708/convertr/cmd/convertr@latest
```

Requires Go 1.25+.

### Homebrew

```sh
brew install mark1708/tap/convertr
```

### From source

```sh
git clone https://github.com/Mark1708/convertr.git
cd convertr
go build -o ~/.local/bin/convertr ./cmd/convertr
```

### Install backend dependencies

```sh
# macOS
./scripts/install-deps-macos.sh

# Debian / Ubuntu
./scripts/install-deps-linux.sh
```

### Verify

```sh
convertr version
convertr doctor
```

`doctor` checks every backend binary, prints the detected version, and suggests `brew install` / `apt-get install` commands for anything missing.

---

## Quick start

```sh
# Convert a single file — format is inferred from the extension.
convertr report.docx -o report.md

# Specify target format explicitly.
convertr notes.md --to pdf -o notes.pdf

# Batch: convert a whole directory, one Markdown per DOCX.
convertr -r ./docs/ -o ./out/ --to md

# Parallel batch with 4 workers.
convertr -r ./inbox/ -o ./outbox/ --to pdf -j 4

# Stdin → stdout (pipe-friendly).
cat data.json | convertr - --from json --to yaml -o -

# Dry run — print the planned conversions without executing.
convertr -r ./docs/ -o ./out/ --to md --dry-run

# Watch a directory and convert on every change.
convertr watch ./inbox -o ./outbox --to md

# Show all known formats and the conversion graph.
convertr formats
convertr formats --dot | dot -Tsvg > /tmp/formats.svg
```

---

## Config — config.toml

Lives at `~/.config/convertr/config.toml` (or wherever `--config` points, or `$CONVERTR_CONFIG`).

Create a default file:

```sh
convertr config init
```

### Full example

```toml
# ~/.config/convertr/config.toml

[defaults]
quality     = 85       # 0–100; backend-specific interpretation
workers     = 0        # 0 = GOMAXPROCS; > 0 = fixed count
on_error    = "skip"   # skip | stop | retry
on_conflict = "overwrite" # overwrite | skip | rename | error

# Extra arguments forwarded to individual backends.
[backend.pandoc]
extra_args = ["--wrap=none", "--variable=lang:ru"]

[backend.ffmpeg]
extra_args = ["-preset", "fast"]

[backend.tesseract]
extra_args = ["--dpi", "300"]

# Named profiles — activate with --profile NAME.
[profile.hi-res]
quality = 100

[profile.ci]
workers     = 4
on_error    = "stop"
on_conflict = "error"

[profile.ocr]
quality = 95
```

### Defaults

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `quality` | int | 85 | Conversion quality hint (0–100) |
| `workers` | int | 0 | Parallel workers; 0 = GOMAXPROCS |
| `on_error` | string | `skip` | What to do when a job fails |
| `on_conflict` | string | `overwrite` | What to do when output already exists |

### Environment overrides

| Variable | Overrides |
|----------|-----------|
| `CONVERTR_QUALITY` | `defaults.quality` |
| `CONVERTR_WORKERS` | `defaults.workers` |
| `CONVERTR_ON_ERROR` | `defaults.on_error` |
| `CONVERTR_ON_CONFLICT` | `defaults.on_conflict` |

### Priority order

```
hardcoded defaults → config file → environment variables → CLI flags
```

CLI flags always win. A missing config file is not an error — defaults are used.

---

## CLI reference

### Global flags

Available on every subcommand:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | | `~/.config/convertr/config.toml` | Path to config file |
| `--profile` | | | Activate a named profile |
| `--lang` | | | Language override (`en`, `ru`) |
| `--verbose` | `-v` | 0 | Increase log verbosity (repeatable: `-vvv`) |
| `--quiet` | `-q` | false | Silence all output except errors |
| `--json` | | false | Emit structured JSON logs |
| `--no-color` | | false | Disable ANSI color |

---

### Convert (default command)

`convertr FILE [FILE...] -o OUTPUT [flags]`

When no subcommand is given, convertr treats arguments as input files and runs the conversion pipeline.

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | | Output file, directory, or `-` for stdout |
| `--to` | | | Target format ID (e.g. `md`, `pdf`, `mp3`) |
| `--from` | | | Source format override (auto-detected if omitted) |
| `--dry-run` | | false | Print planned conversions without running them |
| `--workers` | `-j` | 1 | Parallel workers |
| `--on-error` | | `skip` | Error policy: `skip` \| `stop` \| `retry` |
| `--on-conflict` | | `overwrite` | Conflict policy: `overwrite` \| `skip` \| `rename` \| `error` |
| `--recursive` | `-r` | false | Recurse into directories |
| `--mkdir` | | false | Create output directory if it does not exist |

**Input types:**

```sh
convertr file.md -o out.pdf              # single file
convertr a.md b.md -o out/ --mkdir       # multiple files → output directory (created if absent)
convertr "src/**/*.md" -o out/ --to pdf  # glob
convertr -r ./src/ -o ./out/ --to html   # directory (recursive)
convertr - --from json --to yaml -o -    # stdin → stdout
```

**Output directory resolution:**

When multiple input files are given, the output path is always treated as a directory — even if no trailing `/` is present. If the directory does not yet exist, convertr will:

- ask interactively `Create it? [y/N]` when running in a TTY, or
- return an error with a hint to use `--mkdir` in non-interactive mode.

```sh
# Auto-create the output directory:
convertr -r ./docs/ -o ./out --to md --mkdir

# TTY prompt (no --mkdir needed):
convertr -r ./docs/ -o ./out --to md
# Output directory does not exist: ./out
# Create it? [y/N]
```

**Error policies:**

| Policy | Behaviour |
|--------|-----------|
| `skip` | Record error, continue with remaining jobs |
| `stop` | Abort all remaining jobs on first error |
| `retry` | Retry with exponential backoff (max 3 attempts) |

**Conflict policies:**

| Policy | Behaviour |
|--------|-----------|
| `overwrite` | Replace existing output file |
| `skip` | Leave existing output file unchanged |
| `rename` | Append numeric suffix (`.1`, `.2`, …) |
| `error` | Fail if output file exists |

---

### doctor

```sh
convertr doctor
```

Checks every backend binary: detects the installed version, shows the full path, and prints an install hint for anything missing.

---

### formats

```sh
convertr formats
convertr formats --dot | dot -Tsvg > graph.svg
```

`--dot` emits a Graphviz DOT graph of all available conversion edges (coloured by format category).

---

### watch

```sh
convertr watch SRC -o DST --to FORMAT [flags]
```

Watches `SRC` recursively and converts every new or modified file to `DST`.

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--output` / `-o` | | Output directory (required) |
| `--to` | | Target format (required) |
| `--from` | | Source format override |
| `--debounce` | `300ms` | Wait after last event before converting |
| `--on-delete` | `keep` | What to do when source is deleted: `keep` \| `remove` \| `archive` |

See [Watch mode](#watch-mode) for details.

---

### config

```sh
convertr config print      # Show active config with value sources
convertr config init       # Create default config file
convertr config validate   # Validate TOML syntax
```

`config print` shows each field, its current value, and where it came from (`default` / `file` / `env` / `flag`).

---

### plugins

```sh
convertr plugins list   # List convertr-* executables in PATH
convertr plugins test   # Probe each plugin's capabilities subcommand
```

See [Plugins](#plugins) for the protocol.

---

### info

```sh
convertr info FILE
```

Detects the file format, then probes metadata using the first available tool: `ffprobe` (audio/video), `exiftool` (images/documents), `file` (fallback).

---

### version

```sh
convertr version
```

---

## Backends

Each backend is an `init()`-registered plugin that bridges one or more external binaries to the router.

### Pandoc

**Binary:** `pandoc` · `brew install pandoc` / `apt install pandoc`

| From | To |
|------|----|
| `md` | `html` `docx` `odt` `pdf` `rst` `epub` `tex` `txt` |
| `html` | `md` `docx` `pdf` `txt` |
| `docx` | `md` `html` `pdf` `odt` `txt` `rst` |
| `odt` | `md` `docx` `html` `txt` |
| `rst` | `md` `html` `pdf` `txt` |
| `epub` | `md` `html` `txt` |
| `tex` | `md` `html` `pdf` |
| `org` | `md` `html` `pdf` |

PDF output uses `xelatex` when available, `pdflatex` otherwise.

**Config:** `[backend.pandoc]` `extra_args = ["--wrap=none"]`

---

### FFmpeg

**Binary:** `ffmpeg` · `brew install ffmpeg` / `apt install ffmpeg`

| From | To |
|------|----|
| `mp4` `mkv` `webm` `mov` `avi` | (all video formats, bidirectional) |
| `mp3` `flac` `aac` `ogg` `wav` `m4a` `opus` | (all audio formats, bidirectional) |
| video formats | `gif` (two-pass palette) |
| video formats | audio formats (extract audio) |

Quality hint → CRF (video) or bitrate (audio).

**Config:** `[backend.ffmpeg]` `extra_args = ["-preset", "fast"]`

---

### ImageMagick

**Binary:** `magick` (IM7) or `convert` (IM6) · `brew install imagemagick`

| From | To |
|------|----|
| `jpg` `png` `webp` `gif` `tiff` `bmp` | (all raster formats, bidirectional) |
| `avif` | raster formats (and vice versa) |
| `svg` | raster formats, `avif` |
| `heic` | raster formats |

Quality hint → `-quality N`.

---

### LibreOffice

**Binary:** `soffice` · `brew install --cask libreoffice` / `apt install libreoffice`

| From | To |
|------|----|
| `doc` `docx` `odt` `rtf` | `odt` `docx` `pdf` `txt` |
| `xlsx` `ods` | `csv` `ods` `xlsx` |
| `pptx` `odp` | `odp` `pptx` `pdf` |

Each conversion runs with an isolated `--env:UserInstallation` directory to allow safe parallel execution.

---

### jq

**Binary:** `jq` · `brew install jq`

| From | To | Notes |
|------|----|-------|
| `json` | `json` | pretty-print, minify, or transform |

Named options: `jq.minify=true` (compact output), `jq.filter=EXPR` (custom expression).

---

### yq

**Binary:** `yq` · `brew install yq`

| From | To |
|------|----|
| `yaml` | `json` `toml` |
| `json` | `yaml` `toml` |
| `toml` | `yaml` `json` |

---

### Tesseract (OCR)

**Binary:** `tesseract` · `brew install tesseract tesseract-lang`

| From | To |
|------|----|
| `jpg` `png` `tiff` | `txt` |

Default language: `rus+eng`. Override: `[backend.tesseract]` `extra_args = ["-l", "eng"]`.

---

### CSVKit

**Binary:** `in2csv` / `xlsx2csv` (fallback) · `pip install csvkit`

| From | To |
|------|----|
| `xlsx` | `csv` |
| `csv` | `json` |
| `xlsx` | `json` |

---

### AsciiDoctor

**Binary:** `asciidoctor` · `gem install asciidoctor asciidoctor-pdf`

| From | To |
|------|----|
| `adoc` | `html` |
| `adoc` | `pdf` (requires `asciidoctor-pdf`) |

---

### Figlet (ASCII art)

**Binary:** `figlet` · `brew install figlet`

| From | To |
|------|----|
| `txt` | `ascii` |

---

### textutil (macOS only)

**Binary:** `textutil` (bundled with macOS)

| From | To |
|------|----|
| `doc` `rtf` | `txt` `html` |

Available only on macOS; compiled out on other platforms via build tag.

---

## Watch mode

```sh
convertr watch ./inbox -o ./outbox --to md
convertr watch ./raw   -o ./web   --to html --debounce 500ms --on-delete archive
```

**How it works:**

1. Recursively watches `SRC` using fsnotify.
2. Debounces rapid edits — waits `--debounce` (default 300 ms) after the last event before triggering.
3. Detects the format of the changed file, finds a route, converts, writes to `DST`.
4. New subdirectories are watched automatically.

**Delete policies:**

| Policy | What happens when source is deleted |
|--------|-------------------------------------|
| `keep` | Output file is left as-is (default) |
| `remove` | Output file is deleted |
| `archive` | Output file is moved to `DST/.archive/` |

**Graceful shutdown:** Press Ctrl+C (SIGINT / SIGTERM). convertr drains in-flight conversions and exits 0.

---

## Plugins

convertr supports external plugins — any executable named `convertr-*` found in `PATH`.

### Plugin protocol

A plugin must implement two subcommands:

#### `convertr-NAME capabilities`

Writes a JSON array to stdout and exits 0:

```json
[
  { "from": "wasm", "to": "wat",  "cost": 2 },
  { "from": "wat",  "to": "wasm", "cost": 2 }
]
```

Fields:
- `from` / `to` — format IDs (must match convertr's registry or be new IDs)
- `cost` — routing cost (1–10; lower = preferred over built-in backends); defaults to 5

#### `convertr-NAME convert`

```sh
convertr-NAME convert \
  --from FROM \
  --to   TO   \
  --input  /tmp/in.wat  \
  --output /tmp/out.wasm \
  [--opt key=value ...]
```

Exit 0 on success. On failure: exit non-zero, write a single error line to stderr.

### Writing a plugin

```sh
#!/usr/bin/env bash
case "$1" in
  capabilities)
    echo '[{"from":"foo","to":"bar","cost":3}]'
    ;;
  convert)
    # parse --from --to --input --output from "$@"
    my-tool "$INPUT" "$OUTPUT"
    ;;
esac
```

Name it `convertr-myplugin`, place it in your `$PATH`, and `convertr plugins list` will find it immediately.

See [`pkg/plugin/protocol.go`](./pkg/plugin/protocol.go) for the Go type definitions.

---

## Progress reporting

convertr picks a reporter automatically based on the environment:

| Environment | Reporter |
|-------------|----------|
| Interactive TTY | In-place progress bar (`[3/10] converting report.md`) |
| Non-TTY / pipe | Plain text, one line per job |
| `--json` or `CI=1` | JSON Lines |

### JSON Lines format

```json
{"event":"start",   "total":10, "ts":"2026-04-17T12:00:00Z"}
{"event":"convert", "file":"report.md", "done":1, "total":10, "status":"ok",    "ts":"..."}
{"event":"convert", "file":"broken.doc","done":2, "total":10, "status":"error", "error":"exit status 1", "ts":"..."}
{"event":"done",    "ts":"2026-04-17T12:00:05Z"}
```

JSON output is never localised — it is a stable scripting contract.

---

## Architecture

```
cmd/convertr/          cobra entry point, blank imports for all backends

internal/
  cli/                 subcommand implementations (convert, watch, doctor, …)
  backend/             Backend interface, Options, registry, execx helper
    backends/
      pandoc/          pandoc backend
      ffmpeg/          ffmpeg backend
      imagemagick/     ImageMagick backend
      libreoffice/     LibreOffice backend (process-isolated)
      jq/              jq backend
      yq/              yq backend
      tesseract/       Tesseract OCR backend
      csvkit/          CSVKit backend
      asciidoctor/     AsciiDoctor backend
      figlet/          figlet backend
      textutil/        textutil backend (darwin only)
      plugin/          external plugin discovery and execution
  config/              TOML loader, env overrides, profile merge, FieldSource
  formats/             Format registry (~50 formats), extension/MIME detection
  router/              Dijkstra routing on the capability graph
  runner/              Job execution: serial + parallel pool, retry, conflict resolution
  sink/                Output path resolution, atomic write, conflict policy, template
  source/              Input iterators: file, glob, dir, stdin (iter.Seq2)
  watch/               fsnotify wrapper with debounce, recursive add, delete handling
  progress/            Reporter interface: TUI, plain, JSON Lines, Noop
  i18n/                go-i18n v2, locales/en.json + ru.json
  slogx/               slog initialisation, JSON/text handlers, verbosity levels
  xdg/                 XDG Base Directory paths

pkg/plugin/            Public plugin protocol types (for plugin authors)
```

### Key invariants

- **Atomic write** — all output goes to `os.CreateTemp` then `os.Rename`. Never written directly.
- **TempDir per job** — `os.MkdirTemp("", "convertr-*")` per conversion chain, removed with `defer os.RemoveAll`.
- **LibreOffice isolation** — every soffice call gets `--env:UserInstallation=file:///tmp/convertr-lo-PID` to allow parallel execution.
- **Context propagation** — all backends receive `ctx` and use `exec.CommandContext`.
- **Exit codes** — `0` success · `1` error · `2` CLI usage · `3` no route · `4` partial batch · `5` missing backend.

---

## Troubleshooting

**`cannot determine target format`**
Provide `--to FORMAT` or give the output file a known extension (e.g., `-o out.pdf`).

**`no conversion route`**
No installed backend covers the requested From → To pair. Run `convertr doctor` to see what is missing, and install the relevant tool.

**`convertr doctor` shows a backend as MISSING**
Install it via the suggested command. On macOS run `./scripts/install-deps-macos.sh` to install everything at once.

**LibreOffice hangs or produces a corrupt output**
Make sure no other soffice process is running with the same `UserInstallation` directory. Parallel convertr jobs each get a unique directory; external LibreOffice processes may conflict.

**Watch mode does not pick up changes inside a new subdirectory**
convertr watches new directories recursively — but only directories created *after* watch starts that trigger a `CREATE` event. Re-start watch if you have pre-existing deep trees.

**Output file is not where I expected**
When `--output` is a directory, the output file is named `<stem>.<target-ext>`. Use `--on-conflict rename` if you are converting multiple files with the same stem.

**Plugin not found by `convertr plugins list`**
Ensure the binary is named `convertr-NAME` (not `convertrNAME`), is executable (`chmod +x`), and its directory is in `$PATH`.

**Enable verbose logging**
```sh
convertr -vvv doctor
CI=1 convertr -r ./src/ -o ./out/ --to md   # structured JSON logs
```

---

## License

MIT — see [LICENSE](./LICENSE).
