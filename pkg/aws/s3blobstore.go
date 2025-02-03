package aws

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	multihash "github.com/multiformats/go-multihash"
	"github.com/storacha/storage/pkg/internal/digestutil"
	"github.com/storacha/storage/pkg/presigner"
	"github.com/storacha/storage/pkg/store/blobstore"
)

// S3BlobStore implements the blobstore.BlobStore interface on S3
type S3BlobStore struct {
	bucket    string
	keyPrefix string
	s3Client  *s3.Client
}

var _ blobstore.Blobstore = (*S3BlobStore)(nil)

func NewS3BlobStore(cfg aws.Config, bucket string, keyPrefix string, opts ...func(*s3.Options)) *S3BlobStore {
	return &S3BlobStore{
		s3Client:  s3.NewFromConfig(cfg, opts...),
		bucket:    bucket,
		keyPrefix: keyPrefix,
	}
}

var _ blobstore.Object = (*s3BlobObject)(nil)

type S3BlobPresigner struct {
	bs            *S3BlobStore
	presignClient *s3.PresignClient
}

// SignUploadURL implements presigner.RequestPresigner.
func (s *S3BlobPresigner) SignUploadURL(ctx context.Context, digest multihash.Multihash, size uint64, ttl uint64) (url.URL, http.Header, error) {
	signedReq, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bs.bucket),
		Key:           aws.String(s.bs.keyPrefix + digestutil.Format(digest)),
		ContentLength: aws.Int64(int64(size)),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(int64(ttl) * int64(time.Second))
	})
	if err != nil {
		return url.URL{}, nil, fmt.Errorf("signing request: %w", err)
	}

	reqURL, err := url.Parse(signedReq.URL)
	if err != nil {
		return url.URL{}, nil, fmt.Errorf("parsing signed URL: %w", err)
	}

	return *reqURL, signedReq.SignedHeader, nil
}

// VerifyUploadURL implements presigner.RequestPresigner.
func (s *S3BlobPresigner) VerifyUploadURL(ctx context.Context, url url.URL, headers http.Header) (url.URL, http.Header, error) {
	panic("unimplemented")
}

var _ presigner.RequestPresigner = (*S3BlobPresigner)(nil)

func (s *S3BlobStore) PresignClient() presigner.RequestPresigner {
	presignClient := s3.NewPresignClient(s.s3Client)
	return &S3BlobPresigner{s, presignClient}
}

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
