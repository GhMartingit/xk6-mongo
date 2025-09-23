// Package main is an example of how to use k6provider
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/grafana/k6deps"
	"github.com/grafana/k6provider"
)

// main is an example of how to use k6provider to obtain a k6 binary for an specific k6 version,
// and execute it to check its version.
func main() {
	// get a k6 provider configured with a build service defined in the K6_BUILD_SERVICE_URL
	// environment variable
	provider, err := k6provider.NewDefaultProvider()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(os.Args) == 1 {
		fmt.Println("k6 version must be specified as first argument")
		os.Exit(1)
	}

	k6Version := os.Args[1]
	deps := make(k6deps.Dependencies)
	err = deps.UnmarshalText([]byte(fmt.Sprintf("k6=%s", k6Version)))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// obtain binary from build service
	k6binary, err := provider.GetBinary(context.TODO(), deps)
	for err != nil {
		buildErr, ok := k6provider.AsWrappedError(err)
		if !ok {
			fmt.Println(err)
			break
		}
		fmt.Println(buildErr.Err)
		err = errors.Unwrap(buildErr)
	}

	// execute k6 binary and check version
	cmd := exec.Command(k6binary.Path, "version") //nolint:gosec
	out, err := cmd.Output()
	if err != nil {
		os.Exit(1)
	}

	fmt.Print(string(out))
}
