// Package testutils offers utilities for testing against a k6build service
package testutils

import (
	"context"
	"fmt"
	"net/http/httptest"
	"path/filepath"

	"github.com/grafana/k6build/pkg/builder"
	"github.com/grafana/k6build/pkg/catalog"
	"github.com/grafana/k6build/pkg/server"
	"github.com/grafana/k6build/pkg/store/client"
	filestore "github.com/grafana/k6build/pkg/store/file"
	storesrv "github.com/grafana/k6build/pkg/store/server"
)

// TestEnvConfig is the configuration for the test environment
type TestEnvConfig struct {
	// WorkDir is the working directory for the test environment. The object store will be placed there.
	WorkDir string
	// CatalogURL is the URL or path to the extension catalog. If empty, the default catalog will be used
	CatalogURL string
}

// TestEnv is a test environment for the provider tests
type TestEnv struct {
	buildSrv *httptest.Server
	storeSrv *httptest.Server
}

// BuildServiceURL returns the URL of the build service
func (e *TestEnv) BuildServiceURL() string {
	return e.buildSrv.URL
}

// StoreServiceURL returns the URL of the store service
func (e *TestEnv) StoreServiceURL() string {
	return e.storeSrv.URL
}

// Cleanup closes the test environment
func (e *TestEnv) Cleanup() {
	e.buildSrv.Close()
	e.storeSrv.Close()
}

// NewTestEnv creates a new test environment
func NewTestEnv(cfg TestEnvConfig) (*TestEnv, error) {
	// 1. create local file store
	store, err := filestore.NewFileStore(filepath.Join(cfg.WorkDir, "store"))
	if err != nil {
		return nil, fmt.Errorf("store setup %w", err)
	}
	storeConfig := storesrv.StoreServerConfig{
		Store: store,
	}

	// 2. start an object store server
	storeHandler, err := storesrv.NewStoreServer(storeConfig)
	if err != nil {
		return nil, fmt.Errorf("store setup %w", err)
	}
	storeSrv := httptest.NewServer(storeHandler)

	// 3. configure a local builder
	storeClient, err := client.NewStoreClient(client.StoreClientConfig{Server: storeSrv.URL})
	if err != nil {
		return nil, fmt.Errorf("store client setup %w", err)
	}
	catalogURL := cfg.CatalogURL
	if catalogURL == "" {
		catalogURL = catalog.DefaultCatalogURL
	}
	buildConfig := builder.Config{
		Opts: builder.Opts{
			GoOpts: builder.GoOpts{
				CopyGoEnv: true,
			},
		},
		Catalog: catalogURL,
		Store:   storeClient,
	}
	builder, err := builder.New(context.TODO(), buildConfig)
	if err != nil {
		return nil, fmt.Errorf("builder setup %w", err)
	}

	// 5. start a builder server
	srvConfig := server.APIServerConfig{
		BuildService: builder,
	}
	buildSrv := httptest.NewServer(server.NewAPIServer(srvConfig))

	return &TestEnv{
		buildSrv: buildSrv,
		storeSrv: storeSrv,
	}, nil
}
