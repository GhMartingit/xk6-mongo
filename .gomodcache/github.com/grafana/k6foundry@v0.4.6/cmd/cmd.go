// Package cmd implements the k6foundry command
// nolint:forbidigo,funlen,nolintlint
package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/grafana/k6foundry"
	"github.com/grafana/k6foundry/pkg/util"

	"github.com/spf13/cobra"
)

var ErrTargetPlatformUndefined = errors.New("target platform is required") //nolint:revive

const long = `
builds a custom k6 binary with extensions.

The extensions are specified using the go module format: path[@version][replace@version]

The module's path must follow go conventions (e.g. github.com/my-module)
If version is omitted, 'latest' is used.
The replace path can be a mod path or a relative path (e.g. ../my-module).
If a relative replacement path is specified, the replacement version cannot be specified.
`

const example = `
# build k6 v0.50.0 with latest version of xk6-kubernetes
k6foundry build -v v0.50.0 -d github.com/grafana/xk6-kubernetes

# build k6 v0.49.0 with xk6-kubernetes v0.9.0 and k6-output-kafka v0.7.0
k6foundry build -v v0.49.0 -d github.com/grafana/xk6-kubernetes \
    -d github.com/grafana/xk6-output-kafka@v0.7.0

# build latest version of k6 with latest version of xk6-kubernetes v0.8.0
k6foundry build -d github.com/grafana/xk6-kubernetes@v0.8.0

# build latest version of k6 and replace xk6-kubernetes with a local module
k6foundry build -d github.com/grafana/xk6-kubernetes=../xk6-kubernetes

# build k6 from a local repository
k6foundry build -r ../k6

# build k6 using a custom GOPROXY and force all modules from the proxy
k6foundry build -e GOPROXY=http://localhost:8000 -e GONOPROXY=none

# build k6 using a temporary go cache ignoring go mod cache and go cache
k6foundry build --tmp-cache=true
`

// New creates new cobra command for build command.
func New() *cobra.Command {
	var (
		opts         k6foundry.NativeFoundryOpts
		deps         []string
		k6Version    string
		k6Repo       string
		platformFlag string
		outPath      string
		buildOpts    []string
		verbose      bool
		logLevelText string
		listVersions bool
	)

	cmd := &cobra.Command{
		Use:     "build",
		Short:   "build a custom k6 binary with extensions",
		Long:    long,
		Example: example,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			var err error
			platform := k6foundry.RuntimePlatform()
			if platformFlag != "" {
				platform, err = k6foundry.ParsePlatform(platformFlag)
				if err != nil {
					return err
				}
			}

			mods := []k6foundry.Module{}
			for _, d := range deps {
				mod, err2 := k6foundry.ParseModule(d)
				if err2 != nil {
					return err2
				}
				mods = append(mods, mod)
			}

			// set build output
			if verbose {
				opts.Stdout = os.Stdout
				opts.Stderr = os.Stderr
			}

			// set log
			logLevel, err := util.ParseLogLevel(logLevelText)
			if err != nil {
				return fmt.Errorf("parsing log level %w", err)
			}

			log := slog.New(
				slog.NewTextHandler(
					os.Stderr,
					&slog.HandlerOptions{
						Level: logLevel,
					},
				),
			)

			opts.Logger = log
			opts.K6Repo = k6Repo

			b, err := k6foundry.NewNativeFoundry(ctx, opts)
			if err != nil {
				return err
			}

			// TODO: check file permissions
			outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, 0o777) //nolint:gosec
			if err != nil {
				return err
			}

			defer outFile.Close() //nolint:errcheck
			buildInfo, err := b.Build(
				ctx,
				platform,
				k6Version,
				mods,
				[]k6foundry.Module{},
				buildOpts,
				outFile,
			)
			if err != nil {
				return err
			}

			if listVersions {
				for m, v := range buildInfo.ModVersions {
					fmt.Printf("%s: %s\n", m, v)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(
		&deps,
		"dependency",
		"d",
		[]string{},
		"list of dependencies using go mod format: path[@version][replace@version]",
	)
	cmd.Flags().StringVarP(&k6Version, "k6-version", "v", "latest", "k6 version")
	cmd.Flags().StringVarP(&k6Repo, "k6-repository", "r", "", "k6 repository")
	cmd.Flags().StringVarP(&platformFlag, "platform", "p", "", "target platform in the format os/arch")
	cmd.Flags().StringVarP(&outPath, "output", "o", "k6", "path to output file")
	cmd.Flags().BoolVar(&opts.CopyGoEnv, "copy-go-env", true, "copy current go environment")
	cmd.Flags().StringVar(&logLevelText, "log-level", "INFO", "log level")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "verbose build output")
	cmd.Flags().StringArrayVarP(&buildOpts, "build-opts", "b", []string{}, "go build opts. e.g. -ldflags='-w -s'")
	cmd.Flags().StringToStringVarP(&opts.Env, "env", "e", nil, "build environment variables")
	cmd.Flags().BoolVarP(&opts.TmpCache, "tmp-cache", "t", false, "use a temporary go cache."+
		"Forces downloading all dependencies.")
	cmd.Flags().BoolVar(&listVersions, "list-versions", false, "list built versions")

	return cmd
}
