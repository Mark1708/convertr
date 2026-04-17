# Contributing to convertr

## Package layout

```
cmd/convertr/          entry point, backend blank imports
internal/
  cli/                 subcommand implementations
  backend/             Backend interface, Options, registry, execx helper
    backends/          one package per backend binary
  config/              TOML config, env overrides, profiles
  formats/             format registry, MIME/extension detection
  router/              Dijkstra routing
  runner/              job execution, pool, retry, conflict resolution
  sink/                output path resolution, atomic write
  source/              input iterators (iter.Seq2)
  watch/               fsnotify wrapper with debounce
  progress/            Reporter interface and implementations
  i18n/                locales/en.json + ru.json
  slogx/               slog initialisation
  xdg/                 XDG Base Directory paths
pkg/plugin/            public plugin protocol types
```

## Architectural rules

- **Atomic writes only.** Every output file goes through `os.CreateTemp` + `os.Rename`. Never write directly to the final path.
- **TempDir per job.** Allocate with `os.MkdirTemp("", "convertr-*")`; always `defer os.RemoveAll`.
- **Context everywhere.** All backends receive `ctx` and use `exec.CommandContext`. Never ignore a cancelled context.
- **LibreOffice isolation.** Pass `--env:UserInstallation=file:///tmp/convertr-lo-<PID>` on every `soffice` invocation.
- **Register via init().** Each backend package calls `backend.Register(...)` from `init()`. Wire it in `cmd/convertr/main.go` with a blank import.

## Running tests

```sh
go test -race ./...

# Single package
go test -race ./internal/router/...
```

## Linting

```sh
golangci-lint run ./...
```

## Adding a new backend

1. Create `internal/backend/backends/<name>/<name>.go`.
2. Implement the `backend.Backend` interface (`Name`, `BinaryName`, `Capabilities`, `Convert`).
3. Call `backend.Register(...)` from `init()`.
4. Add a blank import in `cmd/convertr/main.go`.
5. Add the binary to `scripts/install-deps-macos.sh` and `scripts/install-deps-linux.sh`.

Minimal template:

```go
package mybackend

import (
    "context"
    "git.mark1708.ru/me/convertr/internal/backend"
    "git.mark1708.ru/me/convertr/internal/backend/execx"
)

func init() { backend.Register(Backend{}) }

type Backend struct{}

func (Backend) Name() string       { return "mybackend" }
func (Backend) BinaryName() string { return "mybinary" }

func (Backend) Capabilities() []backend.Capability {
    return []backend.Capability{
        {From: "foo", To: "bar", Cost: 2},
    }
}

func (Backend) Convert(ctx context.Context, in, out string, opts backend.Options) error {
    return execx.Run(ctx, "mybinary", in, "-o", out)
}
```

## Writing a plugin

See [`pkg/plugin/protocol.go`](./pkg/plugin/protocol.go) for the type definitions and the [Plugins section in README.md](./README.md#plugins) for the full protocol.

## Commit style

Conventional Commits format:

```
feat: add wasm backend
fix: router skips unreachable nodes
docs: update backend table
```

The body should explain *why*, not *what*.
