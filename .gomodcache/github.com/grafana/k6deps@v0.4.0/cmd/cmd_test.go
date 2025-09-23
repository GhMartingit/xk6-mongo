package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/grafana/k6deps/cmd"
	"github.com/stretchr/testify/require"
)

//nolint:forbidigo
func Test_Root(t *testing.T) {
	t.Parallel()

	t.Run("New", func(t *testing.T) {
		t.Parallel()

		root := cmd.New()
		require.Equal(t, "k6deps [flags] [script-file]", root.Use)
	})

	scriptfile := filepath.Join("testdata", "script.js")
	archive := filepath.Join("testdata", "archive.tar")

	testCases := []struct {
		name     string
		args     []string
		source   string
		expected string
	}{
		{
			name:     "script default format",
			args:     []string{"--ingnore-env", "--ignore-manifest"},
			source:   scriptfile,
			expected: `{"k6/x/faker":">v0.3.0","xk6-top":"*"}` + "\n",
		},
		{
			name:     "script and maifest default format",
			args:     []string{"--ingnore-env"},
			source:   scriptfile,
			expected: `{"k6/x/faker":">v0.3.0","xk6-top":">2.0"}` + "\n",
		},
		{
			name:     "script json format",
			args:     []string{"--ingnore-env", "--ignore-manifest", "--format", "json"},
			source:   scriptfile,
			expected: `{"k6/x/faker":">v0.3.0","xk6-top":"*"}` + "\n",
		},
		{
			name:     "script text format",
			args:     []string{"--ingnore-env", "--ignore-manifest", "--format", "text"},
			source:   scriptfile,
			expected: `k6/x/faker>v0.3.0;xk6-top*` + "\n",
		},
		{
			name:     "script js format",
			args:     []string{"--ingnore-env", "--ignore-manifest", "--format", "js"},
			source:   scriptfile,
			expected: `"use k6 with k6/x/faker>v0.3.0";` + "\n" + `"use k6 with xk6-top*";` + "\n",
		},
		{
			name:     "archive",
			args:     []string{"--ingnore-env", "--ignore-manifest"},
			source:   archive,
			expected: `{"k6":">0.54","k6/x/faker":">0.4.0","k6/x/sql":">=1.0.1","k6/x/sql/driver/ramsql":"*"}` + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out := filepath.Clean(filepath.Join(t.TempDir(), "output"))

			root := cmd.New()
			args := tc.args
			args = append(args, "-o", out, tc.source)
			root.SetArgs(args)
			err := root.Execute()
			require.NoError(t, err)

			contents, err := os.ReadFile(out)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(contents))
		})
	}

	t.Run("using input", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name     string
			source   string
			input    string
			expected string
		}{
			{
				name:     "script",
				input:    "js",
				source:   scriptfile,
				expected: `{"k6/x/faker":">v0.3.0","xk6-top":"*"}` + "\n",
			},
			{
				name:     "archive",
				input:    "tar",
				source:   archive,
				expected: `{"k6":">0.54","k6/x/faker":">0.4.0","k6/x/sql":">=1.0.1","k6/x/sql/driver/ramsql":"*"}` + "\n",
			},
		}

		// the following tests cannot be executed in parallel because they modify the stdin
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var err error
				stdin := os.Stdin
				t.Cleanup(func() { os.Stdin = stdin })

				out := filepath.Clean(filepath.Join(t.TempDir(), "output"))

				root := cmd.New()
				os.Stdin, err = os.Open(tc.source)
				if err != nil {
					t.Fatal(err)
				}
				root.SetArgs([]string{"--ingnore-env", "--ignore-manifest", "--input", tc.input, "--format", "json", "-o", out})
				err = root.Execute()

				require.NoError(t, err)

				contents, err := os.ReadFile(out)
				require.NoError(t, err)
				require.Equal(t, tc.expected, string(contents))
			})
		}
	})
}
