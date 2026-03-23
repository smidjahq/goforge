package postgen_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/smidjahq/goforge/internal/config"
	"github.com/smidjahq/goforge/internal/generator"
	"github.com/smidjahq/goforge/internal/postgen"
)

func generateProject(t *testing.T) string {
	t.Helper()
	out := t.TempDir()
	cfg := config.Config{
		Name:      "testapp",
		Module:    "github.com/test/testapp",
		Framework: "gin",
		DB:        "none",
		Logger:    "slog",
	}
	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return out
}

func TestRun_GoModTidy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgen test in short mode (requires network)")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go binary not in PATH")
	}

	out := generateProject(t)

	if err := postgen.Run(out, false); err != nil {
		t.Fatalf("postgen.Run: %v", err)
	}

	// go.sum should now exist after tidy.
	if _, err := os.Stat(filepath.Join(out, "go.sum")); os.IsNotExist(err) {
		t.Error("expected go.sum to exist after go mod tidy")
	}
}

func TestRun_GitInit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgen test in short mode (requires network)")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go binary not in PATH")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git binary not in PATH")
	}

	out := generateProject(t)

	if err := postgen.Run(out, true); err != nil {
		t.Fatalf("postgen.Run with git=true: %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, ".git")); os.IsNotExist(err) {
		t.Error("expected .git directory to exist after git init")
	}
}

func TestRun_FailedTidyCleansUp(t *testing.T) {
	// Point postgen at a directory with a broken go.mod so tidy fails.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("this is not valid go.mod syntax\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := postgen.Run(dir, false)
	if err == nil {
		t.Fatal("expected error from broken go.mod, got nil")
	}

	// The directory should have been removed.
	if _, statErr := os.Stat(dir); !os.IsNotExist(statErr) {
		t.Error("expected output directory to be removed after go mod tidy failure")
	}
}
