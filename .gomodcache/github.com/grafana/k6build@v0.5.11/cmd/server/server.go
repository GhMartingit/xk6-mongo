// Package server implements the build server command
package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/grafana/k6build"
	"github.com/grafana/k6build/pkg/builder"
	"github.com/grafana/k6build/pkg/catalog"
	"github.com/grafana/k6build/pkg/httpserver"
	"github.com/grafana/k6build/pkg/server"
	"github.com/grafana/k6build/pkg/store"
	"github.com/grafana/k6build/pkg/store/client"
	"github.com/grafana/k6build/pkg/store/s3"
	"github.com/grafana/k6build/pkg/util"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/spf13/cobra"
)

const (
	long = `
Starts a k6build server

API
---

The server exposes an API for building custom k6 binaries.

The API returns the metadata of the custom binary, including an URL for downloading it,
but does not return the binary itself.

For example

	curl http://localhost:8000/build -d \
	'{
	  "k6":"v0.50.0",
	  "dependencies":[
	    {
		"name":"k6/x/kubernetes",
		"constraints":">v0.8.0"
	    }
	  ],
	  "platform":"linux/amd64"
	}' | jq .

	{
	  "artifact": {
	  "id": "5a241ba6ff643075caadbd06d5a326e5e74f6f10",
	  "url": "http://localhost:9000/store/5a241ba6ff643075caadbd06d5a326e5e74f6f10/download",
	  "dependencies": {
	    "k6": "v0.50.0",
	    "k6/x/kubernetes": "v0.10.0"
	  },
	  "platform": "linux/amd64",
	  "checksum": "bfdf51ec9279e6d7f91df0a342d0c90ab4990ff1fb0215938505a6894edaf913"
	  }
	}

Note: The build server disables CGO by default but enables it when a dependency requires it.
      use --enable-cgo=true to enable CGO support by default.


Metrics
--------

The server exposes prometheus metrics at /metrics

Liveness Probe
--------------

The server exposes a liveness check at /alive
`

	example = `
# start the build server using a custom local catalog
k6build server -c /path/to/catalog.json

# start the build server using a custom GOPROXY
k6build server -e GOPROXY=http://localhost:80

# start the build server with a localstack s3 storage backend
# aws credentials are expected in the default location (e.g. env variables)
export AWS_ACCESS_KEY_ID="test"
export AWS_SECRET_ACCESS_KEY="test"
k6build server --s3-endpoint http://localhost:4566 --store-bucket k6build
`
)

type serverConfig struct {
	allowBuildSemvers bool
	catalogURL        string
	copyGoEnv         bool
	enableCgo         bool
	goEnv             map[string]string
	port              int
	s3Bucket          string
	s3Endpoint        string
	s3Region          string
	storeURL          string
	verbose           bool
	shutdownTimeout   time.Duration
}

