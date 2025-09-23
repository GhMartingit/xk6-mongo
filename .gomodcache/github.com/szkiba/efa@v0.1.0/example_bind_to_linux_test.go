package efa_test

import (
	"flag"
	"fmt"
	"os"

	"github.com/szkiba/efa"
)

// To emulate the "ANSWER=42" shell command.
func init() {
	os.Setenv("ANSWER", "42")
}

func ExampleBindTo() {
	answer := flag.Int("answer", 1, "The universal answer")

	// Must be called before parsing!
	efa.BindTo("answer", "ANSWER")

	flag.Parse()

	fmt.Println(*answer)

	// Output:
	// 42
}
