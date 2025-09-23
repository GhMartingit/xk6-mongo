package http

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/stretchr/testify/require"
)

func Test_plugin_onResolveRelative_error(t *testing.T) {
	t.Parallel()

	plugin := new(plugin)

	_, err := plugin.onResolveRelative(api.OnResolveArgs{
		Namespace: "http",
		Importer:  "//%4!",
	})

	require.Error(t, err)

	_, err = plugin.onResolveRelative(api.OnResolveArgs{
		Namespace: "http",
		Importer:  "http://foo.bar",
		Path:      "//%4!",
	})

	require.Error(t, err)
}

func Test_plugin_onLoad_error(t *testing.T) {
	t.Parallel()

	plugin := new(plugin)

	_, err := plugin.onLoad(api.OnLoadArgs{PluginData: "foo"})
	require.Error(t, err)
}

func Test_plugin_onResolve_error(t *testing.T) {
	t.Parallel()

	plugin := new(plugin)

	plugin.resolve = func(_ string, _ api.ResolveOptions) api.ResolveResult {
		return api.ResolveResult{}
	}

	plugin.options = &api.BuildOptions{}

	_, err := plugin.onResolve(api.OnResolveArgs{
		Namespace: "http",
		Path:      "http://localhost/foo",
	})

	require.Error(t, err)
}

func Test_plugin_load_error(t *testing.T) {
	t.Parallel()

	plugin := new(plugin)
	plugin.client = http.DefaultClient

	loc := &url.URL{Host: "127.0.0.1", Scheme: "http", Path: "/no_such_path"}

	_, err := plugin.load(loc)

	require.Error(t, err)
}
