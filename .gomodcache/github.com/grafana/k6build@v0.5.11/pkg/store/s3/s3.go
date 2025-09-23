// Package s3 implements a s3-backed object store
package s3

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	"github.com/grafana/k6build"
	"github.com/grafana/k6build/pkg/store"
)

// DefaultURLExpiration Default expiration for the presigned download URLs.
// After this time attempts to download the object will fail
// TODO: check this default (AWS default is 900 seconds)
const DefaultURLExpiration = time.Hour * 24

// Store a ObjectStore backed by a S3 bucket
type Store struct {
	bucket     string
	client     *s3.Client
	expiration time.Duration
}

// Config S3 Store configuration
type Config struct {
	// Name of the S3 bucket
	Bucket string
	// S3 Client
	Client *s3.Client
	// AWS endpoint (used for testing)
	Endpoint string
	// AWS Region
	Region string
	// Expiration for the presigned download URLs
	URLExpiration time.Duration
}

// returns the S3 client options
func (c Config) s3Opts() []func(o *s3.Options) {
	opts := []func(o *s3.Options){}

	if c.Endpoint != "" {
		opts = append(opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(c.Endpoint)
			o.UsePathStyle = true
		})
	}
	return opts
}

// returns the aws configuration load options from Config
func (c Config) awsOpts() []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	if c.Region != "" {
		opts = append(opts, config.WithRegion(c.Region))
	}

	return opts
}

// WithExpiration sets the expiration for the presigned URL
func WithExpiration(exp time.Duration) func(*s3.PresignOptions) {
	return func(opts *s3.PresignOptions) {
		opts.Expires = exp
	}
}

// New creates an object store backed by a S3 bucket
func New(conf Config) (store.ObjectStore, error) {
	if conf.Bucket == "" {
		return nil, fmt.Errorf("%w: bucket name cannot be empty", store.ErrInitializingStore)
	}

	client := conf.Client
	if client == nil {
		cfg, err := config.LoadDefaultConfig(context.TODO(), conf.awsOpts()...)
		if err != nil {
			return nil, k6build.NewWrappedError(store.ErrInitializingStore, err)
		}
		client = s3.NewFromConfig(cfg, conf.s3Opts()...)
	}

	expiration := conf.URLExpiration
	if expiration == 0 {
		expiration = DefaultURLExpiration
	}
	return &Store{
		client:     client,
		bucket:     conf.Bucket,
		expiration: expiration,
	}, nil
}

// Put stores the object and returns the metadata
// Fails if the object already exists
func (s *Store) Put(ctx context.Context, id string, content io.Reader) (store.Object, error) {
	if id == "" {
		return store.Object{}, fmt.Errorf("%w: id cannot be empty", store.ErrCreatingObject)
	}

	buff, err := io.ReadAll(content)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrCreatingObject, err)
	}

	checksum := sha256.Sum256(buff)
	_, err = s.client.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:            aws.String(s.bucket),
			Key:               aws.String(id),
			Body:              bytes.NewReader(buff),
			ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
			ChecksumSHA256:    aws.String(base64.StdEncoding.EncodeToString(checksum[:])),
			IfNoneMatch:       aws.String("*"),
		},
	)
	if err != nil {
		// check for duplicated object
		var aerr smithy.APIError
		if errors.As(err, &aerr) && aerr.ErrorCode() == "PreconditionFailed" {
			return store.Object{}, fmt.Errorf("%w: %q", store.ErrDuplicateObject, id)
		}
		return store.Object{}, k6build.NewWrappedError(store.ErrCreatingObject, err)
	}

	url, err := s.getDownloadURL(ctx, id)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrCreatingObject, err)
	}

	return store.Object{
		ID:       id,
		Checksum: fmt.Sprintf("%x", checksum),
		URL:      url,
	}, nil
}

// Get retrieves an objects if exists in the object store or an error otherwise
func (s *Store) Get(ctx context.Context, id string) (store.Object, error) {
	obj, err := s.client.GetObjectAttributes(
		ctx,
		&s3.GetObjectAttributesInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(id),
			ObjectAttributes: []types.ObjectAttributes{
				types.ObjectAttributesChecksum,
				types.ObjectAttributesEtag,
			},
		},
	)
	if err != nil {
		// check for object not found
		var aerr smithy.APIError
		if errors.As(err, &aerr) && aerr.ErrorCode() == "NoSuchKey" {
			return store.Object{}, fmt.Errorf("%w (%s)", store.ErrObjectNotFound, id)
		}

		return store.Object{}, k6build.NewWrappedError(store.ErrAccessingObject, err)
	}

	url, err := s.getDownloadURL(ctx, id)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrAccessingObject, err)
	}

	checksum, err := base64.StdEncoding.DecodeString(*obj.Checksum.ChecksumSHA256)
	if err != nil {
		return store.Object{}, k6build.NewWrappedError(store.ErrAccessingObject, err)
	}
	return store.Object{
		ID:       id,
		Checksum: fmt.Sprintf("%x", checksum),
		URL:      url,
	}, nil
}

func (s *Store) getDownloadURL(ctx context.Context, id string) (string, error) {
	// create a presigned get request to get the download URL
	request, err := s3.NewPresignClient(s.client).PresignGetObject(
		ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(id),
		},
		WithExpiration(s.expiration),
	)
	if err != nil {
		return "", k6build.NewWrappedError(store.ErrCreatingObject, err)
	}

	return request.URL, nil
}
