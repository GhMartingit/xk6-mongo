package builder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/grafana/k6build"
	"github.com/grafana/k6build/pkg/catalog"
	"github.com/grafana/k6build/pkg/store/file"
	"github.com/grafana/k6foundry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// DependencyComp compares two dependencies for ordering
func DependencyComp(a, b catalog.Module) bool { return a.Path < b.Path }

type mockFoundry struct {
	opts k6foundry.NativeFoundryOpts
}

// Mocks the Faundry's Build method
// Returns the build info for the given platform, k6 version and modules
func (m *mockFoundry) Build(
	_ context.Context,
	platform k6foundry.Platform,
	k6Version string,
	mods []k6foundry.Module,
	reps []k6foundry.Module,
	buildOpts []string,
	out io.Writer,
) (*k6foundry.BuildInfo, error) {
	modVersions := make(map[string]string)
	for _, mod := range mods {
		modVersions[mod.Path] = mod.Version
	}
	return &k6foundry.BuildInfo{
		Platform:    platform.String(),
		ModVersions: modVersions,
	}, nil
}

func MockFoundryFactory(_ context.Context, opts k6foundry.NativeFoundryOpts) (k6foundry.Foundry, error) {
	return &mockFoundry{
		opts: opts,
	}, nil
}

// SetupTestBuilder setups a local build service for testing
func SetupTestBuilder(t *testing.T) (*Builder, error) {
	store, err := file.NewFileStore(t.TempDir())
	if err != nil {
		return nil, fmt.Errorf("creating temporary object store %w", err)
	}

	return New(context.Background(), Config{
		Opts:    Opts{},
		Catalog: filepath.Join("testdata", "catalog.json"),
		Store:   store,
		Foundry: FoundryFactoryFunction(MockFoundryFactory),
	})
}

func TestDependencyResolution(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		title     string
		k6        string
		deps      []k6build.Dependency
		expectErr error
		expect    k6build.Artifact
	}{
		{
			title:     "build k6 v0.1.0 ",
			k6:        "v0.1.0",
			deps:      []k6build.Dependency{},
			expectErr: nil,
			expect: k6build.Artifact{
				Dependencies: map[string]string{"k6": "v0.1.0"},
			},
		},
		{
			title:     "build k6 >v0.1.0",
			k6:        ">v0.1.0",
			deps:      []k6build.Dependency{},
			expectErr: nil,
			expect: k6build.Artifact{
				Dependencies: map[string]string{"k6": "v0.2.0"},
			},
		},
		{
			title:     "build unsatisfied k6 constrain (>v0.2.0)",
			k6:        ">v0.2.0",
			deps:      []k6build.Dependency{},
			expectErr: catalog.ErrCannotSatisfy,
		},
		{
			title:     "build k6 v0.1.0 exact dependency constraint",
			k6:        "v0.1.0",
			deps:      []k6build.Dependency{{Name: "k6/x/ext", Constraints: "v0.1.0"}},
			expectErr: nil,
			expect: k6build.Artifact{
				Dependencies: map[string]string{
					"k6":       "v0.1.0",
					"k6/x/ext": "v0.1.0",
				},
			},
		},
		{
			title:     "build k6 v0.1.0 unsatisfied dependency constrain",
			k6:        "v0.1.0",
			deps:      []k6build.Dependency{{Name: "k6/x/ext", Constraints: ">v0.2.0"}},
			expectErr: catalog.ErrCannotSatisfy,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			buildsrv, err := SetupTestBuilder(t)
			if err != nil {
				t.Fatalf("test setup %v", err)
			}

			artifact, err := buildsrv.Build(
				context.TODO(),
				"linux/amd64",
				tc.k6,
				tc.deps,
			)

			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("unexpected error wanted %v got %v", tc.expectErr, err)
			}

			// don't check artifact if error is expected
			if tc.expectErr != nil {
				return
			}

			diff := cmp.Diff(tc.expect.Dependencies, artifact.Dependencies, cmpopts.SortSlices(DependencyComp))
			if diff != "" {
				t.Fatalf("dependencies don't match: %s\n", diff)
			}
		})
	}
}

