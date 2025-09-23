package rootfs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestRootFSOpen(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "path", "to", "file")

	err := os.MkdirAll(filepath.Dir(path), 0o750) //nolint:forbidigo
	if err != nil {
		t.Fatalf("test setup %v", err)
	}

	f, err := os.Create(path) //nolint:gosec,forbidigo
	if err != nil {
		t.Fatalf("test setup %v", err)
	}
	_ = f.Close()

	rootFs, err := NewFromDir(root)
	if err != nil {
		t.Fatalf("unexpected setting up test %v", err)
	}

	testCases := []struct {
		title  string
		path   string
		expect error
	}{
		{
			title:  "valid relative path",
			path:   filepath.Join("path", "to", "file"),
			expect: nil,
		},
		{
			title:  "valid navigation",
			path:   filepath.Join(".", "path", "to", "file"),
			expect: nil,
		},
		{
			title:  "absolute path",
			path:   filepath.Join(root, "path", "to", "file"),
			expect: nil,
		},
		{
			title:  "file does not exists",
			path:   filepath.Join("path", "to", "nonexiting"),
			expect: fs.ErrNotExist,
		},
		{
			title:  "invalid navigation",
			path:   filepath.Join("..", "scape", "to", "other", "file"),
			expect: fs.ErrInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()
			file, tErr := rootFs.Open(tc.path)
			if !errors.Is(tErr, tc.expect) {
				t.Fatalf("expected %v got %v", tc.expect, tErr)
			}

			if file != nil {
				_ = file.Close()
			}
		})
	}
}
