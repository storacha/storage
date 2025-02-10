package lambda

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/storacha/storage/internal/telemetry"
	"github.com/storacha/storage/pkg/aws"
)

// SQSEventHandler is a function that handles SQS events, suitable to use as a lambda handler.
type SQSEventHandler func(context.Context, events.SQSEvent) error

// SQSEventHandlerBuilder is a function that creates a SQSEventHandler from a config.
type SQSEventHandlerBuilder func(aws.Config) (SQSEventHandler, error)

// StartSQSEventHandler starts a lambda handler that processes SQS events.
func StartSQSEventHandler(makeHandler SQSEventHandlerBuilder) {
	ctx := context.Background()
	cfg := aws.FromEnv(ctx)
	telemetry.SetupErrorReporting(cfg.SentryDSN, cfg.SentryEnvironment)

	handler, err := makeHandler(cfg)
	if err != nil {
		telemetry.ReportError(err)
		panic(err)
	}

	lambda.StartWithOptions(instrumentSQSEventHandler(handler), lambda.WithContext(ctx))
}

// instrumentSQSEventHandler wraps a SQSEventHandler with error reporting.
func instrumentSQSEventHandler(handler SQSEventHandler) SQSEventHandler {
	return func(ctx context.Context, sqsEvent events.SQSEvent) error {
		err := handler(ctx, sqsEvent)
		if err != nil {
			telemetry.ReportError(err)
		}

		return err
	}
}

// HTTPHandlerBuilder is a function that creates a http.Handler from a config.
type HTTPHandlerBuilder func(aws.Config) (http.Handler, error)

// StartHTTPHandler starts a lambda handler that processes HTTP requests.
func StartHTTPHandler(makeHandler HTTPHandlerBuilder) {
	ctx := context.Background()
	cfg := aws.FromEnv(ctx)
	telemetry.SetupErrorReporting(cfg.SentryDSN, cfg.SentryEnvironment)

	handler, err := makeHandler(cfg)
	if err != nil {
		telemetry.ReportError(err)
		panic(err)
	}

	lambda.StartWithOptions(httpadapter.NewV2(handler).ProxyWithContext, lambda.WithContext(ctx))
}
