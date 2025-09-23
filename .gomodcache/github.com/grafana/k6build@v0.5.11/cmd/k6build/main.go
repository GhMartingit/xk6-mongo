// package main implements the CLI root command for the k6build tool
package main

import (
	"fmt"
	"os"

	"github.com/grafana/k6build/cmd"
)

func main() {
	root := cmd.New()

	err := root.Execute()
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}
}
