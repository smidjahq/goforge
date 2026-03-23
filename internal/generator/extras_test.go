package generator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/smidjahq/goforge/internal/config"
	"github.com/smidjahq/goforge/internal/generator"
)

func extrasConfig(db string, extras ...string) config.Config {
	return config.Config{
		Name:      "testapp",
		Module:    "github.com/test/testapp",
		Framework: "gin",
		DB:        db,
		Logger:    "slog",
		Extras:    extras,
	}
}

func TestGenerate_DockerExtra(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(extrasConfig("postgres-gorm", "docker"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	mustExist := []string{"Dockerfile", "docker-compose.override.yml"}
	for _, f := range mustExist {
		if _, err := os.Stat(filepath.Join(out, f)); os.IsNotExist(err) {
			t.Errorf("missing %s", f)
		}
	}

	data, _ := os.ReadFile(filepath.Join(out, "Dockerfile"))
	if !strings.Contains(string(data), "testapp") {
		t.Error("Dockerfile missing project name")
	}
}

func TestGenerate_DockerOverrideHasDependsOnWhenDBSet(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(extrasConfig("postgres-gorm", "docker"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "docker-compose.override.yml"))
	if !strings.Contains(string(data), "depends_on") {
		t.Error("docker-compose.override.yml missing depends_on for postgres DB")
	}
}

func TestGenerate_DockerOverrideNoDependsOnWhenNoneDB(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(extrasConfig("none", "docker"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "docker-compose.override.yml"))
	if strings.Contains(string(data), "depends_on") {
		t.Error("docker-compose.override.yml should not have depends_on when DB is none")
	}
}

func TestGenerate_MakefileExtra(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(extrasConfig("none", "makefile"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "Makefile")); os.IsNotExist(err) {
		t.Error("missing Makefile")
	}
	data, _ := os.ReadFile(filepath.Join(out, "Makefile"))
	for _, target := range []string{"run", "test", "lint", "compose-up"} {
		if !strings.Contains(string(data), target+":") {
			t.Errorf("Makefile missing target %q", target)
		}
	}
}

func TestGenerate_CIExtra(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(extrasConfig("none", "ci"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	p := filepath.Join(out, ".github/workflows/ci.yml")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Error("missing .github/workflows/ci.yml")
	}
}

func TestGenerate_SwaggerExtra(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(extrasConfig("none", "swagger"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "docs")); os.IsNotExist(err) {
		t.Error("missing docs/ directory")
	}
}

func TestGenerate_MigrationsExtra(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(extrasConfig("postgres-gorm", "migrations"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "migrations")); os.IsNotExist(err) {
		t.Error("missing migrations/ directory")
	}
}

func TestGenerate_MigrationsExtraSkippedWithNoneDB(t *testing.T) {
	out := t.TempDir()
	// Generator should silently skip migrations when DB is none.
	if err := generator.Generate(extrasConfig("none", "migrations"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "migrations")); !os.IsNotExist(err) {
		t.Error("migrations/ should be absent when DB is none")
	}
}

func TestGenerate_LinterExtra(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(extrasConfig("none", "linter"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, ".golangci.yml")); os.IsNotExist(err) {
		t.Error("missing .golangci.yml")
	}
}

func TestGenerate_AllExtras(t *testing.T) {
	out := t.TempDir()
	cfg := config.Config{
		Name:      "testapp",
		Module:    "github.com/test/testapp",
		Framework: "gin",
		DB:        "postgres-gorm",
		Logger:    "slog",
		Extras:    []string{"docker", "makefile", "ci", "swagger", "migrations", "linter"},
	}
	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	mustExist := []string{
		"Dockerfile",
		"docker-compose.override.yml",
		"Makefile",
		".github/workflows/ci.yml",
		"docs",
		"migrations",
		".golangci.yml",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(out, rel)); os.IsNotExist(err) {
			t.Errorf("missing %s", rel)
		}
	}
}
