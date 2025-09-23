// Package downloader implements utility functions for downloading objects from a store
package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/grafana/k6build"
	"github.com/grafana/k6build/pkg/store"
	"github.com/grafana/k6build/pkg/util"
)

// Download returns the content of the object
func Download(ctx context.Context, client *http.Client, object store.Object) (io.ReadCloser, error) {
	url, err := url.Parse(object.URL)
	if err != nil {
		return nil, k6build.NewWrappedError(store.ErrAccessingObject, err)
	}

	switch url.Scheme {
	case "file":
		objectPath, err := util.URLToFilePath(url)
		if err != nil {
			return nil, err
		}

		// prevent malicious path
		objectPath, err = sanitizePath(objectPath)
		if err != nil {
			return nil, err
		}

		objectFile, err := os.Open(objectPath) //nolint:gosec // path is sanitized
		if err != nil {
			// FIXME: is the path has invalid characters, still will return ErrNotExists
			if errors.Is(err, os.ErrNotExist) {
				return nil, store.ErrObjectNotFound
			}
			return nil, k6build.NewWrappedError(store.ErrAccessingObject, err)
		}

		return objectFile, nil
	case "http", "https":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, object.URL, nil)
		if err != nil {
			return nil, k6build.NewWrappedError(store.ErrAccessingObject, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, k6build.NewWrappedError(store.ErrAccessingObject, err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return nil, store.ErrObjectNotFound
		}

		if resp.StatusCode != http.StatusOK {
			return nil, k6build.NewWrappedError(store.ErrAccessingObject, fmt.Errorf("HTTP response: %s", resp.Status))
		}

		return resp.Body, nil
	default:
		return nil, fmt.Errorf("%w unsupported schema: %s", store.ErrInvalidURL, url.Scheme)
	}
}

func sanitizePath(path string) (string, error) {
	path = filepath.Clean(path)

	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("%w : invalid path %s", store.ErrInvalidURL, path)
	}

	return path, nil
}
