package s3

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/docker/go-connections/nat"
	"github.com/grafana/k6build/pkg/store"

	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

type object struct {
	id      string
	content []byte
}

func s3Client(ctx context.Context, l *localstack.LocalStackContainer) (*s3.Client, error) {
	region := "us-east-1"
	host, err := l.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := l.MappedPort(ctx, nat.Port("4566/tcp"))
	if err != nil {
		return nil, err
	}

	awsEndP := fmt.Sprintf("http://%s:%s", host, mappedPort.Port()) //nolint:nosprintfhostport
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),         //nolint:contextcheck
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("accesskey", "secretkey", "token")),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(awsEndP)
		o.UsePathStyle = true
	})

	return client, nil
}

func setupStore(preload []object) (store.ObjectStore, error) {
	bucket := "test"

	localstack, err := localstack.Run(context.TODO(), "localstack/localstack:latest")
	if err != nil {
		return nil, fmt.Errorf("localstack setup %w", err)
	}

	client, err := s3Client(context.TODO(), localstack)
	if err != nil {
		return nil, fmt.Errorf("creating s3 client %w", err)
	}

	_, err = client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 setup %w", err)
	}

	for _, o := range preload {
		checksum := sha256.Sum256(o.content)
		_, err = client.PutObject(
			context.TODO(),
			&s3.PutObjectInput{
				Bucket:            aws.String(bucket),
				Key:               aws.String(o.id),
				Body:              bytes.NewReader(o.content),
				ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
				ChecksumSHA256:    aws.String(base64.StdEncoding.EncodeToString(checksum[:])),
			},
		)
		if err != nil {
			return nil, fmt.Errorf("preload setup %w", err)
		}
	}

	store, err := New(Config{Client: client, Bucket: bucket})
	if err != nil {
		return nil, fmt.Errorf("create store %w", err)
	}

	return store, nil
}

func TestPutObject(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "linux" {
		t.Skip("Skipping test: localstack test container is failing in darwin and windows")
	}

	preload := []object{
		{
			id:      "existing-object",
			content: []byte("content"),
		},
	}

	s, err := setupStore(preload)
	if err != nil {
		t.Fatalf("test setup %v", err)
	}

	testCases := []struct {
		title     string
		preload   []object
		id        string
		content   []byte
		expectErr error
	}{
		{
			title:   "put object",
			id:      "new-object",
			content: []byte("content"),
		},
		{
			title:     "put existing object",
			id:        "existing-object",
			content:   []byte("new content"),
			expectErr: store.ErrDuplicateObject,
		},
		{
			title:   "put empty object",
			id:      "empty",
			content: nil,
		},
		{
			title:     "put empty id",
			id:        "",
			content:   []byte("content"),
			expectErr: store.ErrCreatingObject,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			obj, err := s.Put(context.TODO(), tc.id, bytes.NewBuffer(tc.content))
			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}

			// if expected error, don't validate object
			if tc.expectErr != nil {
				return
			}

			_, err = url.Parse(obj.URL)
			if err != nil {
				t.Fatalf("invalid url %v", err)
			}

			resp, err := http.Get(obj.URL)
			if err != nil {
				t.Fatalf("reading object url %v", err)
			}
			defer resp.Body.Close() //nolint:errcheck

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("reading object url %s", resp.Status)
			}

			content, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("reading object content %v", err)
			}

			if !bytes.Equal(tc.content, content) {
				t.Fatalf("expected %v got %v", tc.content, content)
			}
		})
	}
}

func TestGetObject(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "linux" {
		t.Skip("Skipping test: localstack test container is failing in darwin and windows")
	}

	preload := []object{
		{
			id:      "existing-object",
			content: []byte("content"),
		},
	}

	s, err := setupStore(preload)
	if err != nil {
		t.Fatalf("test setup %v", err)
	}

	testCases := []struct {
		title     string
		preload   []object
		id        string
		expect    []byte
		expectErr error
	}{
		{
			title:     "get existing object",
			id:        "existing-object",
			expect:    []byte("content"),
			expectErr: nil,
		},
		{
			title:     "get non-existing object",
			id:        "non-existing-object",
			expectErr: store.ErrObjectNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			obj, err := s.Get(context.TODO(), tc.id)
			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}

			// if expected error, don't validate object
			if tc.expectErr != nil {
				return
			}

			_, err = url.Parse(obj.URL)
			if err != nil {
				t.Fatalf("invalid url %v", err)
			}

			resp, err := http.Get(obj.URL)
			if err != nil {
				t.Fatalf("reading object url %v", err)
			}
			defer resp.Body.Close() //nolint:errcheck

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("reading object url %s", resp.Status)
			}

			content, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("reading object content %v", err)
			}

			if !bytes.Equal(tc.expect, content) {
				t.Fatalf("expected %v got %v", tc.expect, content)
			}
		})
	}
}
