// Package main contains docsme's own build-time documentation generation tool.
package main

import (
	"github.com/spf13/cobra"
	"github.com/szkiba/docsme"
)

func main() {
	cobra.CheckErr(docsme.For(nil).Execute())
}
