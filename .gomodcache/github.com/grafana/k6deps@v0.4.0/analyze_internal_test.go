package k6deps

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmpty(t *testing.T) {
	t.Parallel()

	inst, err := newEmptyAnalyzer().analyze()
	require.NoError(t, err)
	require.Empty(t, inst)
}

func TestFilterInvalid(t *testing.T) {
	t.Parallel()

	deps := Dependencies{
		"k6":               &Dependency{Name: "k6"},
		"foo":              &Dependency{Name: "foo"},
		"bar":              &Dependency{Name: "bar"},
		"k6/x/faker":       &Dependency{Name: "k6/x/faker"},
		"xk6-foo":          &Dependency{Name: "xk6-foo"},
		"@grafana/xk6-bar": &Dependency{Name: "@grafana/xk6-bar"},
	}

	valid := filterInvalid(deps)

	require.Len(t, valid, 4)
	require.Contains(t, valid, "k6")
	require.Contains(t, valid, "k6/x/faker")
	require.Contains(t, valid, "xk6-foo")
	require.Contains(t, valid, "@grafana/xk6-bar")
}

func TestManifestAnalyzer(t *testing.T) {
	t.Parallel()

	// test empty manifest
	src := io.NopCloser(bytes.NewBuffer(nil))
	deps, err := newManifestAnalyzer(src).analyze()
	require.NoError(t, err)
	require.Empty(t, deps)

	// test manifest
	content := []byte(`{"dependencies":{"@grafana/xk6-faker":"*"}}`)
	src = io.NopCloser(bytes.NewBuffer(content))
	deps, err = newManifestAnalyzer(src).analyze()
	require.NoError(t, err)
	require.NotEmpty(t, deps)
	require.Len(t, deps, 1)
	require.Equal(t, deps["@grafana/xk6-faker"].Constraints.String(), "*")

	// test invalid manifest
	content = []byte(`{`)
	src = io.NopCloser(bytes.NewBuffer(content))
	deps, err = newManifestAnalyzer(src).analyze()
	require.Error(t, err)
	require.Len(t, deps, 0)
}

func TestScriptAnalyzer(t *testing.T) {
	t.Parallel()

	// test scanning script
	content := []byte(`"use k6 with @grafana/xk6-faker>v0.3.0";`)
	src := io.NopCloser(bytes.NewBuffer(content))
	deps, err := newScriptAnalyzer(src).analyze()
	require.NoError(t, err)
	require.NotEmpty(t, deps)
	require.Len(t, deps, 1)
	require.Equal(t, deps["@grafana/xk6-faker"].Constraints.String(), ">v0.3.0")

	// test invalid pragmas
	content = []byte(`"use k6 with k6/x/faker>>1.0";`)
	src = io.NopCloser(bytes.NewBuffer(content))
	deps, err = newScriptAnalyzer(src).analyze()
	require.Error(t, err)
	require.Empty(t, deps)
}

func TestTextAnalyzer(t *testing.T) {
	t.Parallel()

	// test empty text source
	content := []byte{}
	src := io.NopCloser(bytes.NewBuffer(content))
	deps, err := newTextAnalyzer(src).analyze()
	require.NoError(t, err)
	require.Empty(t, deps)

	// test text source
	content = []byte(`@grafana/xk6-faker>v0.3.0`)
	src = io.NopCloser(bytes.NewBuffer(content))
	deps, err = newTextAnalyzer(src).analyze()
	require.NoError(t, err)
	require.NoError(t, err)
	require.NotEmpty(t, deps)
	require.Len(t, deps, 1)
	require.Equal(t, deps["@grafana/xk6-faker"].Constraints.String(), ">v0.3.0")

	// test invalid text source
	content = []byte(`k6/x/faker>>1.0`)
	src = io.NopCloser(bytes.NewBuffer(content))
	deps, err = newTextAnalyzer(src).analyze()
	require.Error(t, err)
	require.Empty(t, deps)
}
