// Package main contains the main function for k6deps CLI tool.
package main

import (
	"log"
	"os"

	"github.com/grafana/k6deps/cmd"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	runCmd(newCmd(os.Args[1:])) //nolint:forbidigo
}

func newCmd(args []string) *cobra.Command {
	cmd := cmd.New()
	cmd.Version = version
	cmd.SetArgs(args)

	return cmd
}

func runCmd(cmd *cobra.Command) {
	log.SetFlags(0)
	log.Writer()

	if err := cmd.Execute(); err != nil {
		log.Fatal(formatError(err))
	}
}
