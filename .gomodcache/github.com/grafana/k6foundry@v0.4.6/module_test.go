package k6foundry

import (
	"errors"
	"testing"
)

func TestParseModule(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		title       string
		dependency  string
		expectError error
		expect      Module
	}{
		{
			title:      "path with canonical version",
			dependency: "github.com/path/module@v0.1.0",
			expect: Module{
				Path:    "github.com/path/module",
				Version: "v0.1.0",
			},
		},
		{
			title:      "path without version",
			dependency: "github.com/path/module",
			expect: Module{
				Path:    "github.com/path/module",
				Version: "",
			},
		},
		{
			title:      "path with latest version",
			dependency: "github.com/path/module@latest",
			expect: Module{
				Path:    "github.com/path/module",
				Version: "latest",
			},
		},
		{
			title:       "invalid path",
			dependency:  "github.com/@v1",
			expectError: ErrInvalidDependencyFormat,
		},
		{
			title:      "versioned replace",
			dependency: "github.com/path/module=github.com/another/module@v0.1.0",
			expect: Module{
				Path:           "github.com/path/module",
				Version:        "",
				ReplacePath:    "github.com/another/module",
				ReplaceVersion: "v0.1.0",
			},
		},
		{
			title:      "unversioned replace",
			dependency: "github.com/path/module=github.com/another/module",
			expect: Module{
				Path:           "github.com/path/module",
				Version:        "",
				ReplacePath:    "github.com/another/module",
				ReplaceVersion: "",
			},
		},
		{
			title:      "relative replace",
			dependency: "github.com/path/module=./another/module",
			expect: Module{
				Path:           "github.com/path/module",
				Version:        "",
				ReplacePath:    "./another/module",
				ReplaceVersion: "",
			},
		},
		{
			title:       "versioned relative replace",
			dependency:  "github.com/path/module=./another/module@v0.1.0",
			expectError: ErrInvalidDependencyFormat,
		},
		{
			title:      "only module name",
			dependency: "module",
			expect: Module{
				Path: "module",
			},
		},
		{
			title:       "empty module",
			dependency:  "",
			expectError: ErrInvalidDependencyFormat,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			module, err := ParseModule(tc.dependency)
			if !errors.Is(err, tc.expectError) {
				t.Fatalf("expected %v got %v", tc.expectError, err)
			}

			if tc.expectError == nil && tc.expect != module {
				t.Fatalf("expected %v got %v", tc.expect, module)
			}
		})
	}
}
