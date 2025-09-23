// Package local implements the local build command
package local

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/grafana/k6build"
	"github.com/grafana/k6build/pkg/catalog"
	"github.com/grafana/k6build/pkg/local"

	"github.com/spf13/cobra"
)

const (
	long = `
k6build local builder creates a custom k6 binary artifacts that satisfies certain
dependencies. Requires the golang toolchain and git.
`

	example = `
# build k6 v0.51.0 with latest version of k6/x/kubernetes
k6build local -k v0.51.0 -d k6/x/kubernetes

platform: linux/amd64
k6: v0.51.0
k6/x/kubernetes: v0.9.0
checksum: 7f06720503c80153816b4ef9f58571c2fce620e0447fba1bb092188ff87e322d

# build k6 v0.51.0 with k6/x/kubernetes v0.8.0 and k6/x/output-kafka v0.7.0
k6build local -k v0.51.0 \
    -d k6/x/kubernetes:v0.8.0 \
    -d k6/x/output-kafka:v0.7.0

platform: linux/amd64
k6: v0.51.0
k6/x/kubernetes: v0.8.0
k6/x/output-kafka": v0.7.0
checksum: f4af178bb2e29862c0fc7d481076c9ba4468572903480fe9d6c999fea75f3793


# build k6 v0.50.0 with latest version of k6/x/kubernetes using a custom catalog
k6build local -k v0.50.0 -d k6/x/kubernetes \
    -c /path/to/catalog.json -q

# build k6 v0.50.0 using a custom GOPROXY
k6build local -k v0.50.0 -e GOPROXY=http://localhost:80 -q
`
)

// New creates new cobra command for local build command.
func New() *cobra.Command { //nolint:funlen
	var (
		config   local.Config
		deps     []string
		k6       string
		output   string
		platform string
		quiet    bool
	)

	cmd := &cobra.Command{
		Use:     "local",
		Short:   "build custom k6 binary locally",
		Long:    long,
		Example: example,
		// prevent the usage help to printed to stderr when an error is reported by a subcommand
		SilenceUsage: true,
		// this is needed to prevent cobra to print errors reported by subcommands in the stderr
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			srv, err := local.NewBuildService(cmd.Context(), config)
			if err != nil {
				return fmt.Errorf("configuring the build service %w", err)
			}

			buildDeps := []k6build.Dependency{}
			for _, d := range deps {
				name, constrains, _ := strings.Cut(d, ":")
				if constrains == "" {
					constrains = "*"
				}
				buildDeps = append(buildDeps, k6build.Dependency{Name: name, Constraints: constrains})
			}

			artifact, err := srv.Build(cmd.Context(), platform, k6, buildDeps)
			if err != nil {
				return fmt.Errorf("building %w", err)
			}

			if !quiet {
				fmt.Println(artifact.PrintSummary())
			}

			binaryURL, err := url.Parse(artifact.URL)
			if err != nil {
				return fmt.Errorf("malformed URL %w", err)
			}
			artifactBinary, err := os.Open(binaryURL.Path)
			if err != nil {
				return fmt.Errorf("opening output file %w", err)
			}
			defer func() {
				_ = artifactBinary.Close()
			}()

			binary, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE, 0o755) //nolint:gosec
			if err != nil {
				return fmt.Errorf("opening output file %w", err)
			}

			_, err = io.Copy(binary, artifactBinary)
			if err != nil {
				return fmt.Errorf("copying artifact %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&deps, "dependency", "d", nil, "list of dependencies in form package:constrains")
	cmd.Flags().StringVarP(&k6, "k6", "k", "*", "k6 version constrains")
	cmd.Flags().StringVarP(&platform, "platform", "p", "", "target platform (default GOOS/GOARCH)")
	_ = cmd.MarkFlagRequired("platform")
	cmd.Flags().StringVarP(&config.Catalog, "catalog", "c", catalog.DefaultCatalogURL, "dependencies catalog")
	cmd.Flags().StringVarP(&config.StoreDir, "store-dir", "f", "/tmp/k6build/store", "object store dir")
	cmd.Flags().BoolVarP(&config.Opts.Verbose, "verbose", "v", false, "print build process output")
	cmd.Flags().BoolVarP(&config.CopyGoEnv, "copy-go-env", "g", true, "copy go environment")
	cmd.Flags().StringToStringVarP(&config.Opts.Env, "env", "e", nil, "build environment variables")
	cmd.Flags().StringVarP(&output, "output", "o", "k6", "path to put the binary as an executable.")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "don't print artifact's details")
	cmd.Flags().BoolVar(
		&config.AllowBuildSemvers,
		"allow-build-semvers",
		false,
		"allow building versions with build metadata (e.g v0.0.0+build).",
	)
	return cmd
}
