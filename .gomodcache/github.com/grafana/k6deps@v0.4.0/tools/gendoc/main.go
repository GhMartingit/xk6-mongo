// Package main contains CLI documentation generator tool.
package main

import (
	_ "embed"

	"github.com/grafana/k6deps/cmd"
	"github.com/grafana/k6deps/internal/clireadme"
)

func main() {
	root := cmd.New()
	clireadme.Main(root, 1)
}
