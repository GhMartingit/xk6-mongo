package util

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestDownload(t *testing.T) {
	t.Parallel()

	files := fstest.MapFS{
		"file": &fstest.MapFile{Data: []byte("hello, world\n")},
	}

	fileSrv := httptest.NewServer(http.FileServerFS(files))

	testCases := []struct {
		title     string
		url       string
		path      string
		expectErr error
	}{
		{
			title: "download file",
			url:   fileSrv.URL + "/file",
			path:  filepath.Join(t.TempDir(), "file"),
		},
		{
			title:     "download non existing file",
			url:       fileSrv.URL + "/non-existing",
			path:      filepath.Join(t.TempDir(), "non-existing"),
			expectErr: ErrDownloadFailed,
		},
		{
			title:     "download to non existing file",
			url:       fileSrv.URL + "/file",
			path:      filepath.Join(t.TempDir(), "non-existing", "file"),
			expectErr: ErrWritingFile,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			err := Download(context.TODO(), tc.url, tc.path)
			if !errors.Is(err, tc.expectErr) {
				t.Errorf("expected %v, got %v", tc.expectErr, err)
			}
		})
	}
}
