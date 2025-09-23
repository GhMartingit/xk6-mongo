// Package local implements a local build service
package local

import (
	"context"

	"github.com/grafana/k6build"
	"github.com/grafana/k6build/pkg/builder"
	"github.com/grafana/k6build/pkg/store/file"
)

// Opts local builder options
type Opts = builder.Opts

// GoOpts Go build options
type GoOpts = builder.GoOpts

// Config defines the configuration for a Local build service
type Config struct {
	Opts
	// path to catalog's json file. Can be a file path or a URL
	Catalog string
	// path to object store dir
	StoreDir string
}

// NewBuildService creates a local build service using the given configuration
func NewBuildService(ctx context.Context, config Config) (k6build.BuildService, error) {
	store, err := file.NewFileStore(config.StoreDir)
	if err != nil {
		return nil, k6build.NewWrappedError(builder.ErrInitializingBuilder, err)
	}

	return builder.New(ctx, builder.Config{
		Opts:    config.Opts,
		Catalog: config.Catalog,
		Store:   store,
	})
}
