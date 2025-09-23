package main

import (
	"github.com/spf13/cobra"
)

// newCmd returns a cobra.Command for k6foundry command
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k6foundry",
		Short: "k6 build tool",
		Long:  "k6foundry is a CLI tool for building custom k6 binaries with extensions",
		// prevent the usage help to printed to stderr when an error is reported by a subcommand
		SilenceUsage: true,
		// this is needed to prevent cobra to print errors reported by subcommands in the stderr
		SilenceErrors: true,
	}

	return cmd
}
