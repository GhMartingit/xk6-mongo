package pack

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Pack(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		title       string
		script      string
		expectError bool
		meta        []string
	}{
		{
			title: "missing import",
			script: `
import { User, newUser } from "./testdata/user"
const user : User = newUser("John")
console.log(user)
`,
			expectError: true,
		},
		{
			title: "k6 imports",
			script: `
		import "k6"
		import "k6/x/foo?bar"
		import "k6/x/foo#dummy"
		`,
			expectError: false,
			meta:        []string{"k6", "k6/x/foo#dummy", "k6/x/foo?bar"},
		},
		{
			title:       "no imports",
			script:      `console.log("hello!")`,
			expectError: false,
			meta:        nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			_, meta, err := Pack(tc.script, &Options{})
			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.meta, meta.Imports)
		})
	}
}