func TestIdempotentBuild(t *testing.T) {
	t.Parallel()
	buildsrv, err := SetupTestBuilder(t)
	if err != nil {
		t.Fatalf("test setup %v", err)
	}

	artifact, err := buildsrv.Build(
		context.TODO(),
		"linux/amd64",
		"v0.1.0",
		[]k6build.Dependency{
			{Name: "k6/x/ext", Constraints: "v0.1.0"},
			{Name: "k6/x/ext2", Constraints: "v0.1.0"},
		},
	)
	if err != nil {
		t.Fatalf("test setup %v", err)
	}

	t.Run("should rebuild same artifact", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			title    string
			platform string
			k6       string
			deps     []k6build.Dependency
		}{
			{
				title:    "same dependencies",
				platform: "linux/amd64",
				k6:       "v0.1.0",
				deps: []k6build.Dependency{
					{Name: "k6/x/ext", Constraints: "v0.1.0"},
					{Name: "k6/x/ext2", Constraints: "v0.1.0"},
				},
			},
			{
				title:    "different order of dependencies",
				platform: "linux/amd64",
				k6:       "v0.1.0",
				deps: []k6build.Dependency{
					{Name: "k6/x/ext2", Constraints: "v0.1.0"},
					{Name: "k6/x/ext", Constraints: "v0.1.0"},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				t.Parallel()

				rebuild, err := buildsrv.Build(
					context.TODO(),
					tc.platform,
					tc.k6,
					tc.deps,
				)
				if err != nil {
					t.Fatalf("unexpected %v", err)
				}

				if artifact.ID != rebuild.ID {
					t.Fatalf("artifact ID don't match")
				}

				diff := cmp.Diff(artifact.Dependencies, rebuild.Dependencies, cmpopts.SortSlices(DependencyComp))
				if diff != "" {
					t.Fatalf("dependencies don't match: %s\n", diff)
				}
			})
		}
	})

	t.Run("should build a different artifact", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			title    string
			platform string
			k6       string
			deps     []k6build.Dependency
		}{
			{
				title:    "different k6 versions",
				platform: "linux/amd64",
				k6:       "v0.2.0",
				deps: []k6build.Dependency{
					{Name: "k6/x/ext", Constraints: "v0.1.0"},
					{Name: "k6/x/ext2", Constraints: "v0.1.0"},
				},
			},
			{
				title:    "different dependency versions",
				platform: "linux/amd64",
				k6:       "v0.1.0",
				deps: []k6build.Dependency{
					{Name: "k6/x/ext", Constraints: "v0.2.0"},
					{Name: "k6/x/ext2", Constraints: "v0.1.0"},
				},
			},
			{
				title:    "different dependencies",
				platform: "linux/amd64",
				k6:       "v0.1.0",
				deps: []k6build.Dependency{
					{Name: "k6/x/ext", Constraints: "v0.1.0"},
				},
			},
			{
				title:    "different platform",
				platform: "linux/arm64",
				k6:       "v0.1.0",
				deps: []k6build.Dependency{
					{Name: "k6/x/ext", Constraints: "v0.1.0"},
					{Name: "k6/x/ext2", Constraints: "v0.1.0"},
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				t.Parallel()

				rebuild, err := buildsrv.Build(
					context.TODO(),
					tc.platform,
					tc.k6,
					tc.deps,
				)
				if err != nil {
					t.Fatalf("unexpected %v", err)
				}

				if artifact.ID == rebuild.ID {
					t.Fatalf("should had built a different artifact")
				}
			})
		}
	})
}

// TestConcurrentBuilds tests that is safe to build the same artifact concurrently and that
// concurrent builds of different artifacts are not affected.
// The test uses a local test setup backed by a file object store.
// Attempting to write the same artifact twice will return an error.
func TestConcurrentBuilds(t *testing.T) {
	t.Parallel()
	buildsrv, err := SetupTestBuilder(t)
	if err != nil {
		t.Fatalf("test setup %v", err)
	}

	builds := []struct {
		k6Ver string
		deps  []k6build.Dependency
	}{
		{
			k6Ver: "v0.1.0",
			deps: []k6build.Dependency{
				{Name: "k6/x/ext", Constraints: "v0.1.0"},
			},
		},
		{
			k6Ver: "v0.1.0",
			deps: []k6build.Dependency{
				{Name: "k6/x/ext", Constraints: "v0.1.0"},
			},
		},
		{
			k6Ver: "v0.2.0",
			deps: []k6build.Dependency{
				{Name: "k6/x/ext", Constraints: "v0.1.0"},
			},
		},
	}

	errch := make(chan error, len(builds))

	wg := sync.WaitGroup{}
	for _, b := range builds {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if _, err := buildsrv.Build(
				context.TODO(),
				"linux/amd64",
				b.k6Ver,
				b.deps,
			); err != nil {
				errch <- err
			}
		}()
	}

	wg.Wait()

	select {
	case err := <-errch:
		t.Fatalf("unexpected %v", err)
	default:
	}
}

