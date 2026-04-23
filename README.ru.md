# convertr

> Универсальный CLI-конвертер для 50+ форматов файлов — один бинарник, множество бэкендов.

[![CI](https://github.com/Mark1708/convertr/actions/workflows/ci.yml/badge.svg)](https://github.com/Mark1708/convertr/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Mark1708/convertr.svg)](https://pkg.go.dev/github.com/Mark1708/convertr)
[![Go Report Card](https://goreportcard.com/badge/github.com/Mark1708/convertr)](https://goreportcard.com/report/github.com/Mark1708/convertr)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

convertr объединяет pandoc, ffmpeg, LibreOffice, ImageMagick, jq, yq, Tesseract и ещё десяток инструментов за единым интерфейсом `convertr FILE -o OUT`. Он автоматически находит кратчайший путь конвертации — вам больше не нужно помнить, какой бинарник обрабатывает какой формат.

**Документы:** MD · DOCX · PDF · ODT · HTML · EPUB · RST · TeX · PPTX  
**Изображения:** JPG · PNG · WebP · SVG · AVIF · HEIC · GIF  
**Видео:** MP4 · MKV · WebM · AVI · MOV  
**Аудио:** MP3 · FLAC · AAC · WAV · OGG  
**Данные:** JSON · YAML · TOML · CSV · XLSX  
**OCR:** изображение → текст через Tesseract

---

## Почему convertr?

Большинство конвертаций требуют разных инструментов: `pandoc` для документов,
`ffmpeg` для видео, `convert` для изображений. convertr объединяет их все —
просто укажите исходный и целевой форматы, и он сам выберет нужный бэкенд,
при необходимости объединяя несколько инструментов в цепочку.

- Не нужно помнить форматы — просто укажите источник и цель
- Пакетная конвертация с параллельными воркерами и политиками повторов
- Режим наблюдения: автоконвертация при изменении файла
- Протокол плагинов для собственных бэкендов

---

## Содержание

- [Установка](#установка)
- [Быстрый старт](#быстрый-старт)
- [Конфигурация — config.toml](#конфигурация--configtoml)
- [Справочник CLI](#справочник-cli)
- [Бэкенды](#бэкенды)
- [Режим наблюдения](#режим-наблюдения)
- [Плагины](#плагины)
- [Отображение прогресса](#отображение-прогресса)
- [Архитектура](#архитектура)
- [Устранение неполадок](#устранение-неполадок)
- [Лицензия](#лицензия)

---

## Установка

### `go install`

```sh
go install github.com/Mark1708/convertr/cmd/convertr@latest
```

Требуется Go 1.25+.

### Homebrew

```sh
brew install mark1708/tap/convertr
```

### Из исходников

```sh
git clone https://github.com/Mark1708/convertr.git
cd convertr
go build -o ~/.local/bin/convertr ./cmd/convertr
```

### Установка зависимостей бэкендов

```sh
# macOS
./scripts/install-deps-macos.sh

# Debian / Ubuntu
./scripts/install-deps-linux.sh
```

### Проверка

```sh
convertr version
convertr doctor
```

`doctor` проверяет каждый бэкенд, выводит обнаруженную версию и предлагает команды установки для отсутствующих инструментов.

---

## Быстрый старт

```sh
# Конвертировать один файл — формат определяется по расширению.
convertr report.docx -o report.md

# Явно указать целевой формат.
convertr notes.md --to pdf -o notes.pdf

# Пакетная конвертация всей директории: один Markdown на каждый DOCX.
convertr -r ./docs/ -o ./out/ --to md

# Параллельная пакетная конвертация с 4 воркерами.
convertr -r ./inbox/ -o ./outbox/ --to pdf -j 4

# Stdin → stdout (удобно в пайпах).
cat data.json | convertr - --from json --to yaml -o -

# Пробный запуск — показать запланированные конвертации без выполнения.
convertr -r ./docs/ -o ./out/ --to md --dry-run

# Наблюдать за директорией и конвертировать при каждом изменении.
convertr watch ./inbox -o ./outbox --to md

# Показать все известные форматы и граф конвертации.
convertr formats
convertr formats --dot | dot -Tsvg > /tmp/formats.svg
```

---

## Конфигурация — config.toml

Расположение: `~/.config/convertr/config.toml` (или где указывает `--config` / `$CONVERTR_CONFIG`).

Создать файл по умолчанию:

```sh
convertr config init
```

### Полный пример

```toml
# ~/.config/convertr/config.toml

[defaults]
quality     = 85       # 0–100; интерпретируется каждым бэкендом по-своему
workers     = 0        # 0 = GOMAXPROCS; > 0 = фиксированное количество
on_error    = "skip"   # skip | stop | retry
on_conflict = "overwrite" # overwrite | skip | rename | error

# Шрифты для PDF-бэкендов (pandoc с xelatex/lualatex).
# `convertr config init` подставляет значения под текущую ОС; пустые
# поля откатываются к встроенным дефолтам платформы.
[fonts]
mainfont = "PT Serif"
monofont = "Menlo"
sansfont = "Helvetica Neue"

# Дополнительные аргументы для конкретных бэкендов.
[backend.pandoc]
extra_args = ["--wrap=none", "--variable=lang:ru"]

[backend.ffmpeg]
extra_args = ["-preset", "fast"]

[backend.tesseract]
extra_args = ["--dpi", "300"]

# Именованные профили — активируются через --profile NAME.
[profile.hi-res]
quality = 100

[profile.ci]
workers     = 4
on_error    = "stop"
on_conflict = "error"

[profile.ocr]
quality = 95
```

### Поля defaults

| Поле | Тип | Умолчание | Описание |
|------|-----|-----------|----------|
| `quality` | int | 85 | Подсказка качества конвертации (0–100) |
| `workers` | int | 0 | Параллельных воркеров; 0 = GOMAXPROCS |
| `on_error` | string | `skip` | Что делать при ошибке задания |
| `on_conflict` | string | `overwrite` | Что делать при конфликте имён выходных файлов |

### Переменные окружения

| Переменная | Переопределяет |
|------------|----------------|
| `CONVERTR_QUALITY` | `defaults.quality` |
| `CONVERTR_WORKERS` | `defaults.workers` |
| `CONVERTR_ON_ERROR` | `defaults.on_error` |
| `CONVERTR_ON_CONFLICT` | `defaults.on_conflict` |

### Порядок приоритетов

```
встроенные умолчания → файл конфигурации → переменные окружения → флаги CLI
```

Флаги CLI всегда побеждают. Отсутствие файла конфигурации не является ошибкой.

---

## Справочник CLI

### Глобальные флаги

Доступны во всех подкомандах:

| Флаг | Короткий | Умолчание | Описание |
|------|----------|-----------|----------|
| `--config` | | `~/.config/convertr/config.toml` | Путь к файлу конфигурации |
| `--profile` | | | Активировать именованный профиль |
| `--lang` | | | Язык интерфейса (`en`, `ru`) |
| `--verbose` | `-v` | 0 | Уровень детализации логов (можно повторять: `-vvv`) |
| `--quiet` | `-q` | false | Подавить весь вывод кроме ошибок |
| `--json` | | false | Структурированные JSON-логи |
| `--no-color` | | false | Отключить ANSI-цвета |

---

### Конвертация (команда по умолчанию)

`convertr FILE [FILE...] -o OUTPUT [флаги]`

Если подкоманда не указана, convertr считает аргументы входными файлами и запускает конвейер конвертации.

**Флаги:**

| Флаг | Короткий | Умолчание | Описание |
|------|----------|-----------|----------|
| `--output` | `-o` | | Выходной файл, директория или `-` для stdout |
| `--to` | | | Целевой формат (например, `md`, `pdf`, `mp3`) |
| `--from` | | | Переопределение исходного формата |
| `--dry-run` | | false | Показать план без выполнения |
| `--workers` | `-j` | 1 | Параллельных воркеров |
| `--on-error` | | `skip` | Политика ошибок: `skip` \| `stop` \| `retry` |
| `--on-conflict` | | `overwrite` | Политика конфликтов: `overwrite` \| `skip` \| `rename` \| `error` |
| `--recursive` | `-r` | false | Рекурсивно обходить директории |
| `--mkdir` | | false | Создать выходную директорию если она не существует |

**Типы ввода:**

```sh
convertr file.md -o out.pdf                   # один файл
convertr a.md b.md -o out/ --mkdir            # несколько файлов → выходная директория (создаётся если нет)
convertr "src/**/*.md" -o out/ --to pdf       # glob
convertr -r ./src/ -o ./out/ --to html        # директория (рекурсивно)
convertr - --from json --to yaml -o -         # stdin → stdout
```

**Определение выходной директории:**

При нескольких входных файлах выходной путь всегда трактуется как директория — даже без завершающего `/`. Если директория не существует, convertr:

- спросит интерактивно `Create it? [y/N]` в TTY, или
- вернёт ошибку с подсказкой использовать `--mkdir` в неинтерактивном режиме.

```sh
# Автоматически создать выходную директорию:
convertr -r ./docs/ -o ./out --to md --mkdir

# Интерактивный промпт в TTY (--mkdir не нужен):
convertr -r ./docs/ -o ./out --to md
# Output directory does not exist: ./out
# Create it? [y/N]
```

**Политики ошибок:**

| Политика | Поведение |
|----------|-----------|
| `skip` | Зафиксировать ошибку, продолжить остальные задания |
| `stop` | Прервать все задания при первой ошибке |
| `retry` | Повтор с экспоненциальной задержкой (максимум 3 попытки) |

**Политики конфликтов:**

| Политика | Поведение |
|----------|-----------|
| `overwrite` | Заменить существующий выходной файл |
| `skip` | Оставить существующий файл без изменений |
| `rename` | Добавить числовой суффикс (`.1`, `.2`, …) |
| `error` | Завершить с ошибкой, если файл существует |

---

### doctor

```sh
convertr doctor
```

Проверяет каждый бэкенд: определяет установленную версию, показывает полный путь и выводит подсказку по установке для отсутствующих инструментов.

---

### formats

```sh
convertr formats
convertr formats --dot | dot -Tsvg > graph.svg
```

`--dot` выводит граф Graphviz DOT со всеми доступными рёбрами конвертации (с цветовой кодировкой по категориям форматов).

---

### watch

```sh
convertr watch SRC -o DST --to FORMAT [флаги]
```

Наблюдает за `SRC` рекурсивно и конвертирует каждый новый или изменённый файл в `DST`.

**Флаги:**

| Флаг | Умолчание | Описание |
|------|-----------|----------|
| `--output` / `-o` | | Выходная директория (обязательно) |
| `--to` | | Целевой формат (обязательно) |
| `--from` | | Переопределение исходного формата |
| `--debounce` | `300ms` | Ожидание после последнего события перед конвертацией |
| `--on-delete` | `keep` | Что делать при удалении источника: `keep` \| `remove` \| `archive` |

Подробности — в разделе [Режим наблюдения](#режим-наблюдения).

---

### config

```sh
convertr config print      # Показать активную конфигурацию с источниками значений
convertr config init       # Создать файл конфигурации по умолчанию
convertr config validate   # Проверить синтаксис TOML
```

`config print` показывает каждое поле, его текущее значение и откуда оно взято (`default` / `file` / `env` / `flag`).

---

### plugins

```sh
convertr plugins list   # Список convertr-* исполняемых файлов в PATH
convertr plugins test   # Проверить capabilities каждого плагина
```

Протокол описан в разделе [Плагины](#плагины).

---

### info

```sh
convertr info FILE
```

Определяет формат файла, затем извлекает метаданные с помощью первого доступного инструмента: `ffprobe` (аудио/видео), `exiftool` (изображения/документы), `file` (запасной вариант).

---

### version

```sh
convertr version
```

---

## Бэкенды

Каждый бэкенд — это зарегистрированный через `init()` плагин, связывающий один или несколько внешних бинарников с роутером.

### Pandoc

**Бинарник:** `pandoc` · `brew install pandoc` / `apt install pandoc`

| Из | В |
|----|---|
| `md` | `html` `docx` `odt` `pdf` `rst` `epub` `tex` `txt` `typst` `ipynb` `pptx` `mediawiki` `jira` `opml` |
| `html` | `md` `docx` `pdf` `txt` |
| `docx` | `md` `html` `pdf` `odt` `txt` `rst` |
| `odt` | `md` `docx` `html` `txt` |
| `rst` | `md` `html` `pdf` `txt` |
| `epub` | `md` `html` `txt` |
| `tex` | `md` `html` `pdf` |
| `org` | `md` `html` `pdf` |
| `typst` | `md` `pdf` |
| `ipynb` | `md` `html` `pdf` |
| `rtf` | `md` (и далее в любой markup через pandoc-цепочку) |
| `fb2` | `md` `html` `epub` |
| `mediawiki` `dokuwiki` `jira` `textile` `docbook` `opml` | `md` |
| `bibtex` ↔ `csljson` | (библиография) |

PDF-вывод выбирает `xelatex` → `lualatex` → `pdflatex` в порядке убывания
возможностей; если выбран fontspec-совместимый движок, convertr подставляет
шрифты из секции `[fonts]` конфига (с разумными дефолтами на каждой ОС) и
поля `-V geometry:margin=2cm`. Любая переменная, явно указанная через
`--named pandoc.mainfont=...` или `-V ...` в `extra_args`, отключает
соответствующий дефолт — пользовательские значения имеют приоритет.

**Конфиг:** `[backend.pandoc]` `extra_args = ["--wrap=none"]`

---

### FFmpeg

**Бинарник:** `ffmpeg` · `brew install ffmpeg` / `apt install ffmpeg`

| Из | В |
|----|---|
| `mp4` `mkv` `webm` `mov` `avi` | (все видеоформаты, двунаправленно) |
| `mp3` `flac` `aac` `ogg` `wav` `m4a` `opus` | (все аудиоформаты, двунаправленно) |
| видеоформаты | `gif` (двухпроходная палитра) |
| видеоформаты | аудиоформаты (извлечение звука) |

Параметр quality → CRF (видео) или битрейт (аудио).

**Конфиг:** `[backend.ffmpeg]` `extra_args = ["-preset", "fast"]`

---

### ImageMagick

**Бинарник:** `magick` (IM7) или `convert` (IM6) · `brew install imagemagick`

| Из | В |
|----|---|
| `jpg` `png` `webp` `gif` `tiff` `bmp` | (все растровые форматы, двунаправленно) |
| `avif` | растровые форматы (и обратно) |
| `svg` | растровые форматы, `avif` |
| `heic` | растровые форматы |

Параметр quality → `-quality N`.

---

### LibreOffice

**Бинарник:** `soffice` · `brew install --cask libreoffice` / `apt install libreoffice`

| Из | В |
|----|---|
| `doc` `docx` `odt` `rtf` | `odt` `docx` `pdf` `txt` |
| `xlsx` `ods` | `csv` `ods` `xlsx` |
| `pptx` `odp` | `odp` `pptx` `pdf` |

Каждая конвертация запускается с изолированной директорией `--env:UserInstallation` для безопасного параллельного выполнения.

---

### jq

**Бинарник:** `jq` · `brew install jq`

| Из | В | Примечание |
|----|---|------------|
| `json` | `json` | форматирование, минификация или трансформация |

Именованные параметры: `jq.minify=true` (компактный вывод), `jq.filter=EXPR` (произвольное выражение).

---

### yq

**Бинарник:** `yq` · `brew install yq`

| Из | В |
|----|---|
| `yaml` | `json` `toml` |
| `json` | `yaml` `toml` |
| `toml` | `yaml` `json` |

---

### Tesseract (OCR)

**Бинарник:** `tesseract` · `brew install tesseract tesseract-lang`

| Из | В |
|----|---|
| `jpg` `png` `tiff` | `txt` |

Язык по умолчанию: `rus+eng`. Переопределение: `[backend.tesseract]` `extra_args = ["-l", "eng"]`.

---

### CSVKit

**Бинарник:** `in2csv` / `xlsx2csv` (запасной) · `pip install csvkit`

| Из | В |
|----|---|
| `xlsx` | `csv` |
| `csv` | `json` |
| `xlsx` | `json` |

---

### AsciiDoctor

**Бинарник:** `asciidoctor` · `gem install asciidoctor asciidoctor-pdf`

| Из | В |
|----|---|
| `adoc` | `html` |
| `adoc` | `pdf` (требуется `asciidoctor-pdf`) |

---

### Figlet (ASCII-арт)

**Бинарник:** `figlet` · `brew install figlet`

| Из | В |
|----|---|
| `txt` | `ascii` |

---

### textutil (только macOS)

**Бинарник:** `textutil` (встроен в macOS)

| Из | В |
|----|---|
| `doc` `rtf` | `txt` `html` |

Доступен только на macOS; исключён на других платформах через build tag.

---

## Режим наблюдения

```sh
convertr watch ./inbox -o ./outbox --to md
convertr watch ./raw   -o ./web   --to html --debounce 500ms --on-delete archive
```

**Как это работает:**

1. Рекурсивно наблюдает за `SRC` через fsnotify.
2. Устраняет дребезг при частых изменениях — ждёт `--debounce` (по умолчанию 300 мс) после последнего события.
3. Определяет формат изменённого файла, находит маршрут, конвертирует, записывает в `DST`.
4. Новые поддиректории добавляются под наблюдение автоматически.

**Политики удаления:**

| Политика | Что происходит при удалении источника |
|----------|---------------------------------------|
| `keep` | Выходной файл остаётся без изменений (по умолчанию) |
| `remove` | Выходной файл удаляется |
| `archive` | Выходной файл перемещается в `DST/.archive/` |

**Корректное завершение:** нажмите Ctrl+C (SIGINT / SIGTERM). convertr завершает текущие конвертации и выходит с кодом 0.

---

## Плагины

convertr поддерживает внешние плагины — любые исполняемые файлы с именем `convertr-*`, найденные в `PATH`.

### Протокол плагина

Плагин должен реализовывать две подкоманды:

#### `convertr-NAME capabilities`

Записывает JSON-массив в stdout и завершается с кодом 0:

```json
[
  { "from": "wasm", "to": "wat",  "cost": 2 },
  { "from": "wat",  "to": "wasm", "cost": 2 }
]
```

Поля:
- `from` / `to` — идентификаторы форматов
- `cost` — стоимость маршрутизации (1–10; меньше = предпочтительнее встроенных бэкендов); по умолчанию 5

#### `convertr-NAME convert`

```sh
convertr-NAME convert \
  --from FROM \
  --to   TO   \
  --input  /tmp/in.wat  \
  --output /tmp/out.wasm \
  [--opt key=value ...]
```

Код выхода 0 при успехе. При ошибке: ненулевой код, одна строка в stderr.

### Написание плагина

```sh
#!/usr/bin/env bash
case "$1" in
  capabilities)
    echo '[{"from":"foo","to":"bar","cost":3}]'
    ;;
  convert)
    # разобрать --from --to --input --output из "$@"
    my-tool "$INPUT" "$OUTPUT"
    ;;
esac
```

Назовите файл `convertr-myplugin`, поместите его в `$PATH`, и `convertr plugins list` сразу его обнаружит.

Типы Go — в [`pkg/plugin/protocol.go`](./pkg/plugin/protocol.go).

---

## Отображение прогресса

convertr автоматически выбирает репортер в зависимости от окружения:

| Окружение | Репортер |
|-----------|----------|
| Интерактивный TTY | Прогресс-бар на месте (`[3/10] converting report.md`) |
| Не-TTY / пайп | Простой текст, одна строка на задание |
| `--json` или `CI=1` | JSON Lines |

### Формат JSON Lines

```json
{"event":"start",   "total":10, "ts":"2026-04-17T12:00:00Z"}
{"event":"convert", "file":"report.md", "done":1, "total":10, "status":"ok",    "ts":"..."}
{"event":"convert", "file":"broken.doc","done":2, "total":10, "status":"error", "error":"exit status 1", "ts":"..."}
{"event":"done",    "ts":"2026-04-17T12:00:05Z"}
```

JSON-вывод никогда не локализуется — это стабильный контракт для скриптинга.

---

## Архитектура

```
cmd/convertr/          точка входа cobra, blank-импорты всех бэкендов

internal/
  cli/                 реализации подкоманд (convert, watch, doctor, …)
  backend/             интерфейс Backend, Options, реестр, хелпер execx
    backends/
      pandoc/          бэкенд pandoc
      ffmpeg/          бэкенд ffmpeg
      imagemagick/     бэкенд ImageMagick
      libreoffice/     бэкенд LibreOffice (изолированные процессы)
      jq/              бэкенд jq
      yq/              бэкенд yq
      tesseract/       бэкенд Tesseract OCR
      csvkit/          бэкенд CSVKit
      asciidoctor/     бэкенд AsciiDoctor
      figlet/          бэкенд figlet
      textutil/        бэкенд textutil (только darwin)
      plugin/          обнаружение и выполнение внешних плагинов
  config/              загрузчик TOML, переопределения env, слияние профилей, FieldSource
  formats/             реестр форматов (~50 форматов), определение по расширению/MIME
  router/              маршрутизация Дейкстры по графу возможностей
  runner/              выполнение заданий: последовательное + параллельный пул, повтор, разрешение конфликтов
  sink/                разрешение пути вывода, атомарная запись, политика конфликтов, шаблон
  source/              итераторы ввода: файл, glob, директория, stdin (iter.Seq2)
  watch/               обёртка fsnotify с debounce, рекурсивным добавлением, обработкой удаления
  progress/            интерфейс Reporter: TUI, plain, JSON Lines, Noop
  i18n/                go-i18n v2, locales/en.json + ru.json
  slogx/               инициализация slog, обработчики JSON/text, уровни детализации
  xdg/                 пути XDG Base Directory

pkg/plugin/            публичные типы протокола плагинов (для авторов плагинов)
```

### Ключевые инварианты

- **Атомарная запись** — весь вывод идёт в `os.CreateTemp`, затем `os.Rename`. Никогда не пишется напрямую.
- **TempDir на задание** — `os.MkdirTemp("", "convertr-*")` для каждой цепочки конвертации, удаляется через `defer os.RemoveAll`.
- **Изоляция LibreOffice** — каждый вызов soffice получает `--env:UserInstallation=file:///tmp/convertr-lo-PID` для параллельного выполнения.
- **Propagation контекста** — все бэкенды получают `ctx` и используют `exec.CommandContext`.
- **Коды выхода** — `0` успех · `1` ошибка · `2` использование CLI · `3` нет маршрута · `4` частичный батч · `5` отсутствует бэкенд.

---

## Устранение неполадок

**`cannot determine target format`**
Укажите `--to FORMAT` или задайте выходному файлу известное расширение (например, `-o out.pdf`).

**`no conversion route`**
Ни один установленный бэкенд не поддерживает запрошенную пару From → To. Запустите `convertr doctor` чтобы увидеть отсутствующие инструменты и установить их.

**`convertr doctor` показывает бэкенд как MISSING**
Установите его через предложенную команду. На macOS запустите `./scripts/install-deps-macos.sh` для установки всего сразу.

**LibreOffice зависает или создаёт повреждённый файл**
Убедитесь, что нет другого процесса soffice с той же директорией `UserInstallation`. Параллельные задания convertr получают уникальные директории; внешние процессы LibreOffice могут конфликтовать.

**Режим наблюдения не замечает изменения в новой поддиректории**
convertr рекурсивно наблюдает за новыми директориями — но только за теми, что созданы *после* запуска watch и которые вызвали событие `CREATE`. Перезапустите watch если у вас есть существующие глубокие деревья.

**Выходной файл не там, где ожидалось**
Когда `--output` — директория, выходной файл называется `<stem>.<target-ext>`. Используйте `--on-conflict rename` при конвертации нескольких файлов с одинаковым именем.

**Плагин не найден командой `convertr plugins list`**
Убедитесь, что бинарник называется `convertr-NAME` (не `convertrNAME`), является исполняемым (`chmod +x`) и его директория находится в `$PATH`.

**Включить подробное логирование**

```sh
convertr -vvv doctor
CI=1 convertr -r ./src/ -o ./out/ --to md   # структурированные JSON-логи
```

---

## Лицензия

MIT — см. [LICENSE](./LICENSE).
