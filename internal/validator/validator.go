// Package validator is the single source of truth for all valid stack
// combinations. It exposes no I/O and has no UI dependencies.
package validator

import (
	"fmt"
	"slices"
	"strings"

	"github.com/smidjahq/goforge/internal/config"
)

// validCombinations maps a DB backend to the ORM/access layers that are
// supported with it. "none" has no layers.
var validCombinations = map[string][]string{
	"postgres": {"gorm", "sqlc", "raw"},
	"sqlite":   {"gorm", "raw"},
	"none":     {},
}

// validFrameworks lists frameworks that are currently selectable.
// Echo and Fiber exist in the template tree but are not yet enabled.
var validFrameworks = []string{"gin", "chi"}

// validLoggers lists supported logger implementations.
var validLoggers = []string{"slog", "zap", "zerolog"}

// validExtras lists all recognised extra options.
var validExtras = []string{"docker", "makefile", "ci", "swagger", "migrations", "linter"}

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
	return nil
}

func validateFramework(framework string) error {
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
