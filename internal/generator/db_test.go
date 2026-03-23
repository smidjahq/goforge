package generator_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/smidjahq/goforge/internal/config"
	"github.com/smidjahq/goforge/internal/generator"
)

// dbConfig returns a gin+slog config with the given DB combo.
func dbConfig(db string) config.Config {
	return config.Config{
		Name:      "testapp",
		Module:    "github.com/test/testapp",
		Framework: "gin",
		DB:        db,
		Logger:    "slog",
	}
}

func TestGenerate_PostgresGorm_FileTree(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(dbConfig("postgres-gorm"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	mustExist := []string{
		"internal/repo/persistent/repo.go",
		"docker-compose.yml",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(out, rel)); os.IsNotExist(err) {
			t.Errorf("missing file: %s", rel)
		}
	}
}

func TestGenerate_PostgresSqlc_FileTree(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(dbConfig("postgres-sqlc"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	mustExist := []string{
		"internal/repo/persistent/repo.go",
		"internal/repo/persistent/queries/.gitkeep",
		"sqlc.yaml",
		"docker-compose.yml",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(out, rel)); os.IsNotExist(err) {
			t.Errorf("missing file: %s", rel)
		}
	}
}

func TestGenerate_PostgresRaw_FileTree(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(dbConfig("postgres-raw"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	mustExist := []string{
		"internal/repo/persistent/repo.go",
		"docker-compose.yml",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(out, rel)); os.IsNotExist(err) {
			t.Errorf("missing file: %s", rel)
		}
	}
}

func TestGenerate_SqliteGorm_FileTree(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(dbConfig("sqlite-gorm"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "internal/repo/persistent/repo.go")); os.IsNotExist(err) {
		t.Error("missing internal/repo/persistent/repo.go")
	}
	// SQLite has no docker-compose
	if _, err := os.Stat(filepath.Join(out, "docker-compose.yml")); !os.IsNotExist(err) {
		t.Error("sqlite-gorm should not have docker-compose.yml")
	}
}

func TestGenerate_SqliteRaw_FileTree(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(dbConfig("sqlite-raw"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "internal/repo/persistent/repo.go")); os.IsNotExist(err) {
		t.Error("missing internal/repo/persistent/repo.go")
	}
	if _, err := os.Stat(filepath.Join(out, "docker-compose.yml")); !os.IsNotExist(err) {
		t.Error("sqlite-raw should not have docker-compose.yml")
	}
}

func TestGenerate_NoneDB_NoRepoDirectory(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(dbConfig("none"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "internal/repo/persistent")); !os.IsNotExist(err) {
		t.Error("DB=none should not produce internal/repo/persistent/")
	}
	if _, err := os.Stat(filepath.Join(out, "docker-compose.yml")); !os.IsNotExist(err) {
		t.Error("DB=none should not produce docker-compose.yml")
	}
}

func TestGenerate_PostgresGoModContainsDeps(t *testing.T) {
	cases := []struct {
		db      string
		wantDep string
	}{
		{"postgres-gorm", "gorm.io/gorm"},
		{"postgres-gorm", "gorm.io/driver/postgres"},
		{"postgres-sqlc", "github.com/jackc/pgx/v5"},
		{"postgres-raw", "github.com/jackc/pgx/v5"},
		{"sqlite-gorm", "gorm.io/driver/sqlite"},
		{"sqlite-raw", "github.com/mattn/go-sqlite3"},
	}
	for _, tc := range cases {
		t.Run(tc.db+"/"+tc.wantDep, func(t *testing.T) {
			out := t.TempDir()
			if err := generator.Generate(dbConfig(tc.db), out, nil); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			data, _ := os.ReadFile(filepath.Join(out, "go.mod"))
			if !strings.Contains(string(data), tc.wantDep) {
				t.Errorf("go.mod missing %q:\n%s", tc.wantDep, data)
			}
		})
	}
}

func TestGenerate_DockerComposeContainsProjectName(t *testing.T) {
	for _, db := range []string{"postgres-gorm", "postgres-sqlc", "postgres-raw"} {
		t.Run(db, func(t *testing.T) {
			out := t.TempDir()
			cfg := dbConfig(db)
			cfg.Name = "myservice"
			if err := generator.Generate(cfg, out, nil); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			data, _ := os.ReadFile(filepath.Join(out, "docker-compose.yml"))
			if !strings.Contains(string(data), "myservice") {
				t.Errorf("docker-compose.yml missing project name:\n%s", data)
			}
		})
	}
}

func TestGenerate_SqlcYamlValid(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(dbConfig("postgres-sqlc"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(out, "sqlc.yaml"))
	if err != nil {
		t.Fatalf("read sqlc.yaml: %v", err)
	}
	content := string(data)
	for _, want := range []string{"version", "postgresql", "pgx/v5"} {
		if !strings.Contains(content, want) {
			t.Errorf("sqlc.yaml missing %q:\n%s", want, content)
		}
	}
}

// TestGenerate_DBBuilds verifies generated projects compile (requires network for go mod tidy).
func TestGenerate_DBBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compilation test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go binary not in PATH")
	}

	// sqlite variants require CGO; skip if CGO_ENABLED=0
	isCGOEnabled := os.Getenv("CGO_ENABLED") != "0"

	cases := []struct {
		db       string
		needsCGO bool
	}{
		{"postgres-gorm", false},
		{"postgres-sqlc", false},
		{"postgres-raw", false},
		{"sqlite-gorm", true},
		{"sqlite-raw", true},
		{"none", false},
	}

	for _, tc := range cases {
		t.Run(tc.db, func(t *testing.T) {
			if tc.needsCGO && !isCGOEnabled {
				t.Skip("CGO disabled; skipping sqlite build test")
			}

			out := t.TempDir()
			if err := generator.Generate(dbConfig(tc.db), out, nil); err != nil {
				t.Fatalf("Generate: %v", err)
			}

			tidy := exec.Command("go", "mod", "tidy")
			tidy.Dir = out
			if output, err := tidy.CombinedOutput(); err != nil {
				t.Fatalf("go mod tidy: %s", output)
			}

			build := exec.Command("go", "build", "./...")
			build.Dir = out
			if output, err := build.CombinedOutput(); err != nil {
				t.Fatalf("go build ./...: %s", output)
			}
		})
	}
}
