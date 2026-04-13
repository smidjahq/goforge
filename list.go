package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/smidjahq/goforge/internal/validator"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <category>",
		Short: "List available options for a category",
		Long: `List prints the valid values for a given category.

Categories:
  frameworks   HTTP frameworks (gin, chi, echo, fiber)
  dbs          Database combos accepted by --db
  loggers      Logger implementations
  extras       Extra options accepted by --extras
  presets      Named project presets (reserved for future use)`,
		ValidArgs: []string{"frameworks", "dbs", "loggers", "extras", "presets"},
		Args:      cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var items []string
			switch strings.ToLower(args[0]) {
			case "frameworks":
				items = validator.Frameworks()
			case "dbs":
				items = validator.DBCombinations()
			case "loggers":
				items = validator.Loggers()
			case "extras":
				items = validator.Extras()
			case "presets":
				// Reserved for T2-8; no presets defined yet.
				fmt.Fprintln(cmd.OutOrStdout(), "(no presets defined)")
				return nil
			default:
				return fmt.Errorf("unknown category %q; valid: frameworks, dbs, loggers, extras, presets", args[0])
			}
			for _, item := range items {
				fmt.Fprintln(cmd.OutOrStdout(), item)
			}
			return nil
		},
	}
	return cmd
}
