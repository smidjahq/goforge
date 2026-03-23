package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/smidjahq/goforge/internal/config"
	"github.com/smidjahq/goforge/internal/generator"
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
				flagDB != "" || flagLogger != "" || flagExtras != ""

			if anyFlagSet || flagDryRun {
				// Non-interactive path: build config from flags and validate.
				cfg = config.Config{
					Name:      flagName,
					Module:    flagModule,
					Framework: flagFramework,
					DB:        flagDB,
					Logger:    flagLogger,
					Extras:    parseExtras(flagExtras),
				}
				if err = validator.Validate(cfg); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", err)
					return fmt.Errorf("invalid configuration")
				}
			} else {
				// Interactive path: launch the TUI prompt.
				cfg, err = prompts.Run()
				if err != nil {
					return fmt.Errorf("prompt cancelled: %w", err)
				}
			}

			if flagDryRun {
				printSummary(cmd, cfg)
				return nil
			}

			spin := ui.New(cmd.OutOrStdout())
			spin.Start("Scaffolding project...")

			genErr := generator.Generate(cfg, cfg.Name, func(msg string) {
				spin.Update(msg + "...")
			})
			if genErr != nil {
				spin.Fail("Generation failed")
				return fmt.Errorf("generation failed: %w", genErr)
			}

			spin.Update("Running go mod tidy...")
			if err := postgen.ModTidy(cfg.Name); err != nil {
				spin.Fail("go mod tidy failed")
				return fmt.Errorf("post-generation failed: %w", err)
			}

			if flagGit {
				spin.Update("Initializing git repository...")
				if err := postgen.GitInit(cfg.Name); err != nil {
					spin.Fail("git init failed")
					return fmt.Errorf("post-generation failed: %w", err)
				}
			}

			spin.Stop(cfg.Name + " created successfully")
			printNextSteps(cmd, cfg)
			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "project directory name (required)")
	cmd.Flags().StringVar(&flagModule, "module", "", "Go module path, e.g. github.com/you/myapp (required)")
	cmd.Flags().StringVar(&flagFramework, "framework", "", "HTTP framework: gin | chi | echo | fiber")
	cmd.Flags().StringVar(&flagDB, "db", "", "database combo: postgres-gorm | postgres-sqlc | postgres-raw | sqlite-gorm | sqlite-raw | none")
	cmd.Flags().StringVar(&flagLogger, "logger", "", "logger: slog | zap | zerolog")
	cmd.Flags().StringVar(&flagExtras, "extras", "", "comma-separated extras: docker,makefile,ci,swagger,migrations,linter")
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
	fmt.Fprintf(cmd.OutOrStdout(), "Configuration:\n")
	fmt.Fprintf(cmd.OutOrStdout(), "  Name:      %s\n", cfg.Name)
	fmt.Fprintf(cmd.OutOrStdout(), "  Module:    %s\n", cfg.Module)
	fmt.Fprintf(cmd.OutOrStdout(), "  Framework: %s\n", cfg.Framework)
	fmt.Fprintf(cmd.OutOrStdout(), "  DB:        %s\n", cfg.DB)
	fmt.Fprintf(cmd.OutOrStdout(), "  Logger:    %s\n", cfg.Logger)
	if len(cfg.Extras) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "  Extras:    %s\n", strings.Join(cfg.Extras, ", "))
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "  Extras:    (none)\n")
	}
}

func printNextSteps(cmd *cobra.Command, cfg config.Config) {
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "\n  cd %s\n", cfg.Name)

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
