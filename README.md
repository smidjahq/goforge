# goforge

[![Go Version](https://img.shields.io/badge/go-1.26+-00ADD8?logo=go)](https://go.dev/dl/)
[![CI](https://github.com/smidjahq/goforge/actions/workflows/ci.yml/badge.svg)](https://github.com/smidjahq/goforge/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/smidjahq/goforge)](https://goreportcard.com/report/github.com/smidjahq/goforge)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

> Scaffold production-ready Go backends in seconds — pick your HTTP framework, database, and logger, and get a compilable Clean Architecture project.

<!--
  DEMO GIF — record before v0.1.0:
    vhs demo.tape   (install: brew install vhs)
  Then replace this comment block with: ![goforge demo](demo.gif)
-->

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Stack Options](#stack-options)
- [Generated Project Structure](#generated-project-structure)
- [How It Works](#how-it-works)
- [goforge Project Structure](#goforge-project-structure)
- [Tech Stack](#tech-stack)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

**goforge** is an interactive CLI that generates fully-wired Go backend projects. It walks you through your stack choices (HTTP framework, database layer, logger, and optional extras) via a terminal UI or a single command, then writes a compilable project with Clean Architecture baked in.

Every generated project compiles out of the box — goforge runs `go mod tidy` automatically after generation. No manual dependency wiring required.

---

## Features

- **Interactive TUI** — guided 3-step wizard that enforces valid stack combinations in real time
- **Flag-based / CI mode** — pass all choices as flags, no prompts required
- **Dry-run mode** — preview the resolved config before writing any files
- **2 HTTP frameworks** — Gin and Chi (Echo and Fiber coming soon)
- **6 database combinations** — PostgreSQL and SQLite, each with GORM, sqlc, or raw `database/sql`
- **3 loggers** — `slog` (stdlib), Zap, and Zerolog — all behind a shared `Logger` interface
- **6 optional extras** — Docker, Makefile, GitHub Actions CI, Swagger, Migrations, golangci-lint
- **Clean Architecture layout** — `cmd/`, `internal/`, `pkg/` structured from the start
- **Git init** — optionally initializes a git repository in the generated project
- **Compilable output** — every supported stack combination is integration-tested

---

## Prerequisites

- **Go 1.26+** — [Download](https://go.dev/dl/)
- `$GOPATH/bin` in your `PATH` (see [Installation](#installation))
- CGO enabled if using SQLite (`sqlite-gorm` or `sqlite-raw`)
- Docker (only if you use the `docker` extra)

---

## Installation

```bash
go install github.com/smidjahq/goforge@latest
```

Ensure `$GOPATH/bin` is in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Add that line to your `~/.zshrc` or `~/.bashrc` to make it permanent.

---

## Quick Start

### Interactive (recommended)

```bash
goforge create
```

Launches a TUI wizard that walks you through project name, module path, framework, database, logger, and extras — enforcing valid combinations at each step.

### Flags only (non-interactive / CI)

```bash
goforge create \
  --name myapp \
  --module github.com/you/myapp \
  --framework gin \
  --db postgres-gorm \
  --logger zap \
  --extras docker,makefile,ci,migrations
```

### Dry run

```bash
goforge create \
  --dry-run \
  --name myapp \
  --module github.com/you/myapp \
  --framework gin \
  --db postgres-gorm \
  --logger zap
```

Prints the resolved configuration and exits without writing any files.

### Print version

```bash
goforge version
```

Output:
```
goforge v0.1.0
  commit: abc1234
  built:  2026-03-23
  go:     go1.26.0 darwin/arm64
```

### Skip git init

```bash
goforge create --git=false ...
```

---

## Stack Options

Every combination in the tables below is validated and produces a compilable project.

### HTTP Frameworks

| Framework | Status      |
|-----------|-------------|
| **Gin**   | Supported   |
| **Chi**   | Supported   |
| Echo      | Coming soon |
| Fiber     | Coming soon |

### Databases

| Backend    | Layer                   | `--db` value      |
|------------|-------------------------|-------------------|
| PostgreSQL | GORM                    | `postgres-gorm`   |
| PostgreSQL | sqlc                    | `postgres-sqlc`   |
| PostgreSQL | `database/sql` + pgx    | `postgres-raw`    |
| SQLite     | GORM                    | `sqlite-gorm`     |
| SQLite     | `database/sql`          | `sqlite-raw`      |
| —          | No database             | `none`            |

### Loggers

| Logger      | Package                 | `--logger` value |
|-------------|-------------------------|-----------------|
| **slog**    | stdlib (`log/slog`)     | `slog`          |
| **Zap**     | `go.uber.org/zap`       | `zap`           |
| **Zerolog** | `github.com/rs/zerolog` | `zerolog`       |

All loggers implement the shared `Logger` interface from `pkg/logger/` so the framework layer stays fully decoupled.

### Extras

| `--extras` value | What it adds |
|------------------|--------------|
| `docker`         | Multi-stage `Dockerfile` + app service in `docker-compose.override.yml` |
| `makefile`       | `Makefile` with `run`, `test`, `lint`, `migrate-up/down`, `compose-up/down` targets |
| `ci`             | `.github/workflows/ci.yml` GitHub Actions workflow for the generated project |
| `swagger`        | `docs/` stub for OpenAPI/Swagger annotation |
| `migrations`     | `migrations/` folder + golang-migrate wiring (requires a non-none DB) |
| `linter`         | `.golangci.yml` with sensible golangci-lint defaults |

> **Note:** `migrations` requires `--db` to be set to a non-none value.

---

## Generated Project Structure

```
myapp/
├── cmd/app/
│   └── main.go                    # Entry point — calls app.Run()
├── config/
│   └── config.go                  # os.Getenv-based configuration
├── internal/
│   ├── app/
│   │   └── app.go                 # Wiring: logger, HTTP server, graceful shutdown
│   ├── controller/http/
│   │   ├── router.go              # GET /healthz + /api/v1 routes
│   │   ├── handler/               # HTTP handlers (person, post examples)
│   │   └── middleware/            # CORS + request logging
│   ├── entity/                    # Domain types
│   ├── repo/
│   │   ├── persistent/            # Database adapter (omitted when db=none)
│   │   └── webapi/                # External API adapter
│   └── usecase/                   # Business logic layer
├── pkg/
│   ├── client/                    # HTTP client (no domain knowledge)
│   ├── constants/
│   ├── logger/                    # Logger interface + chosen implementation
│   └── utils/
├── .env.example
├── .gitignore
├── application.yml
└── go.mod
```

Files from the `docker`, `makefile`, `ci`, `swagger`, `migrations`, and `linter` extras are layered on top of this tree.

---

## How It Works

goforge uses a **layered template composition** model. Layers are applied in order, and later layers can override earlier ones:

```
base → framework (gin|chi) → db (postgres-gorm|…) → logger (slog|zap|zerolog) → extras (docker|…)
```

Each layer is a directory of Go `text/template` files (`.tmpl`) and static files, embedded directly in the binary. When a layer provides a file that already exists (e.g., a framework overrides the base `app.go.tmpl`), the later layer wins.

After all templates are rendered and written to disk, goforge runs `go mod tidy` automatically. If that step fails, the output directory is removed to prevent leaving a broken project behind.

---

## goforge Project Structure

For contributors and maintainers:

```
goforge/
├── main.go                            # Root CLI entry point
├── create.go                          # `create` subcommand — flags, routing, spinner
├── version.go                         # `version` subcommand — ldflags-injected build info
├── internal/
│   ├── config/
│   │   └── config.go                  # Config struct (Name, Module, Framework, DB, Logger, Extras)
│   ├── generator/
│   │   ├── generator.go               # Core rendering engine (layered template composition)
│   │   ├── templates/                 # Embedded template tree (base, frameworks, db, loggers, extras)
│   │   ├── generator_test.go
│   │   ├── frameworks_test.go
│   │   ├── db_test.go
│   │   ├── loggers_test.go
│   │   └── extras_test.go
│   ├── validator/
│   │   ├── validator.go               # Single source of truth for valid stack combinations
│   │   └── validator_test.go
│   ├── prompts/
│   │   └── prompt.go                  # huh-based interactive TUI (3-step wizard)
│   ├── postgen/
│   │   ├── postgen.go                 # Post-generation: go mod tidy, git init
│   │   └── postgen_test.go
│   └── ui/
│       └── spinner.go                 # Animated braille spinner
├── .github/
│   ├── workflows/ci.yml               # CI for goforge itself
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.md
│   │   └── feature_request.md
│   └── pull_request_template.md
├── CONTRIBUTING.md
└── LICENSE
```

---

## Tech Stack

goforge itself is built with:

| Dependency | Purpose |
|------------|---------|
| [Cobra](https://github.com/spf13/cobra) | CLI command and flag parsing |
| [huh](https://github.com/charmbracelet/huh) | Interactive TUI forms |
| [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Terminal styling |
| Go `text/template` + `embed` | Template rendering with embedded file tree |

---

## Roadmap

- [ ] Echo and Fiber framework support
- [ ] MySQL database support
- [ ] Binary releases via GoReleaser (Homebrew tap, GitHub Releases) — `version`, `commit`, and `date` ldflags are already wired
- [ ] `goforge add` command for adding extras to an existing project

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for step-by-step guides on adding a new framework, database combo, logger, or extra.

To run the test suite locally:

```bash
# Fast (skips compilation tests)
go test ./... -short

# Full (compiles generated projects — requires network for go mod tidy)
go test ./...

# Vet
go vet ./...
```

**Good first contributions:**

- Add a new logger implementation (add a layer under `internal/generator/templates/loggers/`)
- Add Echo or Fiber support (templates exist — enable in `internal/validator/validator.go`)
- Improve existing templates or add new example handlers/usecases
- Expand test coverage (see open issues for gaps)

Please open an issue before starting significant work so we can align on approach.

---

## License

[MIT](LICENSE)
