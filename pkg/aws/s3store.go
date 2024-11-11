package aws

import (
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/storacha/ipni-publisher/pkg/store"
)

// S3Store implements the store.Store interface on S3
type S3Store struct {
	bucket    string
	keyPrefix string
	s3Client  *s3.Client
}

var _ store.Store = (*S3Store)(nil)

// Get implements store.Store.
func (s *S3Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	outPut, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.keyPrefix + key),
	})
	if err != nil {
		var noSuchKeyError *types.NoSuchKey
		// wrap in error recognizable as a not found error for Store interface consumers
		if errors.As(err, &noSuchKeyError) {
			return nil, store.NewErrNotFound(err)
		}
		return nil, err
	}
	return outPut.Body, nil
}

// Put implements store.Store.
func (s *S3Store) Put(ctx context.Context, key string, data io.Reader) error {
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.keyPrefix + key),
		Body:   data,
	})
	return err
}

func NewS3Store(cfg aws.Config, bucket string, keyPrefix string) *S3Store {
	return &S3Store{
		s3Client:  s3.NewFromConfig(cfg),
		bucket:    bucket,
		keyPrefix: keyPrefix,
	}
}
