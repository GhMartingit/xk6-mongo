// package main implements the CLI root command for the k6foundry tool
package main

import (
	"fmt"
	"os"

	"github.com/grafana/k6foundry/cmd"
)

//nolint:all
func main() {
	root := newRootCmd()
	root.AddCommand(cmd.New())

	err := root.Execute()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}
}
