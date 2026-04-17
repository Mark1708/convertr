# convertr — план разработки

Универсальный CLI-оркестратор конвертации файлов поверх pandoc, ffmpeg, LibreOffice,
ImageMagick, jq, yq, tesseract и других бэкендов. Маршрутизация через граф форматов (BFS).

Модуль: `git.mark1708.ru/me/convertr`
Remote: `ssh://git@git.mark1708.ru:2224/me/convertr.git`

---

## Статус фаз

| Фаза | Статус | Описание |
|------|--------|----------|
| 1    | ✅ done | Скелет: CLI, формат-регистр, backend-интерфейс, router-граф |
| 2    | ⬜ todo | Ядро маршрутизации: runner, sink, source |
| 3    | ⬜ todo | Бэкенды документов + doctor/formats |
| 4    | ⬜ todo | Бэкенды данных и медиа |
| 5    | ⬜ todo | Batch, Watch, прогресс |
| 6    | ⬜ todo | Конфиг TOML, плагины, дистрибуция |

---

## Фаза 2 — Ядро маршрутизации

### internal/runner

```
internal/runner/
├── runner.go    — Plan → Execute: обходит Route.Steps, управляет tempdir
├── job.go       — Job{Source, Route, Sink}, Stage (in/out path), TempDir
├── pool.go      — errgroup + semaphore per backend type (Workers)
├── retry.go     — экспоненциальный backoff + jitter
└── report.go    — Summary, Collector (atomic счётчики ok/fail/skip)
```

**Контракт:**
```go
type Job struct {
    Source string
    Route  *router.Route
    Sink   *sink.Sink
    Opts   backend.Options
}
func Execute(ctx context.Context, jobs []Job, opts RunOpts) (*Summary, error)
```

### internal/sink

```
internal/sink/
├── sink.go      — Sink interface: path, format, conflict policy
├── resolve.go   — определяет тип sink из -o PATH (file / dir / stdout)
├── atomic.go    — os.CreateTemp в той же FS → os.Rename
├── template.go  — {stem}, {date}, {seq}, {to} парсер для имён файлов
└── conflict.go  — Resolver: overwrite | skip | rename | error
```

### internal/source

```
internal/source/
├── source.go  — SourceFile{Path, Format, Size, ModTime}
├── iter.go    — iter.Seq2[SourceFile, error]: File / Glob / Dir / Stdin / List
└── filter.go  — include/exclude glob, size min/max, mtime, maxDepth
```

**Контракт (Go 1.23 iter):**
```go
func FileSource(path string) iter.Seq2[SourceFile, error]
func GlobSource(pattern string) iter.Seq2[SourceFile, error]
func DirSource(root string, opts DirOpts) iter.Seq2[SourceFile, error]
func StdinSource(format string) iter.Seq2[SourceFile, error]
func Chain(sources ...iter.Seq2[SourceFile, error]) iter.Seq2[SourceFile, error]
```

### Доработка router

- `bfs.go`: убрать мёртвый код `from` переменной, покрыть юнит-тестами
- `graph.go`: `BuildFromBackends(backends []Backend)` — принимать список явно
- DOT-вывод в `cli/formats.go` — подключить реальные рёбра из графа

### cli/convert.go

```
convertr [INPUTS...] -o OUTPUT [--to FORMAT] [--from FORMAT]
                     [--dry-run] [--workers N]
                     [--on-error skip|stop|retry]
                     [--on-conflict overwrite|skip|rename|error]
```

---

## Фаза 3 — Бэкенды документов

```
internal/backend/backends/
├── pandoc/pandoc.go        — doc→formats, md↔html↔rst↔epub↔docx↔pdf↔odt
├── libreoffice/lo.go       — doc→odt, docx→odt, xlsx→csv, pptx→odp
│                             UserInstallation isolation для параллельных запусков
└── textutil/textutil.go    — build tag: darwin; .doc/.rtf → .txt/.html (встроено в macOS)
```

**Capabilities pandoc (примерный набор):**
```
md   → html, docx, odt, pdf, rst, epub, tex
html → md, docx, pdf
docx → md, html, pdf, odt
odt  → md, docx, html
rst  → md, html, docx
epub → md, html
```

**Capabilities LibreOffice:**
```
doc  → odt, docx, pdf, txt
docx → odt, pdf, txt
xlsx → csv, ods
pptx → odp, pdf
```

**doctor** — расширить версионным пробингом:
```
pandoc   2.19.2   OK
soffice  7.6.7.2  OK
```

**formats --dot** — подключить реальные рёбра из `backend.AllCapabilities()`.

---

## Фаза 4 — Бэкенды данных и медиа

```
internal/backend/backends/
├── ffmpeg/ffmpeg.go            — mp4↔mkv↔webm↔mov, mp3↔flac↔aac↔ogg↔wav↔m4a↔opus
│                                  gif←mp4 (palette filter)
├── imagemagick/im.go           — jpg↔png↔webp↔avif↔gif↔tiff↔bmp, svg→png
├── jq/jq.go                    — json→json (transform, минификация/форматирование)
├── yq/yq.go                    — yaml↔json↔toml
├── csvkit/csvkit.go            — csv↔xlsx↔json (csvkit Python или xlsx2csv)
├── asciidoctor/ad.go           — adoc→pdf, adoc→html (asciidoctor-pdf)
├── figlet/figlet.go            — txt→ascii (figlet)
└── tesseract/tesseract.go      — jpg/png/tiff → txt (OCR)
```

