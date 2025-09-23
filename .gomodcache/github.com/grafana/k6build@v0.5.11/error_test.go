package k6build

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func Test_WrappedError(t *testing.T) {
	t.Parallel()

	var (
		err    = errors.New("error")
		reason = errors.New("reason")
	)

	testCases := []struct {
		title  string
		err    error
		reason error
		expect []error
	}{
		{
			title:  "error and reason",
			err:    err,
			reason: reason,
			expect: []error{err, reason},
		},
		{
			title:  "error not reason",
			err:    err,
			reason: nil,
			expect: []error{err},
		},
		{
			title:  "multiple and reasons",
			err:    err,
			reason: reason,
			expect: []error{err, reason},
		},
		{
			title:  "wrapped err",
			err:    fmt.Errorf("wrapped %w", err),
			reason: reason,
			expect: []error{err, reason},
		},
		{
			title:  "wrapped reason",
			err:    err,
			reason: fmt.Errorf("wrapped %w", reason),
			expect: []error{err, reason},
		},
		{
			title:  "wrapped err in target",
			err:    err,
			reason: reason,
			expect: []error{fmt.Errorf("wrapped %w", err)},
		},
		{
			title:  "wrapped reason in target",
			err:    err,
			reason: reason,
			expect: []error{fmt.Errorf("wrapped %w", reason)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			err := NewWrappedError(tc.err, tc.reason)
			for _, expected := range tc.expect {
				if !errors.Is(err, expected) {
					t.Fatalf("expected %v got %v", expected, err)
				}
			}
		})
	}
}

func Test_JsonSerialization(t *testing.T) {
	t.Parallel()

	var (
		err    = errors.New("error")
		reason = errors.New("reason")
		root   = errors.New("root")
	)

	testCases := []struct {
		title  string
		err    *WrappedError
		expect []byte
	}{
		{
			title:  "error with cause",
			err:    NewWrappedError(err, reason),
			expect: []byte(`{"error":"error","reason":{"error":"reason"}}`),
		},
		{
			title:  "error with nested causes",
			err:    NewWrappedError(err, NewWrappedError(reason, root)),
			expect: []byte(`{"error":"error","reason":{"error":"reason","reason":{"error":"root"}}}`),
		},
		{
			title:  "error with nil cause",
			err:    NewWrappedError(err, nil),
			expect: []byte(`{"error":"error","reason":{"error":"reason unknown"}}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			marshalled, err := json.Marshal(tc.err)
			if err != nil {
				t.Fatalf("error marshaling: %v", err)
			}

			if !bytes.Equal(marshalled, tc.expect) {
				t.Fatalf("failed unmarshaling expected %v got %v", string(tc.expect), string(marshalled))
			}

			unmashalled := &WrappedError{}
			err = json.Unmarshal(marshalled, unmashalled)
			if err != nil {
				t.Fatalf("error unmashaling: %v", err)
			}

			if !reflect.DeepEqual(tc.err, unmashalled) {
				t.Fatalf("failed marshaling expected %v got %v", tc.err, unmashalled)
			}
		})
	}
}
