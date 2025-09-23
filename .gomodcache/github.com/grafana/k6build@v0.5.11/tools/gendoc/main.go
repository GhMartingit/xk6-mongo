// Package main contains CLI documentation generator tool.
package main

import (
	"github.com/grafana/k6build/cmd"
	"github.com/grafana/k6build/internal/clireadme"
)

func main() {
	clireadme.Main(cmd.New(), 0)
}
