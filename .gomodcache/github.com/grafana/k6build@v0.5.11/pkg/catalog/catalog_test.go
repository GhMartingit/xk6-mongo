package catalog

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const testCatalog = `{
"dep": {"Module": "github.com/dep", "Versions": ["v0.1.0", "v0.2.0"]},
"dep2": {"Module": "github.com/dep2", "Versions": ["v0.1.0"], "Cgo": true}
}`

func TestResolve(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		title     string
		dep       Dependency
		expect    Module
		expectErr error
	}{
		{
			title:  "resolve exact version",
			dep:    Dependency{Name: "dep", Constrains: "v0.1.0"},
			expect: Module{Path: "github.com/dep", Version: "v0.1.0", Cgo: false},
		},
		{
			title:  "resolve > constrain",
			dep:    Dependency{Name: "dep", Constrains: ">v0.1.0"},
			expect: Module{Path: "github.com/dep", Version: "v0.2.0", Cgo: false},
		},
		{
			title:  "resolve latest version",
			dep:    Dependency{Name: "dep", Constrains: "*"},
			expect: Module{Path: "github.com/dep", Version: "v0.2.0", Cgo: false},
		},
		{
			title:  "resolve cgo dependency",
			dep:    Dependency{Name: "dep2", Constrains: "=v0.1.0"},
			expect: Module{Path: "github.com/dep2", Version: "v0.1.0", Cgo: true},
		},
		{
			title:     "unsatisfied > constrain",
			dep:       Dependency{Name: "dep", Constrains: ">v0.2.0"},
			expectErr: ErrCannotSatisfy,
		},
	}

	json := bytes.NewBuffer([]byte(testCatalog))
	catalog, err := NewCatalogFromJSON(json)
	if err != nil {
		t.Fatalf("test setup %v", err)
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			mod, err := catalog.Resolve(context.TODO(), tc.dep)
			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}

			if tc.expectErr == nil && mod != tc.expect {
				t.Fatalf("expected %v got %v", tc.expect, mod)
			}
		})
	}
}

func TestCatalogFromJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		json      string
		expectErr error
	}{
		{
			name:      "load json",
			json:      testCatalog,
			expectErr: nil,
		},
		{
			name:      "empty json",
			json:      "",
			expectErr: ErrInvalidCatalog,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			json := bytes.NewBuffer([]byte(tc.json))
			_, err := NewCatalogFromJSON(json)
			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}
		})
	}
}

func TestCatalogFromURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		handler   http.HandlerFunc
		expectErr error
	}{
		{
			name: "download catalog",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(testCatalog))
			},
			expectErr: nil,
		},
		{
			name: "catalog not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectErr: ErrDownload,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tc.handler)

			_, err := NewCatalogFromURL(context.TODO(), srv.URL)

			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}
		})
	}
}

func TestCatalogFromFile(t *testing.T) {
	t.Parallel()

	catalogFile := filepath.Join(t.TempDir(), "catalog.json")
	err := os.WriteFile(catalogFile, []byte(testCatalog), 0o644)
	if err != nil {
		t.Fatalf("test setup: %v", err)
	}

	testCases := []struct {
		name      string
		file      string
		expectErr error
	}{
		{
			name:      "open catalog",
			file:      catalogFile,
			expectErr: nil,
		},
		{
			name:      "catalog not found",
			file:      "/path/not/found",
			expectErr: ErrOpening,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewCatalogFromFile(tc.file)

			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}
		})
	}
}
