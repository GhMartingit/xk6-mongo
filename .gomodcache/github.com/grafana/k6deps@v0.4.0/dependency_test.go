package k6deps_test

import (
	"testing"

	"github.com/grafana/k6deps"
	"github.com/stretchr/testify/require"
)

func Test_NewDependency(t *testing.T) {
	t.Parallel()

	dep, err := k6deps.NewDependency("foo", "1.0")

	require.NoError(t, err)
	require.Equal(t, "foo", dep.Name)
	require.Equal(t, "1.0", dep.Constraints.String())

	_, err = k6deps.NewDependency("foo", "invalid")

	require.Error(t, err)
}

func Test_Dependency_MarshalText(t *testing.T) {
	t.Parallel()

	dep, err := k6deps.NewDependency("foo", ">v0.1.0")

	require.NoError(t, err)

	text, err := dep.MarshalText()

	require.NoError(t, err)

	require.Equal(t, "foo>v0.1.0", string(text))

	dep.Constraints = nil

	text, err = dep.MarshalText()

	require.NoError(t, err)

	require.Equal(t, "foo*", string(text))
}

func Test_Dependency_UnmarshalText(t *testing.T) {
	t.Parallel()

	dep := new(k6deps.Dependency)
	err := dep.UnmarshalText([]byte("foo>v0.3"))

	require.NoError(t, err)
	require.Equal(t, "foo", dep.Name)
	require.Equal(t, ">v0.3", dep.Constraints.String())

	dep = new(k6deps.Dependency)
	err = dep.UnmarshalText([]byte(" "))
	require.Error(t, err)
	require.ErrorIs(t, err, k6deps.ErrDependency)

	dep = new(k6deps.Dependency)
	err = dep.UnmarshalText([]byte("foo>bar"))
	require.Error(t, err)
	require.ErrorIs(t, err, k6deps.ErrConstraints)
}

func Test_Dependency_String(t *testing.T) {
	t.Parallel()

	dep, err := k6deps.NewDependency("foo", ">1.0")

	require.NoError(t, err)

	require.Equal(t, "foo>1.0", dep.String())

	text, err := dep.MarshalText()

	require.NoError(t, err)

	require.Equal(t, string(text), dep.String())
}

func Test_Dependency_MarshalJS(t *testing.T) {
	t.Parallel()

	dep, err := k6deps.NewDependency("foo", ">1.0")

	require.NoError(t, err)

	bin, err := dep.MarshalJS()

	require.NoError(t, err)
	require.Equal(t, `"use k6 with foo>1.0";`, string(bin))
}
