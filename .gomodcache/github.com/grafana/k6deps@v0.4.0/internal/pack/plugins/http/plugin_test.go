package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
	phttp "github.com/grafana/k6deps/internal/pack/plugins/http"
	"github.com/stretchr/testify/require"
)

func Test_plugin_load(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		location string
		wantErr  bool
	}{
		{name: "without_extension", location: "http://127.0.0.1/user"},
		{name: "with_extension", location: "http://127.0.0.1/user.ts"},
		{name: "with_directory_index", location: "http://127.0.0.1/users"},
		{name: "with_mime", location: "http://127.0.0.1/user.js"},
		// error
		{name: "http_from_localhost", location: "http://localhost/user", wantErr: true},
		{name: "not_found", location: "http://127.0.0.1/no_such_file", wantErr: true},
		{name: "invalid_url", location: "http://%4!", wantErr: true},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			script := /*ts*/ `
			import { User, newUser } from "` + tt.location + `"
			
			export default function() {
				const user : User = newUser("John")
				return user
			}
			`

			result := pack(t, script)

			if tt.wantErr {
				require.NotEmpty(t, result.Errors)
			} else {
				require.Empty(t, result.Errors)
			}
		})
	}
}

func pack(t *testing.T, script string) api.BuildResult {
	t.Helper()

	const httpBase = "http://127.0.0.1"

	if strings.Contains(script, httpBase) {
		server := httptest.NewServer(http.FileServer(http.Dir("testdata")))
		defer server.Close()

		script = strings.ReplaceAll(script, httpBase, server.URL)
	}

	return api.Build(api.BuildOptions{
		Bundle: true,
		Stdin: &api.StdinOptions{
			Contents:   script,
			ResolveDir: ".",
			Sourcefile: t.Name(),
			Loader:     api.LoaderTS,
		},
		LogLevel: api.LogLevelSilent,
		Plugins:  []api.Plugin{phttp.New()},
		External: []string{"k6", "https://jslib.k6.io"},
	})
}
