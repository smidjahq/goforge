package generator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/smidjahq/goforge/internal/config"
	"github.com/smidjahq/goforge/internal/generator"
)

// aiFilesConfig returns a representative config for AI files tests.
func aiFilesConfig() config.Config {
	return config.Config{
		Name:      "testapp",
		Module:    "github.com/test/testapp",
		Framework: "gin",
		DB:        "postgres-gorm",
		Logger:    "zap",
		Extras:    []string{"docker", "makefile", "migrations"},
		GoVersion: "1.23",
	}
}

func TestGenerate_AgentsMdExists(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(aiFilesConfig(), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("missing AGENTS.md")
	}
}

func TestGenerate_ReadmeMdExists(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(aiFilesConfig(), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "README.md")); os.IsNotExist(err) {
		t.Error("missing README.md")
	}
}

func TestGenerate_AgentsMd_ContainsStackInfo(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(aiFilesConfig(), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "AGENTS.md"))
	s := string(data)
	for _, want := range []string{"gin", "postgres-gorm", "zap", "testapp"} {
		if !strings.Contains(s, want) {
			t.Errorf("AGENTS.md missing %q", want)
		}
	}
}

func TestGenerate_AgentsMd_ContainsDockerCommandsWhenDockerExtra(t *testing.T) {
	out := t.TempDir()
	cfg := aiFilesConfig()
	cfg.Extras = []string{"docker"}
	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "AGENTS.md"))
	if !strings.Contains(string(data), "docker compose") {
		t.Error("AGENTS.md missing docker compose commands when docker extra selected")
	}
}

func TestGenerate_AgentsMd_NoDockerCommandsWithoutDockerExtra(t *testing.T) {
	out := t.TempDir()
	cfg := aiFilesConfig()
	cfg.Extras = []string{}
	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "AGENTS.md"))
	if strings.Contains(string(data), "docker compose") {
		t.Error("AGENTS.md should not contain docker compose when docker extra not selected")
	}
}

func TestGenerate_AgentsMd_ContainsMigrateCommandsWhenMigrationsExtra(t *testing.T) {
	out := t.TempDir()
	cfg := aiFilesConfig()
	cfg.Extras = []string{"migrations"}
	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "AGENTS.md"))
	if !strings.Contains(string(data), "migrate") {
		t.Error("AGENTS.md missing migrate commands when migrations extra selected")
	}
}

func TestGenerate_AgentsMd_ContainsMakeHelpWhenMakefileExtra(t *testing.T) {
	out := t.TempDir()
	cfg := aiFilesConfig()
	cfg.Extras = []string{"makefile"}
	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "AGENTS.md"))
	if !strings.Contains(string(data), "make help") {
		t.Error("AGENTS.md missing 'make help' note when makefile extra selected")
	}
}

func TestGenerate_AgentsMd_FrameworkRoutingSyntax(t *testing.T) {
	cases := []struct {
		framework string
		wantRoute string
	}{
		{"gin", `r.GET(`},
		{"chi", `r.Get(`},
		{"echo", `e.GET(`},
		{"fiber", `app.Get(`},
	}
	for _, tc := range cases {
		t.Run(tc.framework, func(t *testing.T) {
			out := t.TempDir()
			cfg := config.Config{
				Name:      "testapp",
				Module:    "github.com/test/testapp",
				Framework: tc.framework,
				DB:        "none",
				Logger:    "slog",
				GoVersion: "1.23",
			}
			if err := generator.Generate(cfg, out, nil); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			data, _ := os.ReadFile(filepath.Join(out, "AGENTS.md"))
			if !strings.Contains(string(data), tc.wantRoute) {
				t.Errorf("AGENTS.md missing %q routing syntax for %s", tc.wantRoute, tc.framework)
			}
		})
	}
}

func TestGenerate_AgentsMd_LoggerUsageSyntax(t *testing.T) {
	cases := []struct {
		logger   string
		wantSnip string
	}{
		{"slog", `logger.Info(`},
		{"zap", `zap.`},
		{"zerolog", `.Msg(`},
	}
	for _, tc := range cases {
		t.Run(tc.logger, func(t *testing.T) {
			out := t.TempDir()
			cfg := config.Config{
				Name:      "testapp",
				Module:    "github.com/test/testapp",
				Framework: "gin",
				DB:        "none",
				Logger:    tc.logger,
				GoVersion: "1.23",
			}
			if err := generator.Generate(cfg, out, nil); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			data, _ := os.ReadFile(filepath.Join(out, "AGENTS.md"))
			if !strings.Contains(string(data), tc.wantSnip) {
				t.Errorf("AGENTS.md missing %q logger snippet for %s", tc.wantSnip, tc.logger)
			}
		})
	}
}

func TestGenerate_AgentsMd_ContainsLinterCommandWhenLinterExtra(t *testing.T) {
	out := t.TempDir()
	cfg := aiFilesConfig()
	cfg.Extras = []string{"linter"}
	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "AGENTS.md"))
	if !strings.Contains(string(data), "golangci-lint") {
		t.Error("AGENTS.md missing golangci-lint command when linter extra selected")
	}
}

func TestGenerate_AgentsMd_NoLinterCommandWithoutLinterExtra(t *testing.T) {
	out := t.TempDir()
	cfg := aiFilesConfig()
	cfg.Extras = []string{}
	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "AGENTS.md"))
	if strings.Contains(string(data), "golangci-lint") {
		t.Error("AGENTS.md should not contain golangci-lint when linter extra not selected")
	}
}

func TestGenerate_ReadmeMd_ContainsStackTable(t *testing.T) {
	out := t.TempDir()
	if err := generator.Generate(aiFilesConfig(), out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "README.md"))
	s := string(data)
	for _, want := range []string{"gin", "postgres-gorm", "zap", "testapp"} {
		if !strings.Contains(s, want) {
			t.Errorf("README.md missing %q", want)
		}
	}
}

func TestGenerate_ReadmeMd_ContainsEnvVarsForPostgres(t *testing.T) {
	out := t.TempDir()
	cfg := config.Config{
		Name:      "testapp",
		Module:    "github.com/test/testapp",
		Framework: "gin",
		DB:        "postgres-gorm",
		Logger:    "slog",
		GoVersion: "1.23",
	}
	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "README.md"))
	for _, want := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("README.md missing env var %q for postgres config", want)
		}
	}
}

func TestGenerate_ReadmeMd_ContainsDockerStepWhenDockerExtra(t *testing.T) {
	out := t.TempDir()
	cfg := aiFilesConfig()
	cfg.Extras = []string{"docker"}
	if err := generator.Generate(cfg, out, nil); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(out, "README.md"))
	if !strings.Contains(string(data), "docker compose up") {
		t.Error("README.md missing 'docker compose up' step when docker extra selected")
	}
}