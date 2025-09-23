package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/grafana/k6build/pkg/store"
	"github.com/grafana/k6build/pkg/util"
)

type object struct {
	id      string
	content []byte
}

func setupStore(path string, preload []object) (store.ObjectStore, error) {
	store, err := NewFileStore(path)
	if err != nil {
		return nil, fmt.Errorf("test setup %w", err)
	}

	for _, o := range preload {
		_, err = store.Put(context.TODO(), o.id, bytes.NewBuffer(o.content))
		if err != nil {
			return nil, fmt.Errorf("test setup %w", err)
		}
	}

	return store, nil
}

func TestFileStoreStoreObject(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		title     string
		preload   []object
		id        string
		content   []byte
		expectErr error
	}{
		{
			title:   "store object",
			id:      "object",
			content: []byte("content"),
		},
		{
			title: "store existing object",
			preload: []object{
				{
					id:      "object",
					content: []byte("content"),
				},
			},
			id:        "object",
			content:   []byte("new content"),
			expectErr: store.ErrDuplicateObject,
		},
		{
			title:   "store empty object",
			id:      "empty",
			content: nil,
		},
		{
			title:     "store empty id",
			id:        "",
			content:   []byte("content"),
			expectErr: store.ErrCreatingObject,
		},
		{
			title:     "store invalid id (dot slash)",
			id:        "./invalid",
			content:   []byte("content"),
			expectErr: store.ErrCreatingObject,
		},
		{
			title:     "store invalid id (trailing slash)",
			id:        "invalid/",
			content:   []byte("content"),
			expectErr: store.ErrCreatingObject,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()
			store, err := setupStore(t.TempDir(), tc.preload)
			if err != nil {
				t.Fatalf("test setup: %v", err)
			}

			obj, err := store.Put(context.TODO(), tc.id, bytes.NewBuffer(tc.content))
			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}

			// if expected error, don't validate object
			if tc.expectErr != nil {
				return
			}

			objectURL, _ := url.Parse(obj.URL)
			filePath, err := util.URLToFilePath(objectURL)
			if err != nil {
				t.Fatalf("invalid url %v", err)
			}

			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("reading object url %v", err)
			}

			if !bytes.Equal(tc.content, content) {
				t.Fatalf("expected %v got %v", tc.content, content)
			}
		})
	}
}

func TestFileStoreGet(t *testing.T) {
	t.Parallel()

	preload := []object{
		{
			id:      "object",
			content: []byte("content"),
		},
	}

	storeDir := t.TempDir()
	fileStore, err := setupStore(storeDir, preload)
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}

	testCases := []struct {
		title     string
		id        string
		expected  []byte
		expectErr error
	}{
		{
			title:     "retrieve existing object",
			id:        "object",
			expected:  []byte("content"),
			expectErr: nil,
		},
		{
			title:     "retrieve non existing object",
			id:        "another object",
			expectErr: store.ErrObjectNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			obj, err := fileStore.Get(context.TODO(), tc.id)
			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}

			// if expected error, don't check returned object
			if tc.expectErr != nil {
				return
			}

			objectURL, _ := url.Parse(obj.URL)
			fileUPath, err := util.URLToFilePath(objectURL)
			if err != nil {
				t.Fatalf("invalid url %v", err)
			}

			data, err := os.ReadFile(fileUPath)
			if err != nil {
				t.Fatalf("reading object url %v", err)
			}

			if !bytes.Equal(data, tc.expected) {
				t.Fatalf("expected %v got %v", tc.expected, data)
			}
		})
	}
}
