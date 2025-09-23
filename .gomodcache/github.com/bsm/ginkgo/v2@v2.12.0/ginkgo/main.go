package main

import (
	"fmt"
	"os"

	"github.com/bsm/ginkgo/v2/ginkgo/build"
	"github.com/bsm/ginkgo/v2/ginkgo/command"
	"github.com/bsm/ginkgo/v2/ginkgo/run"
	"github.com/bsm/ginkgo/v2/ginkgo/unfocus"
	"github.com/bsm/ginkgo/v2/ginkgo/watch"
	"github.com/bsm/ginkgo/v2/types"
)

var program command.Program

func GenerateCommands() []command.Command {
	return []command.Command{
		watch.BuildWatchCommand(),
		build.BuildBuildCommand(),
		unfocus.BuildUnfocusCommand(),
		BuildVersionCommand(),
	}
}

func main() {
	program = command.Program{
		Name:           "ginkgo",
		Heading:        fmt.Sprintf("Ginkgo Version %s", types.VERSION),
		Commands:       GenerateCommands(),
		DefaultCommand: run.BuildRunCommand(),
		DeprecatedCommands: []command.DeprecatedCommand{
			{Name: "convert", Deprecation: types.Deprecations.Convert()},
			{Name: "blur", Deprecation: types.Deprecations.Blur()},
			{Name: "nodot", Deprecation: types.Deprecations.Nodot()},
		},
	}

	program.RunAndExit(os.Args)
}

func BuildVersionCommand() command.Command {
	return command.Command{
		Name:     "version",
		Usage:    "ginkgo version",
		ShortDoc: "Print Ginkgo's version",
		Command: func(_ []string, _ []string) {
			fmt.Printf("Ginkgo Version %s\n", types.VERSION)
		},
	}
}
