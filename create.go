package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/smidjahq/goforge/internal/config"
	"github.com/smidjahq/goforge/internal/generator"
	"github.com/smidjahq/goforge/internal/goversion"
	"github.com/smidjahq/goforge/internal/postgen"
	"github.com/smidjahq/goforge/internal/prompts"
	"github.com/smidjahq/goforge/internal/ui"
	"github.com/smidjahq/goforge/internal/validator"
)

func newCreateCmd() *cobra.Command {
	var (
		flagName      string
		flagModule    string
		flagFramework string
		flagDB        string
		flagLogger    string
		flagExtras    string
		flagGoVersion string
		flagOutput    string
		flagDryRun    bool
		flagGit       bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Scaffold a new Go backend project",
		Long: `Create scaffolds a production-ready Go backend project with Clean Architecture.

Provide all flags for non-interactive mode, or run without flags to launch
the interactive TUI prompt.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var cfg config.Config
			var err error

			anyFlagSet := flagName != "" || flagModule != "" || flagFramework != "" ||
				flagDB != "" || flagLogger != "" || flagExtras != "" ||
				flagOutput != "" || cmd.Flags().Changed("go-version") || cmd.Flags().Changed("git")

			if anyFlagSet {
				// Non-interactive path: build config from flags and validate.
				cfg = config.Config{
					Name:      flagName,
					Module:    flagModule,
					Framework: flagFramework,
					DB:        flagDB,
					Logger:    flagLogger,
					Extras:    parseExtras(flagExtras),
					GoVersion: flagGoVersion,
				}
				if err = validator.Validate(cfg); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", err)
					return fmt.Errorf("invalid configuration")
				}
			} else {
				// Interactive path: launch the TUI prompt.
				// --dry-run alone falls here so the user can fill in the form
				// and preview the file tree without writing anything.
				cfg, err = prompts.Run()
				if err != nil {
					return fmt.Errorf("prompt cancelled: %w", err)
				}
			}

			outputDir := flagOutput
			if outputDir == "" {
				outputDir = cfg.Name
			}

			if flagDryRun {
				printSummary(cmd, cfg)
				return nil
			}

			spin := ui.New(cmd.OutOrStdout())
			spin.Start("Scaffolding project...")

			genErr := generator.Generate(cfg, outputDir, func(msg string) {
				spin.Update(msg + "...")
			})
			if genErr != nil {
				spin.Fail("Generation failed")
				return fmt.Errorf("generation failed: %w", genErr)
			}

			spin.Update("Running go mod tidy...")
			if err := postgen.ModTidy(outputDir); err != nil {
				spin.Fail("go mod tidy failed")
				return fmt.Errorf("post-generation failed: %w", err)
			}

			if flagGit {
				spin.Update("Initializing git repository...")
				if err := postgen.GitInit(outputDir); err != nil {
					spin.Fail("git init failed")
					return fmt.Errorf("post-generation failed: %w", err)
				}
			}

			spin.Stop(cfg.Name + " created successfully")
			printNextSteps(cmd, cfg, outputDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "project directory name (required)")
	cmd.Flags().StringVar(&flagModule, "module", "", "Go module path, e.g. github.com/you/myapp (required)")
	cmd.Flags().StringVar(&flagFramework, "framework", "", "HTTP framework: gin | chi | echo | fiber")
	cmd.Flags().StringVar(&flagDB, "db", "", "database combo: postgres-gorm | postgres-sqlc | postgres-raw | sqlite-gorm | sqlite-raw | mysql-gorm | mysql-sqlc | mysql-raw | none")
	cmd.Flags().StringVar(&flagLogger, "logger", "", "logger: slog | zap | zerolog")
	cmd.Flags().StringVar(&flagExtras, "extras", "", "comma-separated extras: docker,makefile,ci,swagger,migrations,linter")
	cmd.Flags().StringVar(&flagGoVersion, "go-version", goversion.Default(), "Go version for go.mod and Dockerfile (e.g. 1.26)")
	cmd.Flags().StringVar(&flagOutput, "output", "", "output directory (default: ./<name>)")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "print resolved config without generating files")
	cmd.Flags().BoolVar(&flagGit, "git", true, "run git init in the generated project directory")

	return cmd
}

func parseExtras(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func printSummary(cmd *cobra.Command, cfg config.Config) {
	out := cmd.OutOrStdout()

	// Config block.
	fmt.Fprintf(out, "Configuration:\n")
	fmt.Fprintf(out, "  Name:       %s\n", cfg.Name)
	fmt.Fprintf(out, "  Module:     %s\n", cfg.Module)
	fmt.Fprintf(out, "  Framework:  %s\n", cfg.Framework)
	fmt.Fprintf(out, "  DB:         %s\n", cfg.DB)
	fmt.Fprintf(out, "  Logger:     %s\n", cfg.Logger)
	fmt.Fprintf(out, "  Go version: %s\n", cfg.GoVersion)
	if len(cfg.Extras) > 0 {
		fmt.Fprintf(out, "  Extras:     %s\n", strings.Join(cfg.Extras, ", "))
	} else {
		fmt.Fprintf(out, "  Extras:     (none)\n")
	}

	// ASCII file tree.
	files := generator.ListFiles(cfg)
	fmt.Fprintf(out, "\nFile tree:\n")
	fmt.Fprintln(out, renderFileTree(cfg.Name, files))
}

// renderFileTree renders a sorted list of relative file paths as a
// lipgloss-styled ASCII tree rooted at rootName.
func renderFileTree(rootName string, files []string) string {
	dirStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	// Build a tree: map of dir → children (dirs and files).
	type node struct {
		name     string
		children []*node
		isDir    bool
	}
	root := &node{name: rootName, isDir: true}

	find := func(parent *node, name string, isDir bool) *node {
		for _, c := range parent.children {
			if c.name == name {
				return c
			}
		}
		n := &node{name: name, isDir: isDir}
		parent.children = append(parent.children, n)
		return n
	}

	for _, f := range files {
		parts := strings.Split(f, "/")
		cur := root
		for i, p := range parts {
			isDir := i < len(parts)-1
			cur = find(cur, p, isDir)
		}
	}

	var sb strings.Builder
	var walk func(n *node, prefix string, last bool)
	walk = func(n *node, prefix string, last bool) {
		connector := "├── "
		if last {
			connector = "└── "
		}
		var label string
		if n.isDir {
			label = dirStyle.Render(n.name + "/")
		} else {
			label = fileStyle.Render(n.name)
		}
		sb.WriteString(prefix + connector + label + "\n")

		childPrefix := prefix + "│   "
		if last {
			childPrefix = prefix + "    "
		}
		for i, child := range n.children {
			walk(child, childPrefix, i == len(n.children)-1)
		}
	}

	sb.WriteString(dirStyle.Render(root.name+"/") + "\n")
	for i, child := range root.children {
		walk(child, "", i == len(root.children)-1)
	}
	return sb.String()
}


func printNextSteps(cmd *cobra.Command, cfg config.Config, outputDir string) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "\n  cd %s\n", outputDir)

	hasMakefile := slices.Contains(cfg.Extras, "makefile")
	hasDocker := slices.Contains(cfg.Extras, "docker")

	if hasMakefile {
		fmt.Fprintf(out, "  make run         # start the server\n")
		if hasDocker {
			fmt.Fprintf(out, "  make compose-up  # start dependencies\n")
		}
	} else {
		fmt.Fprintf(out, "  go run ./cmd/app  # start the server\n")
	}
	fmt.Fprintln(out)
}
