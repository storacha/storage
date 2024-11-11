package aws

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	multihash "github.com/multiformats/go-multihash"
	"github.com/storacha/storage/pkg/internal/digestutil"
	"github.com/storacha/storage/pkg/store/blobstore"
)

// S3BlobStore implements the blobstore.BlobStore interface on S3
type S3BlobStore struct {
	bucket    string
	keyPrefix string
	s3Client  *s3.Client
}

var _ blobstore.Blobstore = (*S3BlobStore)(nil)

func NewS3BlobStore(cfg aws.Config, bucket string, keyPrefix string) *S3BlobStore {
	return &S3BlobStore{
		s3Client:  s3.NewFromConfig(cfg),
		bucket:    bucket,
		keyPrefix: keyPrefix,
	}
}

var _ blobstore.Object = (*s3BlobObject)(nil)

// Put implements blobstore.Blobstore.
func (s *S3BlobStore) Put(ctx context.Context, digest multihash.Multihash, size uint64, body io.Reader) error {
	key := digestutil.Format(digest)
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(s.keyPrefix + key),
		Body:          body,
		ContentLength: aws.Int64(int64(size)),
	})
	return err
}

// Get implements blobstore.Blobstore.
func (s *S3BlobStore) Get(ctx context.Context, digest multihash.Multihash, opts ...blobstore.GetOption) (blobstore.Object, error) {
	config := blobstore.NewGetConfig()
	config.ProcessOptions(opts)

	var rangeParam *string
	if config.Range().Offset != 0 || config.Range().Length != nil {
		rangeString := fmt.Sprintf("bytes=%d-", config.Range().Offset)
		if config.Range().Length != nil {
			rangeString += strconv.FormatUint(*config.Range().Length, 10)
		}
		rangeParam = &rangeString
	}
	key := digestutil.Format(digest)
	outPut, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.keyPrefix + key),
		Range:  rangeParam,
	})
	if err != nil {
		return nil, err
	}
	return &s3BlobObject{outPut}, nil
}

type s3BlobObject struct {
	outPut *s3.GetObjectOutput
}

// Body implements blobstore.Object.
func (s *s3BlobObject) Body() io.Reader {
	return s.outPut.Body
}

// Size implements blobstore.Object.
func (s *s3BlobObject) Size() int64 {
	return *s.outPut.ContentLength
}
