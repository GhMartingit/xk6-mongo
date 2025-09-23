//go:build !windows
// +build !windows

package k6provider

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"
)

func TestPruner(t *testing.T) {
	t.Parallel()

	binaries := map[string]fstest.MapFile{
		"binary-1": {
			Data:    make([]byte, 256),
			ModTime: time.Now(),
		},
		"binary-2": {
			Data:    make([]byte, 256),
			ModTime: time.Now().Add(-2 * time.Hour),
		},
		"binary-3": {
			Data:    make([]byte, 256),
			ModTime: time.Now().Add(-time.Hour),
		},
		"binary-4": {
			Data:    make([]byte, 256),
			ModTime: time.Now().Add(-30 * time.Minute),
		},
	}

	testCases := []struct {
		title     string
		hwm       int64
		lastPrune time.Time
		expectErr error
		expect    []string
	}{
		{
			title:     "prune least recent file",
			hwm:       256 * 3,
			expectErr: nil,
			expect:    []string{"binary-1", "binary-3", "binary-4"},
		},
		{
			title:     "should not prune before prune interval passes",
			lastPrune: time.Now(),
			hwm:       256 * 3,
			expectErr: nil,
			expect:    []string{"binary-1", "binary-2", "binary-3", "binary-4"},
		},
		{
			title:     "hwm not exceeded",
			hwm:       256 * 5,
			expectErr: nil,
			expect:    []string{"binary-1", "binary-2", "binary-3", "binary-4"},
		},
		{
			title:     "multiple binaries pruned",
			hwm:       256 * 2,
			expectErr: nil,
			expect:    []string{"binary-1", "binary-4"},
		},
		{
			title:     "fail to delete file",
			hwm:       256,
			expectErr: nil,
			expect:    []string{"binary-4"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			for path, file := range binaries {
				err := os.MkdirAll(filepath.Join(tmpDir, path), 0o750)
				if err != nil {
					t.Fatalf("test setup: creating dir %v", err)
				}
				err = os.WriteFile(filepath.Join(tmpDir, path, k6Binary), file.Data, 0o600)
				if err != nil {
					t.Fatalf("test setup writing file %v", err)
				}
				err = os.Chtimes(filepath.Join(tmpDir, path, k6Binary), file.ModTime, file.ModTime)
				if err != nil {
					t.Fatalf("test setup changing mod timestamp %v", err)
				}
			}
			// mark binary-4 as read only and revert at test end to prevent cleanup failure
			_ = os.Chmod(filepath.Join(tmpDir, "binary-4"), 0o500)
			t.Cleanup(func() {
				_ = os.Chmod(filepath.Join(tmpDir, "binary-4"), 0o750)
			})

			pruner := NewPruner(tmpDir, tc.hwm, time.Hour)
			// force time of last prune
			pruner.lastPrune = tc.lastPrune

			err := pruner.Prune()

			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}

			for _, binary := range tc.expect {
				_, err = os.Stat(filepath.Join(tmpDir, binary))
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
