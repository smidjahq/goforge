# Contributing to goforge

Thank you for your interest in contributing! This guide is written so that a Go developer who has never seen this codebase can add a new framework, database combo, logger, or extra by following it alone.

## Table of contents

- [Project structure](#project-structure)
- [How to add a new framework](#how-to-add-a-new-framework)
- [How to add a new database combo](#how-to-add-a-new-database-combo)
- [How to add a new logger](#how-to-add-a-new-logger)
- [How to add a new extra](#how-to-add-a-new-extra)
- [Running the test suite](#running-the-test-suite)
- [PR checklist](#pr-checklist)

---

## Project structure

```
cmd/goforge/          CLI entry point and Cobra commands
internal/
  config/             Shared Config struct
  validator/          Combination constraint engine (single source of truth)
  generator/          Template renderer + embed
    templates/
      base/           Files every project gets
      frameworks/     One subdirectory per HTTP framework
      db/             One subdirectory per DB combo
      loggers/        One subdirectory per logger
      extras/         One subdirectory per optional extra
  postgen/            Post-generation steps (go mod tidy, git init)
  prompts/            Interactive huh TUI
```

The generator walks the active layers in order — `base → framework → db → logger → extras` — and writes each file to the output directory. Later layers overwrite earlier ones when filenames conflict. Files ending in `.tmpl` are rendered with `text/template` and written without the `.tmpl` suffix; all other files are copied verbatim.

Template data is the `config.Config` struct:

```go
type Config struct {
    Name      string   // project directory name
    Module    string   // Go module path, e.g. "github.com/you/myapp"
    Framework string   // "gin" | "chi" | ...
    DB        string   // "postgres-gorm" | "none" | ...
    Logger    string   // "slog" | "zap" | "zerolog"
    Extras    []string // "docker", "makefile", ...
}
```

---

## How to add a new framework

**Example:** adding `echo`.

### 1. Create the template folder

```
internal/generator/templates/frameworks/echo/
  internal/app/app.go.tmpl
  internal/controller/http/router.go.tmpl
```

`router.go.tmpl` must define `package controller` and export `NewRouter() http.Handler`. It must include a `GET /healthz` route that returns 200. Look at `templates/frameworks/gin/` for a reference implementation.

`app.go.tmpl` must import `"{{.Module}}/pkg/logger"` and call `logger.New()` on startup.

### 2. Register the framework in the validator

Open `internal/validator/validator.go` and add `"echo"` to `validFrameworks`:

```go
var validFrameworks = []string{"gin", "chi", "echo"}
```

### 3. Add it to the TUI prompt

Open `internal/prompts/prompt.go` and add an option to the framework select:

```go
huh.NewOption("Echo", "echo"),
```

### 4. Add integration tests

In `internal/generator/frameworks_test.go`, add a subtest to `TestGenerate_FrameworkFileTree` and `TestGenerate_FrameworkBuilds` for `"echo"`.

### 5. Verify

```bash
go test ./...
go vet ./...
```

---

## How to add a new database combo

**Example:** adding `mysql-gorm`.

### 1. Update the validator

Open `internal/validator/validator.go`. Add `"mysql"` as a backend key in `validCombinations`:

```go
var validCombinations = map[string][]string{
    "postgres": {"gorm", "sqlc", "raw"},
    "sqlite":   {"gorm", "raw"},
    "mysql":    {"gorm", "raw"},
    "none":     {},
}
```

### 2. Create the template folder

```
internal/generator/templates/db/mysql-gorm/
  internal/repo/persistent/repo.go.tmpl
  docker-compose.yml.tmpl
```

`repo.go.tmpl` must define `package persistent` with a `New(dsn string)` constructor that returns `(*Repo, error)`. See `templates/db/postgres-gorm/` for reference.

If the database needs a Docker service, add `docker-compose.yml.tmpl`.

### 3. Update go.mod.tmpl

Open `templates/base/go.mod.tmpl` and add the driver dependency inside the appropriate `{{if}}` block:

```
{{- if eq .DB "mysql-gorm"}}
    github.com/go-sql-driver/mysql v1.8.1
{{- end}}
```

### 4. Add it to the TUI prompt

Open `internal/prompts/prompt.go`. Add `"mysql"` to the backend select and ensure `layerOptions` returns valid layer options for it (it reads from `validator.ValidOptionsFor`, so this is automatic once the validator is updated).

### 5. Add integration tests

In `internal/generator/db_test.go`, add a subtest to `TestGenerate_DBFileTree` and `TestGenerate_DBBuilds` for `"mysql-gorm"`.

### 6. Verify

```bash
go test ./...
go vet ./...
```

---

## How to add a new logger

**Example:** adding `logrus`.

### 1. Register in the validator

Open `internal/validator/validator.go` and add `"logrus"` to `validLoggers`:

```go
var validLoggers = []string{"slog", "zap", "zerolog", "logrus"}
```

### 2. Create the template folder

```
internal/generator/templates/loggers/logrus/
  pkg/logger/impl.go.tmpl
```

`impl.go.tmpl` must define `package logger` and provide:

```go
func New() Logger { ... }
```

The returned type must satisfy the `Logger` interface defined in `templates/base/pkg/logger/logger.go.tmpl`:

```go
type Logger interface {
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    Debug(msg string, args ...any)
}
```

### 3. Update go.mod.tmpl

Open `templates/base/go.mod.tmpl` and add the dependency:

```
{{- if eq .Logger "logrus"}}
    github.com/sirupsen/logrus v1.9.3
{{- end}}
```

### 4. Add it to the TUI prompt

Open `internal/prompts/prompt.go` and add an option to the logger select:

```go
huh.NewOption("Logrus (sirupsen/logrus)", "logrus"),
```

### 5. Add integration tests

In `internal/generator/loggers_test.go`, add `"logrus"` to the logger loop in `TestGenerate_LoggerImplFileExists` and `TestGenerate_LoggerBuilds`. Add a case to `TestGenerate_LoggerGoModContainsDep` for `"logrus"` → `"github.com/sirupsen/logrus"`.

### 6. Verify

```bash
go test ./...
go vet ./...
```

---

## How to add a new extra

**Example:** adding `devcontainer`.

Extras are **purely additive** — they must never modify files produced by other layers. Only add new files.

### 1. Create the template folder

```
internal/generator/templates/extras/devcontainer/
  .devcontainer/devcontainer.json
```

Files may be `.tmpl` (rendered) or plain (copied). No file in this folder should have the same relative path as a file produced by `base/`, `frameworks/`, `db/`, or `loggers/`.

### 2. Register in the validator

Open `internal/validator/validator.go` and add `"devcontainer"` to `validExtras`:

```go
var validExtras = []string{"docker", "makefile", "ci", "swagger", "migrations", "linter", "devcontainer"}
```

If the extra has a constraint (like `migrations` requires a DB), add the validation logic in `Validate`.

### 3. Add it to the TUI prompt

Open `internal/prompts/prompt.go` and add an option to `extrasOptions()`:

```go
huh.NewOption("Dev Container (.devcontainer/)", "devcontainer"),
```

### 4. Add integration tests

In `internal/generator/extras_test.go`, add a test function (e.g. `TestGenerate_DevcontainerExtra`) and add `"devcontainer"` to the `mustExist` list in `TestGenerate_AllExtras`.

### 5. Verify

```bash
go test ./...
go vet ./...
```

---

## Running the test suite

```bash
# Fast: skips compilation tests that require network access
go test ./... -short

# Full: runs go mod tidy + go build on generated projects (needs network)
go test ./...

# Lint
go vet ./...
```

The compilation tests (`TestGenerate_*Builds`, `TestRun_GoModTidy`) call `go mod tidy` and `go build ./...` on real temp directories. They are skipped with `-short` so CI can run them separately from unit tests if needed.

---

## PR checklist

- [ ] All acceptance criteria from the linked issue are met
- [ ] `go test ./... -short` passes
- [ ] `go vet ./...` passes
- [ ] New templates include generator integration tests
- [ ] New framework/DB/logger/extra is registered in `internal/validator/`
- [ ] New framework/DB/logger/extra is wired into the TUI in `internal/prompts/prompt.go`
- [ ] `CONTRIBUTING.md` updated if a new contribution path was added
- [ ] README stack matrix updated if a new stack option was added
