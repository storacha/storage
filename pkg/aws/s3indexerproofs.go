package aws

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/storacha/go-ucanto/core/delegation"
)

// S3IndexerProofs returns delegation proofs for the indexer
type S3IndexerProofs struct {
	bucket    string
	keyPrefix string
	s3Client  *s3.Client
}

// Get implements store.Store.
func (s *S3IndexerProofs) Get(ctx context.Context) ([]delegation.Proof, error) {
	list, err := s.s3Client.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(s.keyPrefix),
	})
	if err != nil {
		return nil, fmt.Errorf("listing objects: %w", err)
	}
	proofs := make([]delegation.Proof, 0, len(list.Contents))
	for _, obj := range list.Contents {

		outPut, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    obj.Key,
		})
		if err != nil {
			return nil, fmt.Errorf("fetching proof from S3: %w", err)
		}
		data, err := io.ReadAll(outPut.Body)
		if err != nil {
			return nil, fmt.Errorf("reading proof data: %w", err)
		}
		d, err := delegation.Extract(data)
		if err != nil {
			return nil, fmt.Errorf("decoding proof: %w", err)
		}
		proofs = append(proofs, delegation.FromDelegation(d))
	}
	return proofs, nil
}

func NewS3IndexerProofs(cfg aws.Config, bucket string, keyPrefix string, opts ...func(*s3.Options)) *S3IndexerProofs {
	return &S3IndexerProofs{
		s3Client:  s3.NewFromConfig(cfg, opts...),
		bucket:    bucket,
		keyPrefix: keyPrefix,
	}
}
