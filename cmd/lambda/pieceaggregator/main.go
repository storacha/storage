package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-capabilities/pkg/types"
	"github.com/storacha/go-piece/pkg/piece"
	"github.com/storacha/storage/cmd/lambda"
	"github.com/storacha/storage/internal/ipldstore"
	"github.com/storacha/storage/pkg/aws"
	"github.com/storacha/storage/pkg/pdp/aggregator"
	"github.com/storacha/storage/pkg/pdp/aggregator/aggregate"
)

func makeHandler(cfg aws.Config) (lambda.SQSEventHandler, error) {
	inProgressWorkspace := aggregator.NewInProgressWorkspace(aws.NewS3Store(cfg.Config, cfg.BufferBucket, cfg.BufferPrefix))
	aggregateStore := ipldstore.IPLDStore[datamodel.Link, aggregate.Aggregate](
		aws.NewS3Store(cfg.Config, cfg.AggregatesBucket, cfg.AggregatesPrefix),
		aggregate.AggregateType(), types.Converters...)
	aggregateSubmitterQueue := aws.NewSQSAggregateQueue(cfg.Config, cfg.SQSPDPAggregateSubmitterURL)
	pieceAggregator := aggregator.NewPieceAggregator(inProgressWorkspace, aggregateStore, aggregateSubmitterQueue.Queue)

	return func(ctx context.Context, sqsEvent events.SQSEvent) error {
		// process messages in parallel
		pieceLinks := make([]piece.PieceLink, 0, len(sqsEvent.Records))
		for _, msg := range sqsEvent.Records {
			var pieceLinkMessage aws.PieceLinkMessage
			err := json.Unmarshal([]byte(msg.Body), &pieceLinkMessage)
			if err != nil {
				return fmt.Errorf("deserializing message json: %w", err)
			}
			c, err := cid.Decode(pieceLinkMessage.Piece)
			if err != nil {
				return fmt.Errorf("decoding piece link: %w", err)
			}
			pieceLink, err := piece.FromLink(cidlink.Link{Cid: c})
			if err != nil {
				return fmt.Errorf("casting to piece link: %w", err)
			}
			pieceLinks = append(pieceLinks, pieceLink)
		}
		return pieceAggregator.AggregatePieces(ctx, pieceLinks)
	}, nil
}

func main() {
	lambda.StartSQSEventHandler(makeHandler)
}
