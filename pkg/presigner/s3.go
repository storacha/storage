package presigner

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/piri/pkg/internal/digestutil"
)

const ISO8601BasicFormat = "20060102T150405Z"

type S3RequestPresigner struct {
	endpoint      url.URL
	bucketName    string
	presignClient *s3.PresignClient
}

func encodeKey(digest multihash.Multihash) string {
	return digestutil.Format(digest)
}

func (ss *S3RequestPresigner) SignUploadURL(ctx context.Context, digest multihash.Multihash, size uint64, ttl uint64) (url.URL, http.Header, error) {
	signedReq, err := ss.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(ss.bucketName),
		Key:           aws.String(encodeKey(digest)),
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

// pointInTimePresigner is a [s3.HTTPPresignerV4] whose signing time is frozen
// to the preconfigured value.
type pointInTimePresigner struct {
	signingTime time.Time
	presigner   s3.HTTPPresignerV4
}

func (pps pointInTimePresigner) PresignHTTP(
	ctx context.Context, credentials aws.Credentials, r *http.Request,
	payloadHash string, service string, region string, signingTime time.Time,
	optFns ...func(*v4.SignerOptions),
) (url string, signedHeader http.Header, err error) {
	return pps.presigner.PresignHTTP(ctx, credentials, r, payloadHash, service,
		region, pps.signingTime, optFns...)
}

func (ss *S3RequestPresigner) VerifyUploadURL(ctx context.Context, requestURL url.URL, requestHeaders http.Header) (url.URL, http.Header, error) {
	requestURL = *ss.endpoint.ResolveReference(&requestURL)
	key := strings.Join(strings.Split(requestURL.Path, "/")[2:], "/")

	contentLength, err := strconv.ParseInt(requestHeaders.Get("Content-Length"), 10, 64)
	if err != nil {
		return url.URL{}, nil, fmt.Errorf("parsing Content-Length header: %w", err)
	}

	expires, err := strconv.ParseInt(requestURL.Query().Get("X-Amz-Expires"), 10, 64)
	if err != nil {
		return url.URL{}, nil, fmt.Errorf("parsing X-Amz-Expires parameter: %w", err)
	}

	signingTime, err := time.Parse(ISO8601BasicFormat, requestURL.Query().Get("X-Amz-Date"))
	if err != nil {
		return url.URL{}, nil, fmt.Errorf("parsing X-Amz-Date parameter: %w", err)
	}

	signedReq, err := ss.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(ss.bucketName),
		Key:           aws.String(key),
		ContentLength: aws.Int64(contentLength),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(expires * int64(time.Second))
		// configure the presigner for the time the original signing took place.
		ps := opts.Presigner
		stp := pointInTimePresigner{signingTime, ps}
		opts.Presigner = stp
	})
	if err != nil {
		return url.URL{}, nil, fmt.Errorf("signing request: %w", err)
	}

	if requestURL.String() != signedReq.URL {
		return url.URL{}, nil, errors.New("signature verification failed")
	}

	u, err := url.Parse(signedReq.URL)
	if err != nil {
		return url.URL{}, nil, fmt.Errorf("parsing signed URL: %w", err)
	}

	return *u, signedReq.SignedHeader, nil
}

var _ RequestPresigner = (*S3RequestPresigner)(nil)

// NewS3RequestPresigner creates a signer that the S3 SDK to sign and verify
// requests. The bucketName parameter is optional and defaults to "blob".
//
// Signed upload URLs take the form {endpoint}/{bucketName}/{b58digest}
func NewS3RequestPresigner(accessKeyID string, secretAcessKey string, endpoint url.URL, bucketName string) (*S3RequestPresigner, error) {
	endpointstr := endpoint.String()

	var credsProvider aws.CredentialsProviderFunc = func(context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAcessKey,
		}, nil
	}

	cfg := aws.Config{
		Region:       "auto",
		Credentials:  credsProvider,
		BaseEndpoint: &endpointstr,
	}

	s3client := s3.NewFromConfig(cfg, func(opts *s3.Options) {
		opts.UsePathStyle = true
	})
	presign := s3.NewPresignClient(s3client)

	if bucketName == "" {
		bucketName = "blob"
	}

	return &S3RequestPresigner{endpoint, bucketName, presign}, nil
}
