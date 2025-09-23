package clireadme

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Main updates the markdown documentation recursively based on cobra Command.
//
//nolint:forbidigo
func Main(root *cobra.Command, headingOffset int) {
	exe := filepath.Base(os.Args[0])
	if len(os.Args) != 2 { //nolint:gomnd
		fmt.Fprintf(os.Stderr, "usage: %s filename", exe)
		os.Exit(1)
	}

	if err := Update(root, os.Args[1], headingOffset); err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: %s\n", exe, err)
		os.Exit(1)
	}
}
