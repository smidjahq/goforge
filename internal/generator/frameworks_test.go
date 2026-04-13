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

// frameworkConfig returns a config with the given framework, no DB, no extras.
func frameworkConfig(framework string) config.Config {
	return config.Config{
		Name:      "testapp",
		Module:    "github.com/test/testapp",
		Framework: framework,
		DB:        "none",
		Logger:    "slog",
	}
}

func TestGenerate_EchoFileTree(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(frameworkConfig("echo"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	mustExist := []string{
		"internal/app/app.go",
		"internal/controller/http/router.go",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(out, rel)); os.IsNotExist(err) {
			t.Errorf("missing file: %s", rel)
		}
	}
}

func TestGenerate_FiberFileTree(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(frameworkConfig("fiber"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	mustExist := []string{
		"internal/app/app.go",
		"internal/controller/http/router.go",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(out, rel)); os.IsNotExist(err) {
			t.Errorf("missing file: %s", rel)
		}
	}
}

func TestGenerate_GinFileTree(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(frameworkConfig("gin"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	mustExist := []string{
		"internal/app/app.go",
		"internal/controller/http/router.go",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(out, rel)); os.IsNotExist(err) {
			t.Errorf("missing file: %s", rel)
		}
	}
}

func TestGenerate_ChiFileTree(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(frameworkConfig("chi"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	mustExist := []string{
		"internal/app/app.go",
		"internal/controller/http/router.go",
	}
	for _, rel := range mustExist {
		if _, err := os.Stat(filepath.Join(out, rel)); os.IsNotExist(err) {
			t.Errorf("missing file: %s", rel)
		}
	}
}

func TestGenerate_EchoRouterContainsHealthz(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(frameworkConfig("echo"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(out, "internal/controller/http/router.go"))
	if !strings.Contains(string(data), "/healthz") {
		t.Error("echo router.go missing /healthz route")
	}
	if !strings.Contains(string(data), "labstack/echo") {
		t.Error("echo router.go missing echo import")
	}
}

func TestGenerate_FiberRouterContainsHealthz(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(frameworkConfig("fiber"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(out, "internal/controller/http/router.go"))
	if !strings.Contains(string(data), "/healthz") {
		t.Error("fiber router.go missing /healthz route")
	}
	if !strings.Contains(string(data), "gofiber/fiber") {
		t.Error("fiber router.go missing fiber import")
	}
}

func TestGenerate_GinRouterContainsHealthz(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(frameworkConfig("gin"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(out, "internal/controller/http/router.go"))
	if !strings.Contains(string(data), "/healthz") {
		t.Error("gin router.go missing /healthz route")
	}
	if !strings.Contains(string(data), "gin-gonic/gin") {
		t.Error("gin router.go missing gin import")
	}
}

func TestGenerate_ChiRouterContainsHealthz(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(frameworkConfig("chi"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(out, "internal/controller/http/router.go"))
	if !strings.Contains(string(data), "/healthz") {
		t.Error("chi router.go missing /healthz route")
	}
	if !strings.Contains(string(data), "go-chi/chi") {
		t.Error("chi router.go missing chi import")
	}
}

func TestGenerate_FrameworkGoModContainsDep(t *testing.T) {
	cases := []struct {
		framework string
		wantDep   string
	}{
		{"gin", "github.com/gin-gonic/gin"},
		{"chi", "github.com/go-chi/chi/v5"},
		{"echo", "github.com/labstack/echo/v4"},
		{"fiber", "github.com/gofiber/fiber/v2"},
	}
	for _, tc := range cases {
		t.Run(tc.framework, func(t *testing.T) {
			out := t.TempDir()
			if err := generator.Generate(frameworkConfig(tc.framework), out, nil); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			data, _ := os.ReadFile(filepath.Join(out, "go.mod"))
			if !strings.Contains(string(data), tc.wantDep) {
				t.Errorf("go.mod missing %q:\n%s", tc.wantDep, data)
			}
		})
	}
}

func TestGenerate_FrameworkAppGoImportsController(t *testing.T) {
	for _, fw := range []string{"gin", "chi", "echo", "fiber"} {
		t.Run(fw, func(t *testing.T) {
			out := t.TempDir()
			cfg := frameworkConfig(fw)
			cfg.Module = "github.com/org/svc"
			if err := generator.Generate(cfg, out, nil); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			data, _ := os.ReadFile(filepath.Join(out, "internal/app/app.go"))
			if !strings.Contains(string(data), "github.com/org/svc/internal/controller/http") {
				t.Errorf("app.go missing controller import:\n%s", data)
			}
		})
	}
}

func TestGenerate_FrameworkBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compilation test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go binary not in PATH")
	}

	for _, fw := range []string{"gin", "chi", "echo", "fiber"} {
		t.Run(fw, func(t *testing.T) {
			out := t.TempDir()
			if err := generator.Generate(frameworkConfig(fw), out, nil); err != nil {
				t.Fatalf("Generate: %v", err)
			}

			tidy := exec.Command("go", "mod", "tidy")
			tidy.Dir = out
			if out, err := tidy.CombinedOutput(); err != nil {
				t.Fatalf("go mod tidy: %s", out)
			}

			build := exec.Command("go", "build", "./...")
			build.Dir = out
			if out, err := build.CombinedOutput(); err != nil {
				t.Fatalf("go build ./...: %s", out)
			}
		})
	}
}
