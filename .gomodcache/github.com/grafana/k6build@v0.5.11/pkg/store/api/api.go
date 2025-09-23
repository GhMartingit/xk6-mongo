// Package api defines the interface to a store server
package api

import (
	"errors"

	"github.com/grafana/k6build"
	"github.com/grafana/k6build/pkg/store"
)

var (
	// ErrInvalidRequest signals the request could not be processed
	// due to erroneous parameters
	ErrInvalidRequest = errors.New("invalid request")
	// ErrRequestFailed signals the request failed, probably due to a network error
	ErrRequestFailed = errors.New("request failed")
	// ErrObjectStoreAccess signals the access to the store failed
	ErrObjectStoreAccess = errors.New("store access failed")
)

// StoreResponse is the response to a store server request
type StoreResponse struct {
	Error  *k6build.WrappedError
	Object store.Object
}
