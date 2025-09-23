// Package store defines the interface of an object store service
package store

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
)

var (
	ErrAccessingObject   = errors.New("accessing object") //nolint:revive
	ErrCreatingObject    = errors.New("creating object")
	ErrInitializingStore = errors.New("initializing store")
	ErrInvalidURL        = errors.New("invalid object URL")
	ErrObjectNotFound    = errors.New("object not found")
	ErrNotSupported      = errors.New("not supported")
	ErrDuplicateObject   = errors.New("duplicate object")
)

// Object represents an object stored in the store
// TODO: add metadata (e.g creation data, size)
type Object struct {
	ID       string
	Checksum string
	// an url for downloading the object's content
	URL string
}

func (o Object) String() string {
	buffer := &bytes.Buffer{}
	buffer.WriteString(fmt.Sprintf("id: %s", o.ID))
	buffer.WriteString(fmt.Sprintf(" checksum: %s", o.Checksum))
	buffer.WriteString(fmt.Sprintf("url: %s", o.URL))

	return buffer.String()
}

// ObjectStore defines an interface for storing and retrieving blobs
type ObjectStore interface {
	// Get retrieves an objects if exists in the store or an error otherwise
	Get(ctx context.Context, id string) (Object, error)
	// Put stores the object and returns the metadata
	Put(ctx context.Context, id string, content io.Reader) (Object, error)
}
