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

func loggerConfig(loggerName string) config.Config {
	return config.Config{
		Name:      "testapp",
		Module:    "github.com/test/testapp",
		Framework: "gin",
		DB:        "none",
		Logger:    loggerName,
	}
}

func TestGenerate_LoggerImplFileExists(t *testing.T) {
	for _, lg := range []string{"slog", "zap", "zerolog"} {
		t.Run(lg, func(t *testing.T) {
			out := t.TempDir()
			if err := generator.Generate(loggerConfig(lg), out, nil); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			p := filepath.Join(out, "pkg/logger/impl.go")
			if _, err := os.Stat(p); os.IsNotExist(err) {
				t.Errorf("missing pkg/logger/impl.go for logger=%s", lg)
			}
		})
	}
}

func TestGenerate_LoggerGoModContainsDep(t *testing.T) {
	cases := []struct {
		logger  string
		wantDep string
	}{
		{"zap", "go.uber.org/zap"},
		{"zerolog", "github.com/rs/zerolog"},
	}
	for _, tc := range cases {
		t.Run(tc.logger, func(t *testing.T) {
			out := t.TempDir()
			if err := generator.Generate(loggerConfig(tc.logger), out, nil); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			data, _ := os.ReadFile(filepath.Join(out, "go.mod"))
			if !strings.Contains(string(data), tc.wantDep) {
				t.Errorf("go.mod missing %q:\n%s", tc.wantDep, data)
			}
		})
	}
}

func TestGenerate_SlogNoExtraDep(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(loggerConfig("slog"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "go.mod"))
	content := string(data)
	for _, unwanted := range []string{"go.uber.org/zap", "github.com/rs/zerolog"} {
		if strings.Contains(content, unwanted) {
			t.Errorf("slog go.mod should not contain %q", unwanted)
		}
	}
}

func TestGenerate_AppGoCallsLoggerNew(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(loggerConfig("slog"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "internal/app/app.go"))
	if !strings.Contains(string(data), "logger.New()") {
		t.Errorf("app.go does not call logger.New():\n%s", data)
	}
}

func TestGenerate_WebAPIExists(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(loggerConfig("slog"), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	p := filepath.Join(out, "internal/repo/webapi/webapi.go")
	if _, err := os.Stat(p); os.IsNotExist(err) {
		t.Error("missing internal/repo/webapi/webapi.go")
	}
	data, _ := os.ReadFile(p)
	if !strings.Contains(string(data), "pkg/client") {
		t.Error("webapi.go should import pkg/client")
	}
}

func TestGenerate_LoggerBuilds(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compilation test in short mode")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go binary not in PATH")
	}

	for _, lg := range []string{"slog", "zap", "zerolog"} {
		t.Run(lg, func(t *testing.T) {
			out := t.TempDir()
			if err := generator.Generate(loggerConfig(lg), out, nil); err != nil {
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
