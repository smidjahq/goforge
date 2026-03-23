package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/smidjahq/goforge/internal/config"
)

const dryRun = "--dry-run"

func TestCreateCmd_ValidFlags(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantOut  []string
		wantCode int
	}{
		{
			name: "gin postgres-gorm zap",
			args: []string{"create", dryRun,
				"--name", "myapp",
				"--module", "github.com/you/myapp",
				"--framework", "gin",
				"--db", "postgres-gorm",
				"--logger", "zap",
			},
			wantOut: []string{"myapp", "github.com/you/myapp", "gin", "postgres-gorm", "zap"},
		},
		{
			name: "chi sqlite-raw slog with extras",
			args: []string{"create", dryRun,
				"--name", "svc",
				"--module", "github.com/org/svc",
				"--framework", "chi",
				"--db", "sqlite-raw",
				"--logger", "slog",
				"--extras", "docker,makefile,ci",
			},
			wantOut: []string{"svc", "chi", "sqlite-raw", "slog", "docker, makefile, ci"},
		},
		{
			name: "none DB no extras",
			args: []string{"create", dryRun,
				"--name", "api",
				"--module", "github.com/org/api",
				"--framework", "gin",
				"--db", "none",
				"--logger", "zerolog",
			},
			wantOut: []string{"none", "zerolog"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			root := newRootCmd()
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs(tc.args)

			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			out := buf.String()
			for _, want := range tc.wantOut {
				if !strings.Contains(out, want) {
					t.Errorf("output %q missing %q", out, want)
				}
			}
		})
	}
}

func TestCreateCmd_InvalidFlags(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name: "missing name",
			args: []string{"create",
				"--module", "github.com/you/myapp",
				"--framework", "gin",
				"--db", "postgres-gorm",
				"--logger", "slog",
			},
			wantErr: "project name is required",
		},
		{
			name: "invalid db combo sqlite-sqlc",
			args: []string{"create",
				"--name", "myapp",
				"--module", "github.com/you/myapp",
				"--framework", "gin",
				"--db", "sqlite-sqlc",
				"--logger", "slog",
			},
			wantErr: "sqlc is not supported with sqlite",
		},
		{
			name: "unknown framework",
			args: []string{"create",
				"--name", "myapp",
				"--module", "github.com/you/myapp",
				"--framework", "express",
				"--db", "postgres-gorm",
				"--logger", "slog",
			},
			wantErr: `unknown framework "express"`,
		},
		{
			name: "migrations with none DB",
			args: []string{"create",
				"--name", "myapp",
				"--module", "github.com/you/myapp",
				"--framework", "gin",
				"--db", "none",
				"--logger", "slog",
				"--extras", "migrations",
			},
			wantErr: "migrations extra requires a database",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}

			root := newRootCmd()
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs(tc.args)

			err := root.Execute()
			if err == nil {
				t.Fatal("expected non-nil error")
			}

			combined := buf.String() + err.Error()
			if !strings.Contains(combined, tc.wantErr) {
				t.Errorf("expected error containing %q, got output=%q err=%v", tc.wantErr, buf.String(), err)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseExtras
// ---------------------------------------------------------------------------

func TestParseExtras_EmptyString(t *testing.T) {
	got := parseExtras("")
	if got != nil {
		t.Errorf("expected nil for empty input, got %v", got)
	}
}

func TestParseExtras_WhitespaceOnly(t *testing.T) {
	got := parseExtras("   ")
	if got != nil {
		t.Errorf("expected nil for whitespace-only input, got %v", got)
	}
}

func TestParseExtras_Single(t *testing.T) {
	got := parseExtras("docker")
	if len(got) != 1 || got[0] != "docker" {
		t.Errorf("expected [docker], got %v", got)
	}
}

func TestParseExtras_Multiple(t *testing.T) {
	got := parseExtras("docker,makefile,ci")
	want := []string{"docker", "makefile", "ci"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("index %d: expected %q, got %q", i, w, got[i])
		}
	}
}

func TestParseExtras_TrimsSpaces(t *testing.T) {
	got := parseExtras(" docker , makefile ")
	want := []string{"docker", "makefile"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("index %d: expected %q, got %q", i, w, got[i])
		}
	}
}

func TestParseExtras_SkipsEmptySegments(t *testing.T) {
	// double-comma produces an empty segment that should be dropped
	got := parseExtras("docker,,makefile")
	want := []string{"docker", "makefile"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

// ---------------------------------------------------------------------------
// printNextSteps
// ---------------------------------------------------------------------------

func TestPrintNextSteps_GoRunWhenNoMakefile(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)

	cfg := minimalCfg()
	printNextSteps(root, cfg)

	if !strings.Contains(buf.String(), "go run ./cmd/app") {
		t.Errorf("expected go run fallback, got: %q", buf.String())
	}
}

func TestPrintNextSteps_MakeRunWhenMakefile(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)

	cfg := minimalCfg()
	cfg.Extras = []string{"makefile"}
	printNextSteps(root, cfg)

	if !strings.Contains(buf.String(), "make run") {
		t.Errorf("expected make run, got: %q", buf.String())
	}
}

func TestPrintNextSteps_MakeComposeUpWhenMakefileAndDocker(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)

	cfg := minimalCfg()
	cfg.Extras = []string{"makefile", "docker"}
	printNextSteps(root, cfg)

	out := buf.String()
	if !strings.Contains(out, "make run") {
		t.Errorf("expected make run, got: %q", out)
	}
	if !strings.Contains(out, "make compose-up") {
		t.Errorf("expected make compose-up, got: %q", out)
	}
}

func TestPrintNextSteps_NoComposeUpWithoutDocker(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)

	cfg := minimalCfg()
	cfg.Extras = []string{"makefile"}
	printNextSteps(root, cfg)

	if strings.Contains(buf.String(), "compose-up") {
		t.Errorf("compose-up should not appear without docker extra, got: %q", buf.String())
	}
}

func TestPrintNextSteps_ContainsCdProjectName(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)

	cfg := minimalCfg()
	printNextSteps(root, cfg)

	if !strings.Contains(buf.String(), "cd "+cfg.Name) {
		t.Errorf("expected cd instruction, got: %q", buf.String())
	}
}

// ---------------------------------------------------------------------------
// printSummary
// ---------------------------------------------------------------------------

func TestPrintSummary_AllFieldsPresent(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)

	cfg := minimalCfg()
	cfg.Extras = []string{"docker", "ci"}
	printSummary(root, cfg)

	out := buf.String()
	for _, want := range []string{cfg.Name, cfg.Module, cfg.Framework, cfg.DB, cfg.Logger, "docker, ci"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q; output: %q", want, out)
		}
	}
}

func TestPrintSummary_NoExtrasLabel(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)

	cfg := minimalCfg() // Extras is nil
	printSummary(root, cfg)

	if !strings.Contains(buf.String(), "(none)") {
		t.Errorf("expected (none) for empty extras, got: %q", buf.String())
	}
}

// minimalCfg returns a valid Config for use in output-only tests.
func minimalCfg() config.Config {
	return config.Config{
		Name:      "testapp",
		Module:    "github.com/you/testapp",
		Framework: "gin",
		DB:        "postgres-gorm",
		Logger:    "slog",
	}
}