// templates for producing metric text output
var metricTemplates = map[string]string{
	"k6build_requests_total": `
# HELP k6build_requests_total The total number of builds requests
# TYPE k6build_requests_total counter
k6build_requests_total %s`,
	"k6build_builds_total": `
# HELP k6build_builds_total The total number of builds
# HELP k6build_builds_total
# TYPE k6build_builds_total counter
k6build_builds_total %s`,
	"k6build_builds_failed_total": `
# HELP k6build_builds_failed_total The total number of failed builds
# TYPE k6build_builds_failed_total counter
k6build_builds_failed_total %s`,
	"k6build_builds_invalid_total": `
# HELP k6build_builds_invalid_total The total number of builds with invalid parameters
# TYPE k6build_builds_invalid_total counter
k6build_builds_invalid_total %s`,
}

func TestMetrics(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		title    string
		requests []string
		expected map[string]string
	}{
		{
			title:    "single build",
			requests: []string{"v0.2.0"},
			expected: map[string]string{
				"k6build_requests_total":       "1",
				"k6build_builds_total":         "1",
				"k6build_builds_invalid_total": "0",
				"k6build_builds_failed_total":  "0",
			},
		},
		{
			title:    "unsatisfied build",
			requests: []string{"v0.3.0"},
			expected: map[string]string{
				"k6build_requests_total":       "1",
				"k6build_builds_total":         "0",
				"k6build_builds_invalid_total": "1",
				"k6build_builds_failed_total":  "0",
			},
		},
		{
			title:    "multiple builds same versions",
			requests: []string{"v0.2.0", "v0.2.0"},
			expected: map[string]string{
				"k6build_requests_total":       "2",
				"k6build_builds_total":         "1",
				"k6build_builds_invalid_total": "0",
				"k6build_builds_failed_total":  "0",
			},
		},
		{
			title:    "multiple builds different versions",
			requests: []string{"v0.2.0", "v0.1.0"},
			expected: map[string]string{
				"k6build_requests_total":       "2",
				"k6build_builds_total":         "2",
				"k6build_builds_invalid_total": "0",
				"k6build_builds_failed_total":  "0",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			register := prometheus.NewPedanticRegistry()

			store, err := file.NewFileStore(t.TempDir())
			if err != nil {
				t.Fatalf("creating temporary object store %v", err)
			}

			builder, err := New(context.Background(), Config{
				Opts:       Opts{},
				Catalog:    filepath.Join("testdata", "catalog.json"),
				Store:      store,
				Foundry:    FoundryFactoryFunction(MockFoundryFactory),
				Registerer: register,
			})
			if err != nil {
				t.Fatalf("creating builder %v", err)
			}

			for _, k6 := range tc.requests {
				_, err = builder.Build(
					context.TODO(),
					"linux/amd64",
					k6,
					[]k6build.Dependency{},
				)
				// ignore unsatisfied builds as they are expected
				if err != nil && !errors.Is(err, ErrInvalidParameters) {
					t.Fatalf("unexpected %v", err)
				}
			}

			// build the prometheus text output for the expected metrics
			metrics := []string{}
			text := strings.Builder{}
			for metric, expected := range tc.expected {
				metrics = append(metrics, metric)
				text.Write([]byte(fmt.Sprintf(metricTemplates[metric], expected)))
			}
			text.Write([]byte("\n"))

			err = testutil.CollectAndCompare(register, strings.NewReader(text.String()), metrics...)
			if err != nil {
				t.Fatalf("unexpected %v", err)
			}
		})
	}
}
