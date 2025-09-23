// Package file implements a file-backed object store
package file

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafana/k6build"
	"github.com/grafana/k6build/pkg/store"
	"github.com/grafana/k6build/pkg/util"
)

// Store a ObjectStore backed by a file system
type Store struct {
	dir string
}

// NewTempFileStore creates a file object store using a temporary file
func NewTempFileStore() (store.ObjectStore, error) {
	return NewFileStore(filepath.Join(os.TempDir(), "k6build", "objectstore"))
}

// NewFileStore creates an object store backed by a directory
func NewFileStore(dir string) (store.ObjectStore, error) {
	err := os.MkdirAll(dir, 0o750)
	if err != nil {
		return nil, k6build.NewWrappedError(store.ErrInitializingStore, err)
	}

	return &Store{
		dir: dir,
	}, nil
}

// Put stores the object and returns the metadata
// Fails if the object already exists
func (f *Store) Put(_ context.Context, id string, content io.Reader) (store.Object, error) {
	if id == "" {
		return store.Object{}, fmt.Errorf("%w: id cannot be empty", store.ErrCreatingObject)
	}

	if strings.Contains(id, "/") {
		return store.Object{}, fmt.Errorf("%w id cannot contain '/'", store.ErrCreatingObject)
	}

	objectDir := filepath.Join(f.dir, id)

	if _, err := os.Stat(objectDir); !errors.Is(err, os.ErrNotExist) {
		return store.Object{}, fmt.Errorf("%w: %q", store.ErrDuplicateObject, id)
	}

	// TODO: check permissions
	err := os.MkdirAll(objectDir, 0o750)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrCreatingObject, err)
	}

	// prevent concurrent modification of an object
	unlock, err := f.lockObject(id)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrCreatingObject, err)
	}
	defer unlock()

	objectFile, err := os.Create(filepath.Join(objectDir, "data")) //nolint:gosec
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrCreatingObject, err)
	}
	defer objectFile.Close() //nolint:errcheck

	// write content to object file and copy to buffer to calculate checksum
	// TODO: optimize memory by copying content in blocks
	buff := bytes.Buffer{}
	_, err = io.Copy(objectFile, io.TeeReader(content, &buff))
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrCreatingObject, err)
	}

	// calculate checksum
	checksum := fmt.Sprintf("%x", sha256.Sum256(buff.Bytes()))

	// write metadata
	err = os.WriteFile(filepath.Join(objectDir, "checksum"), []byte(checksum), 0o644) //nolint:gosec
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrCreatingObject, err)
	}

	objectURL, _ := util.URLFromFilePath(objectFile.Name())
	return store.Object{
		ID:       id,
		Checksum: checksum,
		URL:      objectURL.String(),
	}, nil
}

// Get retrieves an objects if exists in the object store or an error otherwise
func (f *Store) Get(_ context.Context, id string) (store.Object, error) {
	objectDir := filepath.Join(f.dir, id)
	_, err := os.Stat(objectDir)

	if errors.Is(err, os.ErrNotExist) {
		return store.Object{}, fmt.Errorf("%w (%s)", store.ErrObjectNotFound, id)
	}

	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrAccessingObject, err)
	}

	// prevent accessing object while is being written
	unlock, err := f.lockObject(id)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrCreatingObject, err)
	}
	defer unlock()

	checksum, err := os.ReadFile(filepath.Join(objectDir, "checksum")) //nolint:gosec
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrAccessingObject, err)
	}

	objectURL, err := util.URLFromFilePath(filepath.Join(objectDir, "data"))
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrAccessingObject, err)
	}
	return store.Object{
		ID:       id,
		Checksum: string(checksum),
		URL:      objectURL.String(),
	}, nil
}

// lockObject creates a lock for an object's directory using a file lock
func (f *Store) lockObject(id string) (func(), error) {
	objLock := newDirLock(filepath.Join(f.dir, id))
	if err := objLock.lock(0); err != nil {
		return nil, err
	}

	return func() {
		_ = objLock.unlock()
	}, nil
}
