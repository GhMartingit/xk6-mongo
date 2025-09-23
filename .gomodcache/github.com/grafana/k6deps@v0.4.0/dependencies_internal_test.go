package k6deps

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testDeps(t *testing.T, k6 string) Dependencies {
	t.Helper()

	deps := make(Dependencies)

	dep, err := NewDependency("k6/x/yaml", ">v0.2.0")
	require.NoError(t, err)
	deps[dep.Name] = dep

	dep, err = NewDependency("k6/x/faker", ">v0.3.0")
	require.NoError(t, err)
	deps[dep.Name] = dep

	if len(k6) == 0 {
		return deps
	}

	dep, err = NewDependency("k6", k6)
	require.NoError(t, err)
	deps[dep.Name] = dep

	return deps
}

func Test_Dependencies_marshalText(t *testing.T) {
	t.Parallel()

	deps := testDeps(t, "")

	for i := 0; i < 2; i++ {
		err := deps.marshalText(&errorWriter{i})
		require.Error(t, err)
	}
}

func Test_Dependencies_sorted(t *testing.T) {
	t.Parallel()

	deps := testDeps(t, "").Sorted()

	require.Equal(t, "k6/x/faker", deps[0].Name)
	require.Equal(t, "k6/x/yaml", deps[1].Name)

	deps = testDeps(t, ">0.50").Sorted()

	require.Equal(t, "k6", deps[0].Name)
	require.Equal(t, "k6/x/faker", deps[1].Name)
	require.Equal(t, "k6/x/yaml", deps[2].Name)

	depsm := testDeps(t, ">0.50")

	dep, err := NewDependency("bar", ">v0.4.0")
	require.NoError(t, err)
	depsm[dep.Name] = dep
	deps = depsm.Sorted()

	require.Equal(t, "k6", deps[0].Name)
	require.Equal(t, "bar", deps[1].Name)
	require.Equal(t, "k6/x/faker", deps[2].Name)
	require.Equal(t, "k6/x/yaml", deps[3].Name)
}

func Test_Dependencies_update(t *testing.T) {
	t.Parallel()

	deps := make(Dependencies)

	dep, err := NewDependency("bar", ">v0.4.0")
	require.NoError(t, err)

	err = deps.update(dep)
	require.NoError(t, err)

	require.Contains(t, deps, "bar")
	require.Len(t, deps, 1)

	err = deps.update(dep)
	require.NoError(t, err)

	require.Contains(t, deps, "bar")
	require.Len(t, deps, 1)
}

func Test_Dependencies_marshalJS(t *testing.T) {
	t.Parallel()

	deps := testDeps(t, ">0.50")

	for i := 0; i < 6; i++ {
		err := deps.marshalJS(&errorWriter{i})
		require.Error(t, err)
	}
}
