package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grafana/k6build"
	"github.com/grafana/k6build/pkg/api"
)

type testSrv struct {
	handlers []requestHandler
}

// process request and return a boolean indicating if request should be passed to the next handler in the chain
type requestHandler func(w http.ResponseWriter, r *http.Request) bool

func withValidateRequest() requestHandler {
	return func(w http.ResponseWriter, r *http.Request) bool {
		req := api.BuildRequest{}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return false
		}

		if req.K6Constrains == "" || req.Platform == "" || len(req.Dependencies) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return false
		}

		return true
	}
}

func withAuthorizationCheck(authType string, auth string) requestHandler {
	return func(w http.ResponseWriter, r *http.Request) bool {
		authHeader := fmt.Sprintf("%s %s", authType, auth)
		if r.Header.Get("Authorization") != authHeader {
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		return true
	}
}

func withHeadersCheck(headers map[string]string) requestHandler {
	return func(w http.ResponseWriter, r *http.Request) bool {
		for h, v := range headers {
			if r.Header.Get(h) != v {
				w.WriteHeader(http.StatusBadRequest)
				return false
			}
		}
		return true
	}
}

func withResponse(status int, response api.BuildResponse) requestHandler {
	return func(w http.ResponseWriter, _ *http.Request) bool {
		resp := &bytes.Buffer{}
		err := json.NewEncoder(resp).Encode(response)
		if err != nil {
			panic("unexpected error encoding response")
		}

		w.WriteHeader(status)
		_, _ = w.Write(resp.Bytes())

		return false
	}
}

func (t testSrv) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	// check headers
	for _, check := range t.handlers {
		if !check(w, r) {
			return
		}
	}

	// by default return ok and an empty artifact
	resp := &bytes.Buffer{}
	_ = json.NewEncoder(resp).Encode(api.BuildResponse{Artifact: k6build.Artifact{}}) //nolint:errchkjson

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp.Bytes())
}

func TestRemote(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		title     string
		headers   map[string]string
		auth      string
		authType  string
		handlers  []requestHandler
		expectErr error
	}{
		{
			title: "normal build",
			handlers: []requestHandler{
				withValidateRequest(),
			},
		},
		{
			title: "build request failed",
			handlers: []requestHandler{
				withResponse(http.StatusOK, api.BuildResponse{Error: k6build.NewWrappedError(api.ErrBuildFailed, nil)}),
			},
			expectErr: api.ErrBuildFailed,
		},
		{
			title:    "auth header",
			auth:     "token",
			authType: "Bearer",
			handlers: []requestHandler{
				withAuthorizationCheck("Bearer", "token"),
			},
			expectErr: nil,
		},
		{
			title:    "with default auth type",
			auth:     "token",
			authType: "",
			handlers: []requestHandler{
				withAuthorizationCheck("Bearer", "token"),
			},
			expectErr: nil,
		},
		{
			title: "failed auth",
			handlers: []requestHandler{
				withResponse(http.StatusUnauthorized, api.BuildResponse{Error: k6build.NewWrappedError(api.ErrRequestFailed, errors.New("unauthorized"))}),
			},
			expectErr: api.ErrRequestFailed,
		},
		{
			title: "custom headers",
			headers: map[string]string{
				"Custom-Header": "Custom-Value",
			},
			handlers: []requestHandler{
				withHeadersCheck(map[string]string{"Custom-Header": "Custom-Value"}),
			},
			expectErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(testSrv{
				handlers: tc.handlers,
			})

			defer srv.Close()

			client, err := NewBuildServiceClient(
				BuildServiceClientConfig{
					URL:               srv.URL,
					Headers:           tc.headers,
					Authorization:     tc.auth,
					AuthorizationType: tc.authType,
				},
			)
			if err != nil {
				t.Fatalf("unexpected %v", err)
			}

			_, err = client.Build(
				context.TODO(),
				"linux/amd64",
				"v0.1.0",
				[]k6build.Dependency{{Name: "k6/x/test", Constraints: "*"}},
			)

			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}
		})
	}
}
