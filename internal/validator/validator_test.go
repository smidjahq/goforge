package validator_test

import (
	"testing"

	"github.com/smidjahq/goforge/internal/config"
	"github.com/smidjahq/goforge/internal/validator"
)

// base returns a fully-valid Config that individual test cases can override.
func base() config.Config {
	return config.Config{
		Name:      "myapp",
		Module:    "github.com/you/myapp",
		Framework: "gin",
		DB:        "postgres-gorm",
		Logger:    "slog",
		Extras:    nil,
	}
}

func TestValidate_ValidCombinations(t *testing.T) {
	cases := []struct {
		name string
		cfg  config.Config
	}{
		// All valid framework × logger combinations with postgres-gorm
		{"gin+postgres-gorm+slog", base()},
		{"gin+postgres-gorm+zap", func() config.Config { c := base(); c.Logger = "zap"; return c }()},
		{"gin+postgres-gorm+zerolog", func() config.Config { c := base(); c.Logger = "zerolog"; return c }()},
		{"chi+postgres-gorm+slog", func() config.Config { c := base(); c.Framework = "chi"; return c }()},
		{"echo+postgres-gorm+slog", func() config.Config { c := base(); c.Framework = "echo"; return c }()},
		{"fiber+postgres-gorm+slog", func() config.Config { c := base(); c.Framework = "fiber"; return c }()},

		// All valid postgres DB combos
		{"postgres-gorm", base()},
		{"postgres-sqlc", func() config.Config { c := base(); c.DB = "postgres-sqlc"; return c }()},
		{"postgres-raw", func() config.Config { c := base(); c.DB = "postgres-raw"; return c }()},

		// All valid sqlite DB combos
		{"sqlite-gorm", func() config.Config { c := base(); c.DB = "sqlite-gorm"; return c }()},
		{"sqlite-raw", func() config.Config { c := base(); c.DB = "sqlite-raw"; return c }()},

		// All valid MySQL DB combos
		{"mysql-gorm", func() config.Config { c := base(); c.DB = "mysql-gorm"; return c }()},
		{"mysql-sqlc", func() config.Config { c := base(); c.DB = "mysql-sqlc"; return c }()},
		{"mysql-raw", func() config.Config { c := base(); c.DB = "mysql-raw"; return c }()},

		// none DB
		{"none", func() config.Config { c := base(); c.DB = "none"; return c }()},

		// Extras — all valid, with a non-none DB
		{"extras all", func() config.Config {
			c := base()
			c.Extras = []string{"docker", "makefile", "ci", "swagger", "migrations", "linter"}
			return c
		}()},
		{"extras without migrations", func() config.Config {
			c := base()
			c.Extras = []string{"docker", "makefile", "ci", "linter"}
			return c
		}()},
		// migrations with sqlite is fine
		{"sqlite-gorm+migrations", func() config.Config {
			c := base()
			c.DB = "sqlite-gorm"
			c.Extras = []string{"migrations"}
			return c
		}()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := validator.Validate(tc.cfg); err != nil {
				t.Errorf("expected nil error, got: %v", err)
			}
		})
	}
}

func TestValidate_InvalidCombinations(t *testing.T) {
	cases := []struct {
		name        string
		cfg         config.Config
		errContains string
	}{
		// Missing required fields
		{
			name:        "empty name",
			cfg:         func() config.Config { c := base(); c.Name = ""; return c }(),
			errContains: "project name is required",
		},
		{
			name:        "whitespace name",
			cfg:         func() config.Config { c := base(); c.Name = "   "; return c }(),
			errContains: "project name is required",
		},
		{
			name:        "empty module",
			cfg:         func() config.Config { c := base(); c.Module = ""; return c }(),
			errContains: "module path is required",
		},

		// Invalid framework
		{
			name:        "empty framework",
			cfg:         func() config.Config { c := base(); c.Framework = ""; return c }(),
			errContains: "framework is required",
		},
		{
			name:        "unknown framework",
			cfg:         func() config.Config { c := base(); c.Framework = "express"; return c }(),
			errContains: `unknown framework "express"`,
		},
		// Invalid DB combos
		{
			name:        "sqlite-sqlc unsupported",
			cfg:         func() config.Config { c := base(); c.DB = "sqlite-sqlc"; return c }(),
			errContains: "sqlc is not supported with sqlite",
		},
		{
			name:        "unknown backend",
			cfg:         func() config.Config { c := base(); c.DB = "oracle-gorm"; return c }(),
			errContains: `unknown database backend "oracle"`,
		},
		{
			name:        "malformed db no dash",
			cfg:         func() config.Config { c := base(); c.DB = "postgres"; return c }(),
			errContains: "invalid --db value",
		},
		{
			name:        "unknown layer",
			cfg:         func() config.Config { c := base(); c.DB = "postgres-prisma"; return c }(),
			errContains: "prisma is not supported with postgres",
		},

		// Invalid logger
		{
			name:        "unknown logger",
			cfg:         func() config.Config { c := base(); c.Logger = "logrus"; return c }(),
			errContains: `unknown logger "logrus"`,
		},

		// Invalid extras
		{
			name:        "unknown extra",
			cfg:         func() config.Config { c := base(); c.Extras = []string{"prometheus"}; return c }(),
			errContains: `unknown extra "prometheus"`,
		},
		{
			name: "migrations with none DB",
			cfg: func() config.Config {
				c := base()
				c.DB = "none"
				c.Extras = []string{"migrations"}
				return c
			}(),
			errContains: "migrations extra requires a database",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.Validate(tc.cfg)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.errContains)
			}
			if !containsCI(err.Error(), tc.errContains) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.errContains)
			}
		})
	}
}

func TestValidOptionsFor(t *testing.T) {
	cases := []struct {
		db      string
		want    []string
		wantNil bool
	}{
		{"postgres", []string{"gorm", "sqlc", "raw"}, false},
		{"sqlite", []string{"gorm", "raw"}, false},
		{"none", []string{}, false},
		{"mysql", []string{"gorm", "sqlc", "raw"}, false},
		{"", nil, true},
	}

	for _, tc := range cases {
		t.Run(tc.db, func(t *testing.T) {
			got := validator.ValidOptionsFor(tc.db)
			if tc.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				return
			}
			if len(got) != len(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("index %d: expected %q, got %q", i, tc.want[i], got[i])
				}
			}
		})
	}
}

// TestValidOptionsFor_ReturnsCopy ensures callers cannot mutate internal state.
func TestValidOptionsFor_ReturnsCopy(t *testing.T) {
	first := validator.ValidOptionsFor("postgres")
	if len(first) == 0 {
		t.Fatal("expected non-empty slice")
	}
	first[0] = "MUTATED"

	second := validator.ValidOptionsFor("postgres")
	if second[0] == "MUTATED" {
		t.Error("ValidOptionsFor returned a reference to internal slice; mutation affected future calls")
	}
}

// containsCI is a case-insensitive substring check.
func containsCI(s, substr string) bool {
	return len(substr) == 0 ||
		len(s) >= len(substr) &&
			(s == substr ||
				stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
