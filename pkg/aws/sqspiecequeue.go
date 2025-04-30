package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/piece/piece"

	"github.com/storacha/storage/pkg/pdp/aggregator"
)

var PieceQueueMessageGroupID = "piece-queue"

// PieceLinkMessage is the struct that is serialized onto an SQS message queue in JSON
type PieceLinkMessage struct {
	Piece string `json:"piece,omitempty"`
}

// SQSPieceQueue implements the providercacher.CachingQueue interface using SQS
type SQSPieceQueue struct {
	queueURL  string
	sqsClient *sqs.Client
}

// NewSQSPieceQueue returns a new SQSCachingQueue for the given aws config
func NewSQSPieceQueue(cfg aws.Config, queurURL string) *SQSPieceQueue {
	return &SQSPieceQueue{
		queueURL:  queurURL,
		sqsClient: sqs.NewFromConfig(cfg),
	}
}

// Queue implements blobindexlookup.CachingQueue.
func (s *SQSPieceQueue) Queue(ctx context.Context, piece piece.PieceLink) error {
	messageJSON, err := json.Marshal(PieceLinkMessage{Piece: piece.Link().String()})
	if err != nil {
		return fmt.Errorf("serializing message json: %w", err)
	}
	_, err = s.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:       aws.String(s.queueURL),
		MessageBody:    aws.String(string(messageJSON)),
		MessageGroupId: &PieceQueueMessageGroupID,
	})
	if err != nil {
		return fmt.Errorf("enqueueing message: %w", err)
	}
	return nil
}

var _ aggregator.QueuePieceAggregationFn = (&SQSPieceQueue{}).Queue

// DecodePieceMessage extracts a piece link from an SQS queue messagebody
func DecodePieceMessage(messageBody string) (piece.PieceLink, error) {
	var msg PieceLinkMessage
	err := json.Unmarshal([]byte(messageBody), &msg)
	if err != nil {
		return nil, fmt.Errorf("deserializing message: %w", err)
	}
	c, err := cid.Decode(msg.Piece)
	if err != nil {
		return nil, fmt.Errorf("decoding link: %w", err)
	}
	piece, err := piece.FromLink(cidlink.Link{Cid: c})
	if err != nil {
		return nil, fmt.Errorf("decoding piece link: %w", err)
	}
	return piece, nil
}
