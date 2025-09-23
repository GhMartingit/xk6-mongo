// Package client implements an object store service client
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/grafana/k6build"
	"github.com/grafana/k6build/pkg/store"
	"github.com/grafana/k6build/pkg/store/api"
)

// ErrInvalidConfig signals an error with the client configuration
var ErrInvalidConfig = errors.New("invalid configuration")

// StoreClientConfig defines the configuration for accessing a remote object store service
type StoreClientConfig struct {
	Server     string
	HTTPClient *http.Client
}

// StoreClient access blobs in a StoreServer
type StoreClient struct {
	server *url.URL
	client *http.Client
}

// NewStoreClient returns a client for an object store server
func NewStoreClient(config StoreClientConfig) (*StoreClient, error) {
	srvURL, err := url.Parse(config.Server)
	if err != nil {
		return nil, k6build.NewWrappedError(ErrInvalidConfig, err)
	}

	client := config.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	return &StoreClient{
		server: srvURL,
		client: client,
	}, nil
}

// Get retrieves an objects if exists in the store or an error otherwise
func (c *StoreClient) Get(ctx context.Context, id string) (store.Object, error) {
	reqURL := *c.server.JoinPath("store", id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(api.ErrInvalidRequest, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(api.ErrRequestFailed, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return store.Object{}, store.ErrObjectNotFound
		}
		return store.Object{}, k6build.NewWrappedError(api.ErrRequestFailed, fmt.Errorf("status %s", resp.Status))
	}

	storeResponse := api.StoreResponse{}
	err = json.NewDecoder(resp.Body).Decode(&storeResponse)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(api.ErrRequestFailed, err)
	}

	if storeResponse.Error != nil {
		return store.Object{}, storeResponse.Error
	}

	return storeResponse.Object, nil
}

// Put stores the object and returns the metadata
func (c *StoreClient) Put(ctx context.Context, id string, content io.Reader) (store.Object, error) {
	reqURL := *c.server.JoinPath("store", id)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		reqURL.String(),
		content,
	)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(api.ErrInvalidRequest, err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	resp, err := c.client.Do(req)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(api.ErrRequestFailed, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return store.Object{}, k6build.NewWrappedError(api.ErrRequestFailed, fmt.Errorf("status %s", resp.Status))
	}
	storeResponse := api.StoreResponse{}
	err = json.NewDecoder(resp.Body).Decode(&storeResponse)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(api.ErrRequestFailed, err)
	}

	if storeResponse.Error != nil {
		return store.Object{}, storeResponse.Error
	}

	return storeResponse.Object, nil
}

// Download returns the content of the object given its url
func (c *StoreClient) Download(ctx context.Context, object store.Object) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, object.URL, nil)
	if err != nil {
		return nil, k6build.NewWrappedError(api.ErrInvalidRequest, err)
	}

	resp, err := c.client.Do(req) //nolint:bodyclose
	if err != nil {
		return nil, k6build.NewWrappedError(api.ErrRequestFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, k6build.NewWrappedError(api.ErrRequestFailed, fmt.Errorf("status %s", resp.Status))
	}

	return resp.Request.Body, nil
}
