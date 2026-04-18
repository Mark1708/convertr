# Changelog

All notable changes to convertr are documented here.

**Public API** — the following are considered stable and subject to semantic versioning:
- CLI flags and subcommand names
- Exit codes (0 success · 1 error · 2 CLI usage · 3 no route · 4 partial batch · 5 missing backend)
- Plugin protocol (`convertr-NAME capabilities` / `convertr-NAME convert`)
- `pkg/plugin` Go types

Internal packages (`internal/`) are not part of the public API.

---

## [Unreleased]

### Fixed
- Pandoc backend now passes explicit `--from`/`--to` flags on every call, preventing format misdetection when the file extension doesn't match actual content (e.g. HTML saved as `.doc`)
- i18n language detection from `$LANG` environment variable now persists through command execution (previously reset to English by `PersistentPreRunE`)

### Added
- `--mkdir` flag: automatically create the output directory without prompting
- Multi-input directory detection: when multiple input files are given, the output path is treated as a directory even without a trailing `/`
- Interactive `Create it? [y/N]` prompt when the output directory is missing and running in a TTY
- `--named key=value` flag (repeatable): pass backend-specific named options from the CLI without using `extra_args`
- `--strip-meta` flag: strip file metadata (EXIF, ID3, document properties) — supported by imagemagick (`-strip`), ffmpeg (`-map_metadata -1`), pandoc (`--strip-comments`)
- **i18n completeness**: all CLI flag descriptions, error messages, subcommand descriptions, and progress output are now fully translated into English and Russian
- **Backend named options** — new per-backend `--named` keys:
  - `ffmpeg`: `ffmpeg.video_codec`, `ffmpeg.audio_codec`, `ffmpeg.fps`, `ffmpeg.audio_rate`, `ffmpeg.gif_fps`, `ffmpeg.gif_scale`
  - `pandoc`: `pandoc.toc`, `pandoc.standalone`, `pandoc.highlight`, `pandoc.template`, `pandoc.pdf_engine`
  - `imagemagick`: `imagemagick.resize`, `imagemagick.density`, `imagemagick.depth`
  - `figlet`: `figlet.font` (previously hardcoded to `standard`)
  - `yq`: `yq.sort_keys`, `yq.indent`
  - `csvkit`: `csvkit.sheet`, `csvkit.delimiter`
  - `asciidoctor`: `asciidoctor.toc`, `asciidoctor.attribute`
- `--workers` / `-j` value is now forwarded to ffmpeg as `-threads` for multi-threaded encoding

---

## [0.1.0] — 2026-04-17

Initial release.

### Added

**Core pipeline**
- Dijkstra routing through the backend capability graph (maxHops = 4)
- Atomic writes via `os.CreateTemp` + `os.Rename`
- Per-job temp directories, parallel pool via `errgroup` + `semaphore`
- Error policies: `skip` | `stop` | `retry` (exponential backoff with jitter)
- Conflict policies: `overwrite` | `skip` | `rename` | `error`
- Input sources: file, glob, directory (recursive), stdin

**Backends**
- `pandoc` — document and markup formats (md, html, docx, odt, rst, epub, tex, org, pdf)
- `ffmpeg` — video ↔ video, audio ↔ audio, video → gif, video → audio
- `imagemagick` — raster images, svg, heic, avif
- `libreoffice` — Office formats with per-process isolation
- `jq` — JSON pretty-print, minify, transform
- `yq` — YAML ↔ JSON ↔ TOML
- `tesseract` — OCR: jpg/png/tiff → txt
- `csvkit` — xlsx ↔ csv, csv → json
- `asciidoctor` — adoc → html, adoc → pdf
- `figlet` — txt → ascii art
- `textutil` — doc/rtf → txt/html (macOS only)
- `plugin` — external `convertr-*` executables via capabilities protocol

**Commands**
- `convertr FILE -o OUT` — convert files (default command)
- `convertr doctor` — check backend availability and versions
- `convertr formats [--dot]` — list formats and conversion graph
- `convertr watch SRC -o DST --to FORMAT` — watch mode with debounce and delete policies
- `convertr config print|init|validate` — configuration management
- `convertr plugins list|test` — plugin discovery and testing
- `convertr info FILE` — file metadata via ffprobe/exiftool/file
- `convertr version` — print version

**Configuration**
- TOML config at `~/.config/convertr/config.toml`
- `[defaults]`, `[backend.NAME]`, `[profile.NAME]` sections
- `CONVERTR_*` environment variable overrides
- Named profiles via `--profile`

**Progress reporting**
- In-place TUI progress bar (interactive TTY)
- Plain text one-line-per-job (non-TTY)
- JSON Lines (`--json` / `CI=1`)

**i18n**
- English and Russian locales bundled
- `--lang` flag and auto-detection from environment

**Distribution**
- goreleaser config for macOS (arm64/amd64) and Linux (amd64)
- Homebrew tap: `mark1708/tap/convertr`
- `scripts/install-deps-macos.sh` and `scripts/install-deps-linux.sh`
