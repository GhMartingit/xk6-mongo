package downloader

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/grafana/k6build/pkg/store"
	"github.com/grafana/k6build/pkg/util"
)

func fileURL(dir string, path string) string {
	url, err := util.URLFromFilePath(filepath.Join(dir, path))
	if err != nil {
		panic(err)
	}
	return url.String()
}

func httpURL(srv *httptest.Server, path string) string {
	return srv.URL + "/" + path
}

func TestDownload(t *testing.T) {
	t.Parallel()

	storeDir := t.TempDir()

	objects := []struct {
		id      string
		content []byte
	}{
		{
			id:      "object",
			content: []byte("content"),
		},
	}

	for _, o := range objects {
		if err := os.WriteFile(filepath.Join(storeDir, o.id), o.content, 0o600); err != nil {
			t.Fatalf("test setup %v", err)
		}
	}

	srv := httptest.NewServer(http.FileServer(http.Dir(storeDir)))
	t.Cleanup(srv.Close)

	testCases := []struct {
		title     string
		id        string
		url       string
		expected  []byte
		expectErr error
	}{
		{
			title:     "download file url",
			id:        "object",
			url:       fileURL(storeDir, "object"),
			expected:  []byte("content"),
			expectErr: nil,
		},
		{
			title:     "download non existing file url",
			id:        "object",
			url:       fileURL(storeDir, "another_object"),
			expectErr: store.ErrObjectNotFound,
		},
		//  FIXME: can't check url is outside object store's directory
		// {
		// 	title:     "download malicious file url",
		// 	id:        "object",
		// 	url:       fileURL(storeDir, "/../../object"),
		// 	expectErr: store.ErrInvalidURL,
		// },
		{
			title:     "download http url",
			id:        "object",
			url:       httpURL(srv, "object"),
			expected:  []byte("content"),
			expectErr: nil,
		},
		{
			title:     "download non existing http url",
			id:        "object",
			url:       httpURL(srv, "another-object"),
			expectErr: store.ErrObjectNotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			object := store.Object{ID: tc.id, URL: tc.url}
			content, err := Download(context.TODO(), http.DefaultClient, object)
			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}

			// if expected error, don't check returned object
			if tc.expectErr != nil {
				return
			}

			defer content.Close() //nolint:errcheck

			data := bytes.Buffer{}
			_, err = data.ReadFrom(content)
			if err != nil {
				t.Fatalf("reading content: %v", err)
			}

			if !bytes.Equal(data.Bytes(), tc.expected) {
				t.Fatalf("expected %v got %v", tc.expected, data)
			}
		})
	}
}
