package cmd_test

import "github.com/grafana/k6deps/cmd"

func ExampleNew() {
	c := cmd.New()
	c.SetArgs([]string{"testdata/combined.js"})
	_ = c.Execute()
	// Output:
	// {"k6":">0.54","k6/x/faker":">0.4.0","k6/x/sql":">=1.0.1","k6/x/sql/driver/ramsql":"*"}
}