**Особенности реализации:**
- ffmpeg: `-y -hide_banner -loglevel error` по умолчанию; Quality → `-crf` / `-b:a`
- ImageMagick: `convert` → `magick` (IM7+), Quality → `-quality`
- Tesseract: `--oem 1 --psm 3 -l rus+eng` по умолчанию
- csvkit: проверить наличие `in2csv` / `csvjson`; fallback на `xlsx2csv`

---

## Фаза 5 — Batch и Watch

### Batch (runner/pool.go)

```go
type RunOpts struct {
    Workers     int           // per backend type (default: GOMAXPROCS)
    OnError     ErrorPolicy   // skip | stop | retry
    OnConflict  ConflictPolicy
    Retry       RetryOpts     // MaxAttempts, BaseDelay, MaxDelay, Multiplier
    DryRun      bool
}
```

- `errgroup.WithContext` + `semaphore.NewWeighted` per backend type
- Retry: экспоненциальный backoff с jitter, только на `ErrConvertFail`
- `--on-error skip`: накапливать ошибки, exit code 4

### Watch (internal/watch)

```
internal/watch/
├── watcher.go    — fsnotify.Watcher + debounce (time.AfterFunc 300ms)
├── recursive.go  — динамическое добавление поддиректорий при CREATE
└── deletion.go   — on-delete: keep | remove | archive
```

**cli/watch.go:**
```
convertr watch SRC -o DST [--to FORMAT] [--debounce 300ms]
                           [--on-delete keep|remove|archive]
```

Graceful shutdown: `os.Signal` → cancel context → drain pool → exit 0.

### Progress (internal/progress)

```
internal/progress/
├── reporter.go  — Reporter interface: Start/Update/Done/Error
├── plain.go     — не-TTY или --quiet: одна строка на файл
├── tui.go       — bubbletea (TTY + batch > 1 файла): прогресс-бар
└── jsonlog.go   — --json / CI=1: JSON-строка per событие
```

---

## Фаза 6 — Конфиг, плагины, дистрибуция

### internal/config

```
internal/config/
├── config.go   — Config struct (TOML):
│                   [defaults] quality, workers, on_error, on_conflict
│                   [profile.NAME] — именованные профили
│                   [backend.pandoc] extra_args = ["--wrap=none"]
├── loader.go   — порядок: hardcoded defaults → file → env (CONVERTR_*) → CLI flags
│                 cmd.Flags().Changed() для определения явных CLI-флагов
└── profile.go  — merge профиля поверх defaults
```

Пример `~/.config/convertr/config.toml`:
```toml
[defaults]
quality    = 85
workers    = 4
on_error   = "skip"

[backend.pandoc]
extra_args = ["--wrap=none", "--standalone"]

[profile.hi-res]
quality = 100
```

### cli/config.go

```
convertr config print    — вывести активный конфиг (с указанием источника каждого поля)
convertr config init     — создать ~/.config/convertr/config.toml с defaults
convertr config validate — проверить синтаксис конфига
```

### Plugin protocol v1

```
internal/backend/backends/plugin/plugin.go  — discovers convertr-* binaries in PATH

Protocol:
  convertr-NAME capabilities           → JSON stdout
  convertr-NAME convert \
    --from FOO --to BAR \
    --input /abs/in.foo \
    --output /abs/out.bar \
    [--opt key=value ...]
```

```
pkg/plugin/protocol.go  — публичный контракт (для авторов плагинов)
```

### cli/plugins.go

```
convertr plugins list     — найти convertr-* в PATH + статус
convertr plugins test     — вызвать capabilities на каждом
convertr plugins install  — (future) установить из реестра
```

### cli/info.go

```
convertr info FILE  — метаданные через ffprobe / exiftool / file(1)
```

### Дистрибуция

- `.goreleaser.yaml` — бинари macOS arm64/amd64 + Linux amd64
- `scripts/install-deps-macos.sh` — `brew install pandoc ffmpeg ...`
- `scripts/install-deps-linux.sh` — `apt-get install ...`
- Homebrew formula `mark1708/tap/convertr`

```makefile
snapshot:
	goreleaser release --snapshot --clean
release:
	goreleaser release --clean
```

---

## Ключевые инварианты

- **Atomic write**: всегда `os.CreateTemp` в той же FS → `os.Rename`. Никогда не писать напрямую в целевой файл.
- **TempDir per job**: создавать в `os.TempDir()`, удалять в defer даже при панике.
- **LibreOffice isolation**: `--env UserInstallation=file:///tmp/convertr-lo-PID` для каждого процесса.
- **Context propagation**: все backends получают `ctx`, отменяют subprocess через `exec.CommandContext`.
- **Exit codes**: 0 success | 1 error | 2 CLI usage | 3 no route | 4 partial batch | 5 missing backend.

---

## CLI контракт (целевой)

```bash
# Одиночный файл
convertr report.doc -o report.md

# Batch с сохранением структуры директорий
convertr -r ./docs/ -o ./out/ --to md --preserve-structure

# Stdin → stdout
cat table.csv | convertr --from csv --to json --stdout

# Watch
convertr watch ./inbox -o ./outbox --to md

# Dry-run с маршрутом
convertr report.doc -o report.pdf --dry-run

# DOT-граф форматов
convertr formats --dot | dot -Tsvg > formats.svg

# Метаданные
convertr info video.mp4
```
