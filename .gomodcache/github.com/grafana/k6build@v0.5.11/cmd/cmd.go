// Package cmd contains build cobra command factory function.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/grafana/k6build/cmd/local"
	"github.com/grafana/k6build/cmd/remote"
	"github.com/grafana/k6build/cmd/server"
	"github.com/grafana/k6build/cmd/store"
)

// New creates a new root command for k6build
func New() *cobra.Command {
	root := &cobra.Command{
		Use:               "k6build",
		Short:             "Build custom k6 binaries with extensions",
		SilenceUsage:      true,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	root.AddCommand(store.New())
	root.AddCommand(remote.New())
	root.AddCommand(local.New())
	root.AddCommand(server.New())

	return root
}
