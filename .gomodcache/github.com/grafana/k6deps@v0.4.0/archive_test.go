package k6deps

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_analyzeArchive(t *testing.T) {
	t.Parallel()

	opts := &Options{
		Archive: Source{Name: filepath.Join("testdata", "archive.tar")},
	}

	actual, err := Analyze(opts)
	require.NoError(t, err)
	expected := &Dependencies{}
	_ = expected.UnmarshalText([]byte(`k6>0.54;k6/x/faker>0.4.0;k6/x/sql>=1.0.1;k6/x/sql/driver/ramsql*`))
	require.Equal(t, expected.String(), actual.String())
}

func Test_analyzeArchive_Reader(t *testing.T) {
	t.Parallel()

	file, err := os.Open(filepath.Join("testdata", "archive.tar")) //nolint:forbidigo
	require.NoError(t, err)
	defer file.Close() //nolint:errcheck

	opts := &Options{
		Archive: Source{Reader: file},
	}

	expected := &Dependencies{}
	_ = expected.UnmarshalText([]byte(`k6>0.54;k6/x/faker>0.4.0;k6/x/sql>=1.0.1;k6/x/sql/driver/ramsql*`))

	actual, err := Analyze(opts)
	require.NoError(t, err)
	require.Equal(t, expected.String(), actual.String())
}
