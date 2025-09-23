package efa_test

import (
	"flag"
	"fmt"

	"github.com/szkiba/efa"
)

func ExampleEfa_Bind() {
	env := map[string]string{"CLI_FOO": "5", "CLI_BAR": "7"}
	lookup := func(name string) (string, bool) {
		value, found := env[name]

		return value, found
	}

	flags := flag.NewFlagSet("example", flag.ContinueOnError)

	e := efa.New(flags, "cli", lookup)

	foo := flags.Int("foo", 1, "foo value")
	bar := flags.String("bar", "?", "bar value")

	// Must be called before parsing!
	e.Bind("foo", "bar")

	flags.Parse([]string{""})

	fmt.Println(*foo)
	fmt.Println(*bar)

	// Output:
	// 5
	// 7
}
