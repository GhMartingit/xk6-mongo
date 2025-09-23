package k6provider

import (
	"context"
	"crypto/rand"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/grafana/k6build/pkg/testutils"
	"github.com/grafana/k6deps"
)

// checks request has the correct Authorization header
func newAuthorizationProxy(buildSrv string, header string, authorization string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(header) != authorization {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		url, _ := url.Parse(buildSrv)
		httputil.NewSingleHostReverseProxy(url).ServeHTTP(w, r)
	}
}

// Pass through requests
func newTransparentProxy(upstream string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, _ := url.Parse(upstream)
		httputil.NewSingleHostReverseProxy(url).ServeHTTP(w, r)
	}
}

// fail with the given error up to a number of times
func newUnreliableProxy(upstream string, status int, failures int) http.HandlerFunc {
	requests := 0
	return func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests <= failures {
			w.WriteHeader(status)
			return
		}

		url, _ := url.Parse(upstream)
		httputil.NewSingleHostReverseProxy(url).ServeHTTP(w, r)
	}
}

// returns a corrupted random content
func newCorruptedProxy() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		buffer := make([]byte, 1024)
		_, _ = rand.Read(buffer)
		_, _ = w.Write(buffer)
	}
}

func Test_Provider(t *testing.T) { //nolint:tparallel
	testEnv, err := testutils.NewTestEnv(
		testutils.TestEnvConfig{
			WorkDir:    t.TempDir(),
			CatalogURL: "testdata/catalog.json",
		},
	)
	if err != nil {
		t.Fatalf("test env setup %v", err)
	}
	t.Cleanup(testEnv.Cleanup)

	// reuse the same dependencies for all tests to avoid multiple builds
	deps := k6deps.Dependencies{}
	err = deps.UnmarshalText([]byte("k6=v0.50.0"))
	if err != nil {
		t.Fatalf("analyzing dependencies %v", err)
	}

	t.Run("test binary provisioning", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			title         string
			buildProxy    http.HandlerFunc
			downloadProxy http.HandlerFunc
			config        Config
			expectErr     error
			expect        string
		}{
			{
				title: "test authentication using bearer token",
				config: Config{
					BuildServiceAuth: "token",
				},
				buildProxy: newAuthorizationProxy(testEnv.BuildServiceURL(), "Authorization", "Bearer token"),
				expectErr:  nil,
			},
			{
				title: "test authentication using custom header",
				config: Config{
					BuildServiceHeaders: map[string]string{
						"Custom-Auth": "token",
					},
				},
				buildProxy: newAuthorizationProxy(testEnv.BuildServiceURL(), "Custom-Auth", "token"),
				expectErr:  nil,
			},
			{
				title: "test authentication failed (missing bearer token)",
				config: Config{
					BuildServiceAuth: "",
				},
				buildProxy: newAuthorizationProxy(testEnv.BuildServiceURL(), "Authorization", "Bearer token"),
				expectErr:  ErrBuild,
			},
			{
				title:         "test download using proxy",
				downloadProxy: newTransparentProxy(testEnv.StoreServiceURL()),
			},
			{
				title: "test download proxy unavailable",
				config: Config{
					DownloadConfig: DownloadConfig{
						ProxyURL: "http://127.0.0.1:12345",
					},
				},
				expectErr: ErrDownload,
			},
			{
				title: "test download authentication using bearer token",
				config: Config{
					BuildServiceAuth: "token",
					DownloadConfig: DownloadConfig{
						Authorization: "token",
					},
				},
				downloadProxy: newAuthorizationProxy(testEnv.StoreServiceURL(), "Authorization", "Bearer token"),
				expectErr:     nil,
			},
			{
				title: "test download authentication failed (missing bearer token)",
				config: Config{
					DownloadConfig: DownloadConfig{
						Authorization: "",
					},
				},
				downloadProxy: newAuthorizationProxy(testEnv.StoreServiceURL(), "Authorization", "Bearer token"),
				expectErr:     ErrDownload,
			},
			{
				title:         "test download with default retries",
				downloadProxy: newUnreliableProxy(testEnv.StoreServiceURL(), http.StatusServiceUnavailable, 1),
			},
			{
				title:         "test we don't retry forever",
				config:        Config{DownloadConfig: DownloadConfig{Retries: 1}},
				downloadProxy: newUnreliableProxy(testEnv.StoreServiceURL(), http.StatusServiceUnavailable, math.MaxInt),
				expectErr:     ErrDownload,
			},
			{
				title:         "detect corrupted binary",
				downloadProxy: newCorruptedProxy(),
				expectErr:     ErrDownload,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				t.Parallel()

				// by default, we use the build service, but if there's a
				// proxy defined, we use it
				testSrvURL := testEnv.BuildServiceURL()
				if tc.buildProxy != nil {
					testSrv := httptest.NewServer(tc.buildProxy)
					defer testSrv.Close()
					testSrvURL = testSrv.URL
				}

				// if there's a download proxy, we use it
				testStoreProxy := ""
				if tc.downloadProxy != nil {
					downloadProxy := httptest.NewServer(tc.downloadProxy)
					defer downloadProxy.Close()
					testStoreProxy = downloadProxy.URL
				}

				config := tc.config
				config.BinaryCacheDir = filepath.Join(t.TempDir(), "provider")
				config.BuildServiceURL = testSrvURL
				// FIXME: override download proxy if not set in the test. This is needed to test wrong proxy URL
				if config.DownloadConfig.ProxyURL == "" {
					config.DownloadConfig.ProxyURL = testStoreProxy
				}

				provider, err := NewProvider(config)
				if err != nil {
					t.Fatalf("initializing provider %v", err)
				}

				k6, err := provider.GetBinary(context.TODO(), deps)
				if !errors.Is(err, tc.expectErr) {
					t.Fatalf("expected %v got %v", tc.expectErr, err)
				}

				// in case of error the binary should not be downloaded
				if tc.expectErr != nil {
					_, err := os.Stat(k6.Path)
					if !os.IsNotExist(err) {
						t.Fatalf("expected binary not to be downloaded")
					}
					return
				}

				// in case of not error, we expect the binary to be downloaded
				cmd := exec.Command(k6.Path, "version")

				out, err := cmd.Output()
				if err != nil {
					t.Fatalf("running command %v", err)
				}

				t.Log(string(out))
			})
		}
	})

	t.Run("test concurrent downloads", func(t *testing.T) {
		t.Parallel()

		provider, err := NewProvider(Config{
			BinaryCacheDir:  filepath.Join(t.TempDir(), "provider"),
			BuildServiceURL: testEnv.BuildServiceURL(),
		})
		if err != nil {
			t.Fatalf("initializing provider %v", err)
		}

		wg := sync.WaitGroup{}
		errs := make(chan error, 10)

		for range 3 {
			wg.Add(1)
			go func() {
				defer wg.Done()

				k6, err := provider.GetBinary(context.TODO(), deps)
				if err != nil {
					errs <- err
					return
				}
				cmd := exec.Command(k6.Path, "version")

				err = cmd.Run()
				if err != nil {
					errs <- err
				}
			}()
		}

		wg.Wait()

		select {
		case err := <-errs:
			t.Fatalf("expected no error, got %v", err)
		default:
		}
	})

	t.Run("test NerProvider", func(t *testing.T) {
		cacheDir := filepath.Join(t.TempDir(), "k6provider")

		testCases := []struct {
			title     string
			env       map[string]string
			config    Config
			expectErr error
		}{
			{
				title: "new from Config",
				env:   map[string]string{},
				config: Config{
					BinaryCacheDir:  cacheDir,
					BuildServiceURL: testEnv.BuildServiceURL(),
				},
			},
			{
				title: "use deprecated BinDir",
				env:   map[string]string{},
				config: Config{
					BinDir:          cacheDir,
					BuildServiceURL: testEnv.BuildServiceURL(),
				},
			},
			{
				title: "new from env",
				env: map[string]string{
					"K6_BINARY_CACHE":      cacheDir,
					"K6_BUILD_SERVICE_URL": testEnv.BuildServiceURL(),
				},
				config: Config{},
			},
			{
				title: "config should override env",
				env: map[string]string{
					"K6_BINARY_CACHE":      "/path/to/wrong/cache/dir",
					"K6_BUILD_SERVICE_URL": "http://localhost:9999",
				},
				config: Config{
					BinaryCacheDir:  cacheDir,
					BuildServiceURL: testEnv.BuildServiceURL(),
				},
			},
			{
				title: "new from empty Config",
				env: map[string]string{
					"K6_BINARY_CACHE":      "",
					"K6_BUILD_SERVICE_URL": "",
				},
				config:    Config{},
				expectErr: ErrConfig,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				for k, v := range tc.env {
					t.Setenv(k, v)
				}

				// cleanup cache dir to avoid cached binaries
				err = os.RemoveAll(cacheDir)
				if err != nil {
					t.Fatalf("cleaning up cache dir %v", err)
				}

				provider, err := NewProvider(tc.config)
				if !errors.Is(err, tc.expectErr) {
					t.Fatalf("expected %v got %v", tc.expectErr, err)
				}

				if tc.expectErr != nil {
					return
				}

				if provider == nil {
					t.Fatal("expected provider to be initialized")
				}

				binary, err := provider.GetBinary(context.TODO(), k6deps.Dependencies{})
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}

				if !strings.HasPrefix(binary.Path, cacheDir) {
					t.Fatalf("expected binary path to be in %s, got %s", cacheDir, binary.Path)
				}

				if !strings.HasPrefix(binary.DownloadURL, testEnv.StoreServiceURL()) {
					t.Fatalf("expected download url to be in %s, got %s", testEnv.StoreServiceURL(), binary.DownloadURL)
				}
			})
		}
	})
}
