// Package remote implements the client command
package remote

import (
	"fmt"
	"strings"

	"github.com/grafana/k6build"
	"github.com/grafana/k6build/pkg/client"
	"github.com/grafana/k6build/pkg/util"

	"github.com/spf13/cobra"
)

const (
	long = `
Builds custom k6 binaries using a k6build server returning the details of the
binary artifact and optionally download it.
`

	example = `
# build k6 v0.51.0 with k6/x/kubernetes v0.8.0 and k6/x/output-kafka v0.7.0
k6build remote -s http://localhost:8000 \
    -k v0.51.0 \
    -p linux/amd64 \
    -d k6/x/kubernetes:v0.8.0 \
    -d k6/x/output-kafka:v0.7.0

id: 62d08b13fdef171435e2c6874eaad0bb35f2f9c7
platform: linux/amd64
k6: v0.51.0
k6/x/kubernetes: v0.9.0
k6/x/output-kafka": v0.7.0
checksum: f4af178bb2e29862c0fc7d481076c9ba4468572903480fe9d6c999fea75f3793
url: http://localhost:8000/store/62d08b13fdef171435e2c6874eaad0bb35f2f9c7/download


# build k6 v0.51 with k6/x/output-kafka v0.7.0 and download as 'build/k6'
k6build remote -s http://localhost:8000 \
    -p linux/amd64  \
    -k v0.51.0 -d k6/x/output-kafka:v0.7.0 \
    -o build/k6 -q

# check downloaded binary
build/k6 version
k6 v0.51.0 (go1.22.2, linux/amd64)
Extensions:
  github.com/grafana/xk6-output-kafka v0.7.0, xk6-kafka [output]
`
)

// New creates new cobra command for build client command.
func New() *cobra.Command {
	var (
		config   client.BuildServiceClientConfig
		deps     []string
		k6       string
		output   string
		platform string
		quiet    bool
	)

	cmd := &cobra.Command{
		Use:     "remote",
		Short:   "build a custom k6 using a remote build server",
		Long:    long,
		Example: example,
		// prevent the usage help to printed to stderr when an error is reported by a subcommand
		SilenceUsage: true,
		// this is needed to prevent cobra to print errors reported by subcommands in the stderr
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := client.NewBuildServiceClient(config)
			if err != nil {
				return fmt.Errorf("configuring the client %w", err)
			}

			buildDeps := []k6build.Dependency{}
			for _, d := range deps {
				name, constrains, _ := strings.Cut(d, ":")
				if constrains == "" {
					constrains = "*"
				}
				buildDeps = append(buildDeps, k6build.Dependency{Name: name, Constraints: constrains})
			}

			artifact, err := client.Build(cmd.Context(), platform, k6, buildDeps)
			if err != nil {
				return fmt.Errorf("building %w", err)
			}

			if !quiet {
				fmt.Println(artifact.Print())
			}

			if output != "" {
				err = util.Download(cmd.Context(), artifact.URL, output)
				if err != nil {
					return fmt.Errorf("downloading artifact %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&config.URL, "server", "s", "http://localhost:8000", "url for build server")
	cmd.Flags().StringArrayVarP(&deps, "dependency", "d", nil, "list of dependencies in form package:constrains")
	cmd.Flags().StringVarP(&k6, "k6", "k", "*", "k6 version constrains")
	cmd.Flags().StringVarP(&platform, "platform", "p", "", "target platform (default GOOS/GOARCH)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "path to download the custom binary as an executable."+
		"\nIf not specified, the artifact is not downloaded.")
	cmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "don't print artifact's details")

	return cmd
}
