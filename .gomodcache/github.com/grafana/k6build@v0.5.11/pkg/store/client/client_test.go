package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grafana/k6build"
	"github.com/grafana/k6build/pkg/store"
	"github.com/grafana/k6build/pkg/store/api"
)

// returns a HandleFunc that returns a canned status and response
func handlerMock(status int, resp *api.StoreResponse) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		// send canned response
		respBuffer := &bytes.Buffer{}
		if resp != nil {
			err := json.NewEncoder(respBuffer).Encode(resp)
			if err != nil {
				// set uncommon status code to signal something unexpected happened
				w.WriteHeader(http.StatusTeapot)
				return
			}
		}

		w.WriteHeader(status)
		_, _ = w.Write(respBuffer.Bytes())
	}
}

// returns a HandleFunc that returns a canned status and content for a download
func downloadMock(status int, content []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "application/octet-stream")
		w.WriteHeader(status)
		if content != nil {
			_, _ = w.Write(content)
		}
	}
}

func TestStoreClientGet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		title     string
		status    int
		resp      *api.StoreResponse
		expectErr error
	}{
		{
			title:  "normal get",
			status: http.StatusOK,
			resp: &api.StoreResponse{
				Error:  nil,
				Object: store.Object{},
			},
		},
		{
			title:     "object not found",
			status:    http.StatusNotFound,
			resp:      nil,
			expectErr: store.ErrObjectNotFound,
		},
		{
			title:  "error accessing object",
			status: http.StatusInternalServerError,
			resp: &api.StoreResponse{
				Error:  k6build.NewWrappedError(store.ErrAccessingObject, k6build.ErrReasonUnknown),
				Object: store.Object{},
			},
			expectErr: api.ErrRequestFailed,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(handlerMock(tc.status, tc.resp))

			client, err := NewStoreClient(StoreClientConfig{Server: srv.URL})
			if err != nil {
				t.Fatalf("test setup %v", err)
			}

			_, err = client.Get(context.TODO(), "object")
			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}
		})
	}
}

func TestStoreClientPut(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		title     string
		status    int
		resp      *api.StoreResponse
		expectErr error
	}{
		{
			title:  "normal response",
			status: http.StatusOK,
			resp: &api.StoreResponse{
				Error:  nil,
				Object: store.Object{},
			},
		},
		{
			title:  "error creating object",
			status: http.StatusInternalServerError,
			resp: &api.StoreResponse{
				Error:  k6build.NewWrappedError(store.ErrCreatingObject, k6build.ErrReasonUnknown),
				Object: store.Object{},
			},
			expectErr: api.ErrRequestFailed,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(handlerMock(tc.status, tc.resp))

			client, err := NewStoreClient(StoreClientConfig{Server: srv.URL})
			if err != nil {
				t.Fatalf("test setup %v", err)
			}

			_, err = client.Put(context.TODO(), "object", bytes.NewBuffer(nil))
			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}
		})
	}
}

func TestStoreClientDownload(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		title     string
		status    int
		content   []byte
		expectErr error
	}{
		{
			title:   "normal response",
			status:  http.StatusOK,
			content: []byte("object content"),
		},
		{
			title:     "error downloading object",
			status:    http.StatusInternalServerError,
			expectErr: api.ErrRequestFailed,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(downloadMock(tc.status, tc.content))

			client, err := NewStoreClient(StoreClientConfig{Server: srv.URL})
			if err != nil {
				t.Fatalf("test setup %v", err)
			}

			obj := store.Object{
				ID:  "object",
				URL: srv.URL,
			}
			_, err = client.Download(context.TODO(), obj)
			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}
		})
	}
}
