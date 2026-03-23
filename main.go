package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "goforge",
		Short: "Scaffold production-ready Go backends in seconds",
		Long: `goforge is an interactive CLI that scaffolds Go backend projects
with Clean Architecture, your choice of HTTP framework, database layer, and logger.`,
	}
	root.AddCommand(newCreateCmd())
	root.AddCommand(newVersionCmd())
	return root
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
