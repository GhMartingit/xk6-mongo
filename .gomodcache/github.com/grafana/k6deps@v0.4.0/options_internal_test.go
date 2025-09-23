package k6deps

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptionsScriptAnalyzer(t *testing.T) {
	t.Parallel()

	// empty script source
	opts := new(Options)
	opts.Script.Contents = nil
	scriptAnalyzer, err := opts.scriptAnalyzer()
	require.NoError(t, err)
	deps, err := scriptAnalyzer.analyze()
	require.NoError(t, err)
	require.NotNil(t, deps)

	// ignore script
	opts = new(Options)
	opts.Script.Name = filepath.Join("testdata", "foo", "foo.js")
	opts.Script.Ignore = true
	scriptAnalyzer, err = opts.scriptAnalyzer()
	require.NoError(t, err)
	deps, err = scriptAnalyzer.analyze()
	require.NoError(t, err)
	require.NotNil(t, deps)

	// load script
	opts = new(Options)
	opts.Script.Name = filepath.Join("testdata", "foo", "foo.js")
	opts.Script.Ignore = false
	scriptAnalyzer, err = opts.scriptAnalyzer()
	require.NoError(t, err)
	deps, err = scriptAnalyzer.analyze()
	require.NoError(t, err)
	require.NotNil(t, deps)

	// load script with includes
	opts = new(Options)
	opts.Script.Name = filepath.Join("testdata", "combined.js")
	opts.Script.Ignore = false
	scriptAnalyzer, err = opts.scriptAnalyzer()
	require.NoError(t, err)
	deps, err = scriptAnalyzer.analyze()
	require.NoError(t, err)
	require.NotNil(t, deps)

	// load bad script
	opts = new(Options)
	opts.Script.Name = filepath.Join("testdata", "bad.js")
	scriptAnalyzer, err = opts.scriptAnalyzer()
	require.Error(t, err)
	require.Nil(t, scriptAnalyzer)

	// // load missing script
	opts = new(Options)
	opts.Script.Name = filepath.Join("testdata", "missing.js")
	scriptAnalyzer, err = opts.scriptAnalyzer()
	require.Error(t, err)
	require.Nil(t, scriptAnalyzer)
}

// This test modifies the environment so can't be run in parallel
func TestOptionsEnvAnalizer(t *testing.T) {
	opts := new(Options)

	// test empty environment
	t.Setenv(EnvDependencies, "")
	envAnalizer := opts.envAnalyzer()
	require.NotNil(t, envAnalizer)
	deps, err := envAnalizer.analyze()
	require.NoError(t, err)
	require.NotNil(t, deps)

	// test ignore environment
	opts = new(Options)
	opts.Env.Ignore = true
	t.Setenv(EnvDependencies, "k6>0.49")
	envAnalizer = opts.envAnalyzer()
	require.NotNil(t, envAnalizer)
	deps, err = envAnalizer.analyze()
	require.NoError(t, err)
	require.NotNil(t, deps)
	require.Equal(t, 0, len(deps))

	// test environment
	opts = new(Options)
	t.Setenv(EnvDependencies, "k6>0.49")
	envAnalizer = opts.envAnalyzer()
	require.NotNil(t, envAnalizer)
	deps, err = envAnalizer.analyze()
	require.NoError(t, err)
	require.NotNil(t, deps)
	require.Equal(t, 1, len(deps))
}

func TestOptionsManifestAnalyzer(t *testing.T) {
	t.Parallel()

	// empty manifest
	opts := new(Options)
	opts.Manifest.Contents = nil
	manifestAnalyzer, err := opts.manifestAnalyzer()
	require.NoError(t, err)
	deps, err := manifestAnalyzer.analyze()
	require.NoError(t, err)
	require.NotNil(t, deps)
	require.Equal(t, 0, len(deps))

	// ignore manifest
	opts = new(Options)
	opts.Manifest.Name = filepath.Join("testdata", "package.json")
	opts.Manifest.Ignore = true
	manifestAnalyzer, err = opts.manifestAnalyzer()
	require.NoError(t, err)
	deps, err = manifestAnalyzer.analyze()
	require.NoError(t, err)
	require.NotNil(t, deps)
	require.Equal(t, 0, len(deps))

	// load manifest
	opts = new(Options)
	opts.Manifest.Name = filepath.Join("testdata", "package.json")
	manifestAnalyzer, err = opts.manifestAnalyzer()
	require.NoError(t, err)
	deps, err = manifestAnalyzer.analyze()
	require.NoError(t, err)
	require.NotNil(t, deps)
	// TODO: add a dependency to the package
	require.Equal(t, 3, len(deps))

	// load empty manifest
	opts = new(Options)
	opts.Manifest.Name = filepath.Join("testdata", "empty-package.json")
	manifestAnalyzer, err = opts.manifestAnalyzer()
	require.NoError(t, err)
	deps, err = manifestAnalyzer.analyze()
	require.NoError(t, err)
	require.NotNil(t, deps)
	require.Equal(t, 0, len(deps))

	// load bad manifest
	opts = new(Options)
	opts.Manifest.Name = filepath.Join("testdata", "bad-package.json")
	manifestAnalyzer, err = opts.manifestAnalyzer()
	require.NoError(t, err)
	deps, err = manifestAnalyzer.analyze()
	require.Error(t, err)
	require.Nil(t, deps)

	// load missing manifest
	opts = new(Options)
	opts.Manifest.Name = filepath.Join("testdata", "missing.json")
	manifestAnalyzer, err = opts.manifestAnalyzer()
	require.Error(t, err)
	require.Nil(t, manifestAnalyzer)
}