// New creates new cobra command for the server command.
func New() *cobra.Command { //nolint:funlen
	var (
		cfg      = serverConfig{}
		logLevel string
	)

	cmd := &cobra.Command{
		Use:     "server",
		Short:   "k6 build service",
		Long:    long,
		Example: example,
		// prevent the usage help to printed to stderr when an error is reported by a subcommand
		SilenceUsage: true,
		// this is needed to prevent cobra to print errors reported by subcommands in the stderr
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			log, err := getLogger(logLevel)
			if err != nil {
				return err
			}

			if cfg.enableCgo {
				log.Warn("CGO is enabled by default. Use --enable-cgo=false to disable it.")
			}

			buildSrv, err := cfg.getBuildService(cmd.Context())
			if err != nil {
				return err
			}

			apiConfig := server.APIServerConfig{
				BuildService: buildSrv,
				Log:          log,
			}
			buildAPI := server.NewAPIServer(apiConfig)

			srvConfig := httpserver.ServerConfig{
				Logger:            log,
				Port:              cfg.port,
				EnableMetrics:     true,
				LivenessProbe:     true,
				ReadHeaderTimeout: 5 * time.Second,
			}

			srv := httpserver.NewServer(srvConfig)
			srv.Handle("/build", buildAPI)

			err = srv.Start(cmd.Context())
			if err != nil {
				return fmt.Errorf("error serving requests %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(
		&cfg.catalogURL,
		"catalog",
		"c",
		catalog.DefaultCatalogURL,
		"dependencies catalog. Can be path to a local file or an URL.",
	)
	cmd.Flags().StringVar(
		&cfg.storeURL,
		"store-url",
		"http://localhost:9000",
		"store server url",
	)
	cmd.Flags().StringVar(&cfg.s3Bucket, "store-bucket", "", "s3 bucket for storing binaries")
	cmd.Flags().StringVar(&cfg.s3Endpoint, "s3-endpoint", "", "s3 endpoint")
	cmd.Flags().StringVar(&cfg.s3Region, "s3-region", "", "aws region")
	cmd.Flags().BoolVarP(&cfg.verbose, "verbose", "v", false, "print build process output")
	cmd.Flags().BoolVarP(&cfg.copyGoEnv, "copy-go-env", "g", true, "copy go environment")
	cmd.Flags().StringToStringVarP(&cfg.goEnv, "env", "e", nil, "build environment variables")
	cmd.Flags().IntVarP(&cfg.port, "port", "p", 8000, "port server will listen")
	cmd.Flags().StringVarP(&logLevel, "log-level", "l", "INFO", "log level")
	cmd.Flags().BoolVar(&cfg.enableCgo, "enable-cgo", false, "enable CGO for building binaries.")
	cmd.Flags().BoolVar(
		&cfg.allowBuildSemvers,
		"allow-build-semvers",
		false,
		"allow building versions with build metadata (e.g v0.0.0+build).",
	)
	cmd.Flags().DurationVar(
		&cfg.shutdownTimeout,
		"shutdown-timeout",
		10*time.Second,
		"maximum time to wait for graceful shutdown",
	)

	return cmd
}

func getLogger(logLevel string) (*slog.Logger, error) {
	ll, err := util.ParseLogLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("parsing log level %w", err)
	}

	return slog.New(
		slog.NewTextHandler(
			os.Stderr,
			&slog.HandlerOptions{
				Level: ll,
			},
		),
	), nil
}

func (cfg serverConfig) getBuildService(ctx context.Context) (k6build.BuildService, error) {
	store, err := cfg.getStore() //nolint:contextcheck
	if err != nil {
		return nil, err
	}

	if cfg.goEnv == nil {
		cfg.goEnv = make(map[string]string)
	}
	cgoEnabled := "0"
	if cfg.enableCgo {
		cgoEnabled = "1"
	}
	cfg.goEnv["CGO_ENABLED"] = cgoEnabled

	config := builder.Config{
		Opts: builder.Opts{
			GoOpts: builder.GoOpts{
				Env:       cfg.goEnv,
				CopyGoEnv: cfg.copyGoEnv,
			},
			Verbose:           cfg.verbose,
			AllowBuildSemvers: cfg.allowBuildSemvers,
		},
		Catalog:    cfg.catalogURL,
		Store:      store,
		Registerer: prometheus.DefaultRegisterer,
	}
	builder, err := builder.New(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("creating local build service  %w", err)
	}

	return builder, nil
}

func (cfg serverConfig) getStore() (store.ObjectStore, error) {
	var (
		err   error
		store store.ObjectStore
	)

	if cfg.s3Bucket != "" {
		store, err = s3.New(s3.Config{
			Bucket:   cfg.s3Bucket,
			Endpoint: cfg.s3Endpoint,
			Region:   cfg.s3Region,
		})
		if err != nil {
			return nil, fmt.Errorf("creating s3 store %w", err)
		}
	} else {
		store, err = client.NewStoreClient(client.StoreClientConfig{
			Server: cfg.storeURL,
		})
		if err != nil {
			return nil, fmt.Errorf("creating store %w", err)
		}
	}

	return store, nil
}
