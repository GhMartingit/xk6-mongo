package k6deps_test

import (
	"fmt"

	"github.com/grafana/k6deps"
)

const scriptWithPragma = `
"use k6 > 0.54";
"use k6 with k6/x/faker > 0.4.0";
"use k6 with k6/x/sql >= 1.0.1";

import { Faker } from "k6/x/faker";
import sql from "k6/x/sql";
import driver from "k6/x/sql/driver/ramsql";

export default function() {
}
`

func ExampleAnalyze_with_pragma() {
	deps, _ := k6deps.Analyze(&k6deps.Options{
		Script: k6deps.Source{
			Name:     "script.js",
			Contents: []byte(scriptWithPragma),
		},
		// disable automatic source detection
		Manifest: k6deps.Source{Ignore: true},
		Env:      k6deps.Source{Ignore: true},
	})

	fmt.Println(deps.String())

	out, _ := deps.MarshalJSON()
	fmt.Println(string(out))
	// Output:
	// k6>0.54;k6/x/faker>0.4.0;k6/x/sql>=1.0.1;k6/x/sql/driver/ramsql*
	// {"k6":">0.54","k6/x/faker":">0.4.0","k6/x/sql":">=1.0.1","k6/x/sql/driver/ramsql":"*"}
}
