package k6deps

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Dependency_getConstraints(t *testing.T) {
	t.Parallel()

	dep := new(Dependency)

	require.Equal(t, defaultConstraints, dep.GetConstraints())

	dep, err := NewDependency("foo", "1.0")

	require.NoError(t, err)
	require.Equal(t, "1.0", dep.GetConstraints().String())
}

func Test_Dependency_update(t *testing.T) {
	t.Parallel()

	newdep := func(name, constraints string) *Dependency {
		dep, err := NewDependency(name, constraints)

		require.NoError(t, err)
		return dep
	}

	tests := []struct {
		name    string
		from    *Dependency
		to      *Dependency
		wantErr bool
	}{
		{name: "reqular", from: newdep("foo", "1.0"), to: newdep("foo", "")},
		{name: "reqular", from: newdep("foo", ""), to: newdep("foo", "1.0")},
		{name: "reqular", from: newdep("foo", "1.0"), to: newdep("foo", "1.0")},
		{name: "reqular", from: newdep("foo", "1.0"), to: newdep("foo", "2.0"), wantErr: true},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.to.update(tt.from)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type errorWriter struct {
	count int
}

func (e *errorWriter) Write([]byte) (int, error) {
	if e.count--; e.count >= 0 {
		return 0, nil
	}

	return 0, io.ErrUnexpectedEOF
}

func Test_Dependency_marshalText(t *testing.T) {
	t.Parallel()

	dep := new(Dependency)

	err := dep.marshalText(&errorWriter{1})

	require.Error(t, err)
}

func Test_Dependency_marshalJS(t *testing.T) {
	t.Parallel()

	dep := new(Dependency)

	for i := 0; i < 5; i++ {
		err := dep.marshalJS(&errorWriter{i})
		require.Error(t, err)
	}
}

func Test_reDependency(t *testing.T) {
	t.Parallel()

	var d Dependency

	require.NoError(t, d.UnmarshalText([]byte("k6*")))
	require.Equal(t, "*", d.Constraints.String())

	require.NoError(t, d.UnmarshalText([]byte("k6 >= v0.55")))
	require.Equal(t, ">=v0.55", d.Constraints.String())

	require.NoError(t, d.UnmarshalText([]byte("k6 v0.0.0+135f85b")))
	require.Equal(t, "v0.0.0+135f85b", d.Constraints.String())

	require.NoError(t, d.UnmarshalText([]byte("k6 v0.0.0+90bb941")))
	require.Equal(t, "v0.0.0+90bb941", d.Constraints.String())
}
