# clireadme

> This code is copied from https://github.com/grafana/clireadme. It is embedded to use the same cobra version than k6deps uses.

A small library that helps to update the documentation in the README file of [cobra](https://github.com/spf13/cobra)-based CLI tools.

## Usage

Create a `tools/gendoc/main.go` file with the following content (the Main function must be called with your own cobra Command as a parameter).

Then run: `go run ./tools/gendoc README.md`

```go
// Package main contains CLI documentation generator tool.
package main

import (
	"github.com/grafana/clireadme"
	"github.com/grafana/k6tb/cmd"
)

func main() {
	clireadme.Main(cmd.New())
}

```