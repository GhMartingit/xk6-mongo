// Package clireadme contains internal CLI documentation generator.
package clireadme

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Update updates the markdown documentation recursively based on cobra Command.
func Update(root *cobra.Command, filename string, headingOffset int) error {
	var buff bytes.Buffer

	if err := generateMarkdown(root, &buff, headingOffset); err != nil {
		return err
	}

	filename = filepath.Clean(filename)

	src, err := os.ReadFile(filename) //nolint:forbidigo
	if err != nil {
		return err
	}

	res, found, err := replace(src, "cli", buff.Bytes())
	if err != nil {
		return err
	}
	if found {
		src = res
	}

	return os.WriteFile(filename, src, 0o600) //nolint:gomnd,forbidigo
}
