package efa_test

import (
	"flag"
	"fmt"
	"os"

	"github.com/szkiba/efa"
)

// To emulate the "EFA_TEST_QUESTION='To be, or not to be?'" shell command.
func init() {
	// Note: the go test framework sets the executable name to "efa.test".
	os.Setenv("EFA_TEST_QUESTION", "To be, or not to be?")
}

func ExampleBind() {
	question := flag.String("question", "How many?", "The question")

	// Must be called before parsing!
	efa.Bind("question")

	flag.Parse()

	fmt.Println(*question)

	// Output:
	// To be, or not to be?
}
