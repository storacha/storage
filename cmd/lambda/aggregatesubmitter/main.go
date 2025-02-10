package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-capabilities/pkg/types"
	"github.com/storacha/storage/internal/ipldstore"
	"github.com/storacha/storage/pkg/aws"
	"github.com/storacha/storage/pkg/pdp/aggregator"
	"github.com/storacha/storage/pkg/pdp/aggregator/aggregate"
	"github.com/storacha/storage/pkg/pdp/curio"
)

func makeHandler(aggregateSubmitter *aggregator.AggregateSubmitter) func(ctx context.Context, sqsEvent events.SQSEvent) error {
	return func(ctx context.Context, sqsEvent events.SQSEvent) error {
		// process messages in parallel
		aggregateLinks := make([]datamodel.Link, 0, len(sqsEvent.Records))
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

			aggregateLinks = append(aggregateLinks, cidlink.Link{Cid: c})
		}
		return aggregateSubmitter.SubmitAggregates(ctx, aggregateLinks)
	}
}

func main() {
	config := aws.FromEnv(context.Background())
	curioURL, err := url.Parse(config.CurioURL)
	if err != nil {
		panic(fmt.Errorf("parsing curio URL: %w", err))
	}
	curioAuth, err := curio.CreateCurioJWTAuthHeader("storacha", config.Signer)
	if err != nil {
		panic(fmt.Errorf("generating curio JWT: %w", err))
	}
	curioClient := curio.New(http.DefaultClient, curioURL, curioAuth)
	aggregateStore := ipldstore.IPLDStore[datamodel.Link, aggregate.Aggregate](
		aws.NewS3Store(config.Config, config.AggregatesBucket, config.AggregatesPrefix),
		aggregate.AggregateType(), types.Converters...)
	aggregateSubmitterQueue := aws.NewSQSAggregateQueue(config.Config, config.SQSPDPPieceAggregatorURL)
	aggregateSubmitter := aggregator.NewAggregateSubmitteer(config.PDPProofSet, aggregateStore, curioClient, aggregateSubmitterQueue.Queue)
	lambda.Start(makeHandler(aggregateSubmitter))
}
