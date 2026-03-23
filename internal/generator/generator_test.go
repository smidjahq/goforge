package generator_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/smidjahq/goforge/internal/config"
	"github.com/smidjahq/goforge/internal/generator"
)

func baseConfig() config.Config {
	return config.Config{
		Name:      "testapp",
		Module:    "github.com/test/testapp",
		Framework: "gin",
		DB:        "postgres-gorm",
		Logger:    "slog",
	}
}

func TestGenerate_ProducesBaseFileTree(t *testing.T) {
	out := t.TempDir()
	cfg := baseConfig()

	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	mustExist := []string{
		"go.mod",
		"cmd/app/main.go",
		"config/config.go",
		"internal/app/app.go",
		"internal/entity/.gitkeep",
		"internal/usecase/.gitkeep",
		"pkg/constants/.gitkeep",
		"pkg/utils/.gitkeep",
		"pkg/client/http_client.go",
		"pkg/logger/logger.go",
		".env.example",
		".gitignore",
	}

	for _, rel := range mustExist {
		p := filepath.Join(out, rel)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected file %q to exist", rel)
		}
	}
}

func TestGenerate_GoModContainsModulePath(t *testing.T) {
	out := t.TempDir()
	cfg := baseConfig()
	cfg.Module = "github.com/acme/myservice"

	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(out, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}

	content := string(data)
	if !contains(content, "module github.com/acme/myservice") {
		t.Errorf("go.mod missing module declaration, got:\n%s", content)
	}
}

func TestGenerate_MainGoImportsCorrectModule(t *testing.T) {
	out := t.TempDir()
	cfg := baseConfig()
	cfg.Module = "github.com/org/svc"

	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(out, "cmd/app/main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}

	if !contains(string(data), `"github.com/org/svc/internal/app"`) {
		t.Errorf("main.go missing correct import, got:\n%s", string(data))
	}
}

func TestGenerate_NoTmplFilesInOutput(t *testing.T) {
	out := t.TempDir()

	if err := generator.Generate(baseConfig(), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	err := filepath.Walk(out, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(p) == ".tmpl" {
			t.Errorf("unexpected .tmpl file in output: %s", p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk output: %v", err)
	}
}

func TestGenerate_OutputBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compilation test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go binary not in PATH")
	}

	out := t.TempDir()

	if err := generator.Generate(baseConfig(), out, nil); err != nil {
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
		t.Fatalf("go build ./... failed:\n%s", output)
	}
}

func TestGenerate_OutputDirectoryCreated(t *testing.T) {
	base := t.TempDir()
	out := filepath.Join(base, "nested", "newproject")

	if err := generator.Generate(baseConfig(), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if _, err := os.Stat(out); os.IsNotExist(err) {
		t.Error("output directory was not created")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
