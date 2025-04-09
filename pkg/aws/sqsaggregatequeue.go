package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/storacha/storage/pkg/pdp/aggregator"
)

// LinkMessage is a struct that is serialized onto an SQS message queue in JSON
type LinkMessage struct {
	Link string `json:"link,omitempty"`
}

// SQSAggregateQueue implements the providercacher.CachingQueue interface using SQS
type SQSAggregateQueue struct {
	queueURL  string
	sqsClient *sqs.Client
}

// NewSQSAggregateQueue returns a new SQSCachingQueue for the given aws config
func NewSQSAggregateQueue(cfg aws.Config, queurURL string) *SQSAggregateQueue {
	return &SQSAggregateQueue{
		queueURL:  queurURL,
		sqsClient: sqs.NewFromConfig(cfg),
	}
}

// Enqueue implements blobindexlookup.CachingQueue.
func (s *SQSAggregateQueue) Enqueue(ctx context.Context, _ string, link datamodel.Link) error {

	messageJSON, err := json.Marshal(LinkMessage{Link: link.String()})
	if err != nil {
		return fmt.Errorf("serializing message json: %w", err)
	}
	_, err = s.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(string(messageJSON)),
	})
	if err != nil {
		return fmt.Errorf("enqueueing message: %w", err)
	}
	return nil
}

var _ aggregator.LinkQueue = (*SQSAggregateQueue)(nil)

// DecodeLinkMessage extracts a link from an SQS queue messagebody
func DecodeLinkMessage(messageBody string) (datamodel.Link, error) {
	var msg LinkMessage
	err := json.Unmarshal([]byte(messageBody), &msg)
	if err != nil {
		return nil, fmt.Errorf("deserializing message: %w", err)
	}
	c, err := cid.Decode(msg.Link)
	if err != nil {
		return nil, fmt.Errorf("decoding link: %w", err)
	}
	return cidlink.Link{Cid: c}, nil
}
