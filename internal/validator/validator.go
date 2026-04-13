// Package validator is the single source of truth for all valid stack
// combinations. It exposes no I/O and has no UI dependencies.
package validator

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/smidjahq/goforge/internal/config"
)

var goVersionRe = regexp.MustCompile(`^\d+\.\d+$`)

// validCombinations maps a DB backend to the ORM/access layers that are
// supported with it. "none" has no layers.
var validCombinations = map[string][]string{
	"postgres": {"gorm", "sqlc", "raw"},
	"sqlite":   {"gorm", "raw"},
	"mysql":    {"gorm", "sqlc", "raw"},
	"none":     {},
}

// validFrameworks lists frameworks that are currently selectable.
// Echo and Fiber exist in the template tree but are not yet enabled.
var validFrameworks = []string{"gin", "chi", "echo", "fiber"}

// validLoggers lists supported logger implementations.
var validLoggers = []string{"slog", "zap", "zerolog"}

// validExtras lists all recognised extra options.
var validExtras = []string{"docker", "makefile", "ci", "swagger", "migrations", "linter"}

// Frameworks returns the list of currently enabled HTTP frameworks.
func Frameworks() []string {
	out := make([]string, len(validFrameworks))
	copy(out, validFrameworks)
	return out
}

// Loggers returns the list of supported logger implementations.
func Loggers() []string {
	out := make([]string, len(validLoggers))
	copy(out, validLoggers)
	return out
}

// Extras returns the list of recognised extra options.
func Extras() []string {
	out := make([]string, len(validExtras))
	copy(out, validExtras)
	return out
}

// DBCombinations returns all valid DB values accepted by the --db flag,
// including "none" and every "backend-layer" pair.
func DBCombinations() []string {
	out := []string{}
	// Stable ordering: postgres, sqlite, mysql, none
	order := []string{"postgres", "sqlite", "mysql", "none"}
	for _, backend := range order {
		layers, ok := validCombinations[backend]
		if !ok {
			continue
		}
		if len(layers) == 0 {
			out = append(out, backend)
			continue
		}
		for _, layer := range layers {
			out = append(out, backend+"-"+layer)
		}
	}
	return out
}

// ValidOptionsFor returns the valid ORM/access-layer options for the given DB
// backend (e.g. "postgres", "sqlite", "none"). Returns nil for unknown backends.
func ValidOptionsFor(db string) []string {
	layers, ok := validCombinations[db]
	if !ok {
		return nil
	}
	// return a copy so callers cannot mutate the package-level slice
	out := make([]string, len(layers))
	copy(out, layers)
	return out
}

// Validate checks that cfg describes a valid, self-consistent stack.
// Returns nil when the config is valid, or a descriptive error otherwise.
func Validate(cfg config.Config) error {
	if strings.TrimSpace(cfg.Name) == "" {
		return fmt.Errorf("project name is required")
	}
	if strings.TrimSpace(cfg.Module) == "" {
		return fmt.Errorf("module path is required")
	}

	if err := validateFramework(cfg.Framework); err != nil {
		return err
	}
	if err := validateDB(cfg.DB); err != nil {
		return err
	}
	if err := validateLogger(cfg.Logger); err != nil {
		return err
	}
	if err := validateExtras(cfg.Extras, cfg.DB); err != nil {
		return err
	}
	if err := validateGoVersion(cfg.GoVersion); err != nil {
		return err
	}
	return nil
}

func validateGoVersion(v string) error {
	if v == "" {
		return nil // empty means "use toolchain default"; CLI/TUI always populate this
	}
	if !goVersionRe.MatchString(v) {
		return fmt.Errorf("invalid go version %q; expected format: major.minor (e.g. 1.26)", v)
	}
	return nil
}

func validateFramework(framework string) error {
	if framework == "" {
		return fmt.Errorf("framework is required; valid options: %s", strings.Join(validFrameworks, ", "))
	}
	if !slices.Contains(validFrameworks, framework) {
		return fmt.Errorf(
			"unknown framework %q; valid options: %s",
			framework, strings.Join(validFrameworks, ", "),
		)
	}
	return nil
}

// validateDB checks that the combined DB value (e.g. "postgres-gorm") is one
// of the valid combinations, or "none".
func validateDB(db string) error {
	if db == "none" {
		return nil
	}

	backend, layer, ok := strings.Cut(db, "-")
	if !ok {
		return fmt.Errorf(
			"invalid --db value %q; expected format: <backend>-<layer> or \"none\"",
			db,
		)
	}

	layers, known := validCombinations[backend]
	if !known {
		backends := make([]string, 0, len(validCombinations))
		for k := range validCombinations {
			backends = append(backends, k)
		}
		return fmt.Errorf(
			"unknown database backend %q; valid backends: %s",
			backend, strings.Join(backends, ", "),
		)
	}

	if !slices.Contains(layers, layer) {
		return fmt.Errorf(
			"%s is not supported with %s; valid options: %s",
			layer, backend, strings.Join(layers, ", "),
		)
	}
	return nil
}

func validateLogger(logger string) error {
	if !slices.Contains(validLoggers, logger) {
		return fmt.Errorf(
			"unknown logger %q; valid options: %s",
			logger, strings.Join(validLoggers, ", "),
		)
	}
	return nil
}

func validateExtras(extras []string, db string) error {
	for _, extra := range extras {
		if !slices.Contains(validExtras, extra) {
			return fmt.Errorf(
				"unknown extra %q; valid options: %s",
				extra, strings.Join(validExtras, ", "),
			)
		}
		if extra == "migrations" && db == "none" {
			return fmt.Errorf(
				"the migrations extra requires a database; set --db to a non-none value or remove migrations from --extras",
			)
		}
	}
	return nil
}
