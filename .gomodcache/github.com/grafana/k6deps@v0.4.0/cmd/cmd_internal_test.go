package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/grafana/k6deps"
	"github.com/stretchr/testify/require"
)

func Test_format_Set(t *testing.T) {
	t.Parallel()

	f := formatJSON

	require.Equal(t, "json|text|js", f.Type())
}

func Test_formatString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value   string
		want    format
		wantErr bool
	}{
		{value: "json", want: formatJSON},
		{value: "JSON", want: formatJSON},
		{value: "text", want: formatText},
		{value: "TEXT", want: formatText},
		{value: "js", want: formatJS},
		{value: "JS", want: formatJS},
		{value: "invalid", wantErr: true},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()

			got, err := formatString(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}

func Test_format_IsAformat(t *testing.T) {
	t.Parallel()

	for _, f := range formatValues() {
		require.True(t, f.IsAformat())
	}

	var f format = 100
	require.False(t, f.IsAformat())
}

func Test_format_String(t *testing.T) {
	t.Parallel()

	var f format = 100

	require.Equal(t, "format(100)", f.String())
	require.Equal(t, "json", formatJSON.String())
}

//nolint:forbidigo,paralleltest
func Test_deps_stdout(t *testing.T) {
	dir := t.TempDir()

	scriptfile := filepath.Join("testdata", "script.js")
	out := filepath.Clean(filepath.Join(dir, "output"))

	file, err := os.Create(out)
	require.NoError(t, err)

	saved := os.Stdout
	os.Stdout = file

	defer func() { os.Stdout = saved }()

	opts := &options{
		format: formatText,
		Options: k6deps.Options{
			Script: k6deps.Source{
				Name: scriptfile,
			},
			Manifest: k6deps.Source{
				Ignore: true,
			},
			Env: k6deps.Source{
				Ignore: true,
			},
		},
	}

	err = deps(opts, []string{scriptfile})
	require.NoError(t, err)

	require.NoError(t, file.Close())

	contents, err := os.ReadFile(out)

	require.NoError(t, err)
	require.Equal(t, `k6/x/faker>v0.3.0;xk6-top*`+"\n", string(contents))
}

func Test_deps_invalid_output(t *testing.T) {
	t.Parallel()

	out := filepath.Clean(filepath.Join("__NO__SUCH__DIRECTORY__", "output"))

	err := deps(&options{output: out}, nil)

	require.Error(t, err)
}

func Test_deps_invalid_input(t *testing.T) {
	t.Parallel()

	script := filepath.Clean(filepath.Join("testdata", "no-such-file.js"))

	err := deps(&options{Options: k6deps.Options{Script: k6deps.Source{Name: script}}}, nil)

	require.Error(t, err)
}
