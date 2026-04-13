// Package prompts contains the interactive huh-based TUI for goforge create.
package prompts

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/smidjahq/goforge/internal/config"
	"github.com/smidjahq/goforge/internal/goversion"
	"github.com/smidjahq/goforge/internal/validator"
)

// steps is the ordered list of step labels shown in the progress header.
var steps = []string{
	"Project Identity",
	"Framework & Database",
	"Logger & Extras",
}

// Run launches the interactive grouped form and returns a validated Config.
// It is called when the user runs `goforge create` without all required flags.
func Run() (config.Config, error) {
	var (
		name      string
		module    string
		framework string
		dbBackend string
		dbLayer   string
		loggerVal string
		extras    []string
		goVersion = goversion.Default()
	)

	theme := buildTheme()

	// Step 1 — project identity
	printStepHeader(1)
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Placeholder("myapp").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("project name is required")
					}
					return nil
				}).
				Value(&name),

			huh.NewInput().
				Title("Go module path").
				Placeholder("github.com/you/myapp").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("module path is required")
					}
					return nil
				}).
				Value(&module),

			huh.NewInput().
				Title("Go version").
				Placeholder(goVersion).
				Value(&goVersion),
		),
	).WithTheme(theme).Run(); err != nil {
		return config.Config{}, err
	}

	// Step 2 — framework + database
	printStepHeader(2)
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("HTTP framework").
				Options(
					huh.NewOption("Gin", "gin"),
					huh.NewOption("Chi", "chi"),
					huh.NewOption("Echo (labstack/echo)", "echo"),
					huh.NewOption("Fiber (gofiber/fiber)", "fiber"),
				).
				Value(&framework),

			huh.NewSelect[string]().
				Title("Database backend").
				Options(
					huh.NewOption("PostgreSQL", "postgres"),
					huh.NewOption("SQLite", "sqlite"),
					huh.NewOption("MySQL / MariaDB", "mysql"),
					huh.NewOption("None (no database)", "none"),
				).
				Value(&dbBackend),

			// DB layer options update in real time when dbBackend changes.
			huh.NewSelect[string]().
				Title("Database layer").
				OptionsFunc(func() []huh.Option[string] {
					return layerOptions(dbBackend)
				}, &dbBackend).
				Value(&dbLayer),
		),
	).WithTheme(theme).Run(); err != nil {
		return config.Config{}, err
	}

	// Step 3 — logger + extras
	printStepHeader(3)
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Logger").
				Options(
					huh.NewOption("slog (stdlib, no extra deps)", "slog"),
					huh.NewOption("Zap (uber-go/zap)", "zap"),
					huh.NewOption("Zerolog (rs/zerolog)", "zerolog"),
				).
				Value(&loggerVal),

			huh.NewMultiSelect[string]().
				Title("Extras  (space to toggle, enter to confirm)").
				Options(extrasOptions()...).
				Validate(func(selected []string) error {
					for _, s := range selected {
						if s == "migrations" && dbBackend == "none" {
							return errors.New("migrations extra requires a database (DB is set to none)")
						}
					}
					return nil
				}).
				Value(&extras),
		),
	).WithTheme(theme).Run(); err != nil {
		return config.Config{}, err
	}

	db := buildDB(dbBackend, dbLayer)

	cfg := config.Config{
		Name:      strings.TrimSpace(name),
		Module:    strings.TrimSpace(module),
		Framework: framework,
		DB:        db,
		Logger:    loggerVal,
		Extras:    extras,
		GoVersion: strings.TrimSpace(goVersion),
	}

	if err := validator.Validate(cfg); err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}

// printStepHeader clears the screen and prints the step progress indicator.
// current is 1-based.
func printStepHeader(current int) {
	// Clear screen.
	fmt.Fprint(os.Stdout, "\033[H\033[2J")

	completedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("2")).
		Faint(true)

	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15"))

	upcomingStyle := lipgloss.NewStyle().
		Faint(true)

	checkMark := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓")
	arrow := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true).Render("▶")

	fmt.Fprintln(os.Stdout)
	for i, label := range steps {
		n := i + 1
		switch {
		case n < current:
			// Completed step.
			line := fmt.Sprintf("  %s  %d · %s", checkMark, n, label)
			fmt.Fprintln(os.Stdout, completedStyle.Render(line))
		case n == current:
			// Active step.
			line := fmt.Sprintf("  %s  %d · %s", arrow, n, label)
			fmt.Fprintln(os.Stdout, activeStyle.Render(line))
		default:
			// Upcoming step.
			line := fmt.Sprintf("     %d · %s", n, label)
			fmt.Fprintln(os.Stdout, upcomingStyle.Render(line))
		}
	}
	fmt.Fprintln(os.Stdout)
}

// buildTheme returns a huh theme with square-bracket checkboxes for MultiSelect.
func buildTheme() *huh.Theme {
	t := huh.ThemeCharm()
	t.Focused.SelectedPrefix = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#02CF92", Dark: "#02A877"}).
		SetString("[x] ")
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "", Dark: "243"}).
		SetString("[ ] ")
	t.Blurred.SelectedPrefix = lipgloss.NewStyle().
		Faint(true).
		SetString("[x] ")
	t.Blurred.UnselectedPrefix = lipgloss.NewStyle().
		Faint(true).
		SetString("[ ] ")
	return t
}

// layerOptions returns huh options for the DB layer select based on the
// selected backend. When the backend is "none", a placeholder is returned
// so the select still renders; buildDB ignores the layer value in that case.
func layerOptions(backend string) []huh.Option[string] {
	labels := map[string]string{
		"gorm": "GORM",
		"sqlc": "sqlc (code generation)",
		"raw":  "raw database/sql",
	}

	layers := validator.ValidOptionsFor(backend)
	if len(layers) == 0 {
		return []huh.Option[string]{huh.NewOption("—  (no layer needed)", "")}
	}

	opts := make([]huh.Option[string], len(layers))
	for i, l := range layers {
		label := labels[l]
		if label == "" {
			label = l
		}
		opts[i] = huh.NewOption(label, l)
	}
	return opts
}

// extrasOptions returns all available extras as huh MultiSelect options.
func extrasOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("Docker  (Dockerfile + compose override)", "docker"),
		huh.NewOption("Makefile  (run, test, lint, migrate targets)", "makefile"),
		huh.NewOption("CI  (GitHub Actions workflow)", "ci"),
		huh.NewOption("Swagger  (docs/ stub)", "swagger"),
		huh.NewOption("Migrations  (golang-migrate folder)", "migrations"),
		huh.NewOption("Linter  (.golangci.yml)", "linter"),
	}
}


// buildDB combines backend and layer into the Config.DB format.
func buildDB(backend, layer string) string {
	if backend == "none" || layer == "" {
		return backend
	}
	return backend + "-" + layer
}
