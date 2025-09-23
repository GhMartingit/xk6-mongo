// Package catalog defines the extension catalog
//
// A catalog maps a dependency for k6 extension with optional semantic versioning
// constrains to the corresponding golang modules.
//
// For example `k6/x/output-kafka:>0.1.0` ==> `github.com/grafana/xk6-output-kafka@v0.2.0`
//
// The catalog is a json file with the following schema:
//
//		{
//		     "<dependency>": {
//	              "module": "<module path>",
//	              "versions": ["<version>", "<version>", ... "<version>"],
//	              "cgo": <bool>
//		     },
//		}
//
// where:
// <dependency>: is the import path for the dependency
// module: is the path to the go module that implements the dependency
// versions: is the list of supported versions
// cgo: is a boolean that indicates if the module requires cgo
//
// Example:
//
//	{
//	     "k6": {"module": "go.k6.io/k6", "versions": ["v0.50.0", "v0.51.0"]},
//	     "k6/x/kubernetes": {"module": "github.com/grafana/xk6-kubernetes", "versions": ["v0.8.0","v0.9.0"]},
//	     "k6/x/output-kafka": {"module": "github.com/grafana/xk6-output-kafka", "versions": ["v0.7.0"]},
//	     "k6/x/xk6-sql-driver-sqlite3": {"module": "github.com/grafana/xk6-sql", "cgo": true, "versions": ["v0.1.0"]}
//	}
package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

const (
	DefaultCatalogFile = "catalog.json"                        //nolint:revive
	DefaultCatalogURL  = "https://registry.k6.io/catalog.json" //nolint:revive
)

var (
	ErrCannotSatisfy     = errors.New("cannot satisfy dependency") //nolint:revive
	ErrDownload          = errors.New("downloading catalog")
	ErrInvalidConstrain  = errors.New("invalid constrain")
	ErrInvalidCatalog    = fmt.Errorf("invalid catalog")
	ErrOpening           = errors.New("opening catalog")
	ErrUnknownDependency = errors.New("unknown dependency")
)

// Dependency defines a Dependency with a version constrain
// Examples:
// Name: k6/x/k6-kubernetes   Constrains *
// Name: k6/x/k6-output-kafka Constrains >v0.9.0
type Dependency struct {
	Name       string `json:"name,omitempty"`
	Constrains string `json:"constrains,omitempty"`
}

// Module defines a go module that resolves a Dependency
type Module struct {
	Path    string `json:"path,omitempty"`
	Version string `json:"version,omitempty"`
	Cgo     bool   `json:"cgo,omitempty"`
}

// Catalog defines the interface of the extension catalog service
type Catalog interface {
	// Resolve returns a Module that satisfies a Dependency
	Resolve(ctx context.Context, dep Dependency) (Module, error)
}

// entry defines a catalog entry
type entry struct {
	Module   string   `json:"module,omitempty"`
	Versions []string `json:"versions,omitempty"`
	Cgo      bool     `json:"cgo,omitempty"`
}

type catalog struct {
	dependencies map[string]entry
}

// getVersions returns the versions for a given module
func (c catalog) getVersions(_ context.Context, mod string) (entry, error) {
	e, found := c.dependencies[mod]
	if !found {
		return entry{}, fmt.Errorf("%w : %s", ErrUnknownDependency, mod)
	}

	return e, nil
}

// NewCatalogFromJSON creates a Catalog from a json file that follows the [schema](./schema.json):
func NewCatalogFromJSON(stream io.Reader) (Catalog, error) {
	buff := &bytes.Buffer{}
	_, err := buff.ReadFrom(stream)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCatalog, err)
	}

	dependencies := map[string]entry{}
	err = json.Unmarshal(buff.Bytes(), &dependencies)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCatalog, err)
	}

	return catalog{
		dependencies: dependencies,
	}, nil
}

// NewCatalog returns a catalog loaded from a location.
// The location can be a local path or an URL
func NewCatalog(ctx context.Context, location string) (Catalog, error) {
	if strings.HasPrefix(location, "http") {
		return NewCatalogFromURL(ctx, location)
	}

	return NewCatalogFromFile(location)
}

// NewCatalogFromFile creates a Catalog from a json file
func NewCatalogFromFile(catalogFile string) (Catalog, error) {
	json, err := os.ReadFile(catalogFile) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpening, err)
	}

	buff := bytes.NewBuffer(json)
	return NewCatalogFromJSON(buff)
}

// NewCatalogFromURL creates a Catalog from a URL
func NewCatalogFromURL(ctx context.Context, catalogURL string) (Catalog, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, catalogURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %w", ErrDownload, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %w", ErrDownload, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w %s", ErrDownload, resp.Status)
	}

	catalog, err := NewCatalogFromJSON(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %w", ErrDownload, err)
	}

	return catalog, nil
}

// DefaultCatalog creates a Catalog from the default catalog URL
func DefaultCatalog() (Catalog, error) {
	return NewCatalogFromURL(context.TODO(), DefaultCatalogURL)
}

func (c catalog) Resolve(ctx context.Context, dep Dependency) (Module, error) {
	entry, err := c.getVersions(ctx, dep.Name)
	if err != nil {
		return Module{}, err
	}

	constrain, err := semver.NewConstraint(dep.Constrains)
	if err != nil {
		return Module{}, fmt.Errorf("%w : %s", ErrInvalidConstrain, dep.Constrains)
	}

	versions := []*semver.Version{}
	for _, v := range entry.Versions {
		version, err := semver.NewVersion(v)
		if err != nil {
			return Module{}, err
		}
		versions = append(versions, version)
	}

	if len(versions) > 0 {
		// try to find the higher version that satisfies the condition
		sort.Sort(sort.Reverse(semver.Collection(versions)))
		for _, v := range versions {
			if constrain.Check(v) {
				return Module{Path: entry.Module, Version: v.Original(), Cgo: entry.Cgo}, nil
			}
		}
	}

	return Module{}, fmt.Errorf("%w : %s %s", ErrCannotSatisfy, dep.Name, dep.Constrains)
}
