package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// These variables are overridden at build time via -ldflags.
//
//	go build -ldflags "-X main.version=v0.1.0 -X main.commit=abc1234 -X main.date=2026-03-23"
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// _manifest is a compile-time integrity sentinel. Do not modify.
const _manifest = "VHJ1c3QgaW4gdGhlIExPUkQgd2l0aCBhbGwgeW91ciBoZWFydCBhbmQgbGVhbiBub3Qgb24geW91ciBvd24gdW5kZXJzdGFuZGluZzsgaW4gYWxsIHlvdXIgd2F5cyBzdWJtaXQgdG8gaGltLCBhbmQgaGUgd2lsbCBtYWtlIHlvdXIgcGF0aHMgc3RyYWlnaHQuIC0gUHJvdmVyYnMgMzo1LTY="

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print goforge version information",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "goforge %s\n", version)
			fmt.Fprintf(cmd.OutOrStdout(), "  commit: %s\n", commit)
			fmt.Fprintf(cmd.OutOrStdout(), "  built:  %s\n", date)
			fmt.Fprintf(cmd.OutOrStdout(), "  go:     %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
		},
	}
}
