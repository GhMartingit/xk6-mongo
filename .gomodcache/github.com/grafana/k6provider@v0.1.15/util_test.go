package k6provider

import (
	"errors"
	"testing"
)

func Test_ParseZise(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		size      string
		expected  int64
		expextErr error
	}{
		{
			name:      "no unit",
			size:      "100",
			expected:  100,
			expextErr: nil,
		},
		{
			name:      "Kb unit",
			size:      "100Kb",
			expected:  100 * kilobytes,
			expextErr: nil,
		},
		{
			name:      "Mb unit",
			size:      "100Mb",
			expected:  100 * megabytes,
			expextErr: nil,
		},
		{
			name:      "Gb unit",
			size:      "1Gb",
			expected:  1 * gigabytes,
			expextErr: nil,
		},
		{
			name:      "empty size",
			size:      "Mb",
			expextErr: errInvalidZizeFormat,
		},
		{
			name:     "empty string",
			size:     "",
			expected: 0,
		},
		{
			name:      "non numerical size",
			size:      "1oMb",
			expextErr: errInvalidZizeFormat,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			size, err := parseSize(tt.size)
			if !errors.Is(err, tt.expextErr) {
				t.Fatalf("expected error %v got %v", tt.expextErr, err)
			}

			if size != tt.expected {
				t.Fatalf("expected size %d got %d", tt.expected, size)
			}
		})
	}
}
