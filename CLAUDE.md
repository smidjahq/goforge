# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Test (skips compilation tests that shell out to go build/go mod tidy)
go test ./... -short

# Full test suite (shells out to go mod tidy + go build on generated temp projects — needs network)
go test ./...

# Single test
go test ./internal/generator/... -run TestGenerate_GoModContainsModulePath -v

# Vet
go vet ./...

# Build
go build ./...

# Run (interactive TUI)
go run . create

# Run (non-interactive)
go run . create --name myapp --module github.com/you/myapp --framework gin --db postgres-gorm --logger zap

# Dry-run (no files written)
go run . create --dry-run --name myapp --module github.com/you/myapp --framework gin --db postgres-gorm --logger zap

# Version
go run . version
```

## Architecture

goforge is a **layered code generator**. The core idea: a set of template directories ("layers") are walked in a fixed order and merged into a single output directory. Later layers overwrite earlier ones when filenames collide.

### Layer order

```
base → frameworks/{fw} → db/{backend}-{layer} → loggers/{logger} → extras/{extra}…
```

All template files live under `internal/generator/templates/` and are embedded into the binary via `//go:embed`. Files ending in `.tmpl` are rendered with `text/template` (data = `config.Config`); all other files are copied verbatim.

### Package responsibilities (MESE)

| Package | Sole responsibility |
|---------|-------------------|
| `internal/config` | The `Config` struct — shared data contract between all packages |
| `internal/validator` | Single source of truth for every valid stack combination; no I/O |
| `internal/generator` | Walk + render layers into the output directory |
| `internal/prompts` | Interactive huh TUI — converts user input to a `Config` |
| `internal/postgen` | Post-generation shell-outs: `go mod tidy`, `git init` |
| `internal/ui` | Animated braille spinner; no business logic |
| `create.go` | Cobra `create` subcommand — routes to TUI or flag path, orchestrates generator + postgen |
| `version.go` | Cobra `version` subcommand — prints ldflags-injected build metadata |

### Adding a new stack option (framework / DB / logger / extra)

Every new option requires **exactly three touch points**:

1. **Template folder** — add a directory under the appropriate `internal/generator/templates/` subdirectory
2. **Validator** — register the new value in `internal/validator/validator.go`
3. **TUI prompt** — add a `huh.NewOption(...)` in `internal/prompts/prompt.go`

Then add integration tests in the matching `internal/generator/*_test.go` file. See `CONTRIBUTING.md` for step-by-step examples.

### Validation rules worth knowing

- DB value format: `"<backend>-<layer>"` (e.g. `postgres-gorm`) or `"none"`
- SQLite does **not** support `sqlc` — only `gorm` and `raw`
- The `migrations` extra is invalid when `db=none` (enforced in both validator and TUI)
- Echo and Fiber have template folders but are **not** registered in the validator yet

### Test patterns

- All generator tests use `t.TempDir()` for isolation
- Compilation tests (`*Builds`, `TestRun_GoModTidy`) call real `go build`/`go mod tidy` — skip with `-short`
- `create_test.go` uses `--dry-run` to test the full flag → validate → summary path without touching the filesystem
- `version_test.go` and `create_test.go` share a `newRootCmd()` helper so tests exercise the real Cobra wiring
