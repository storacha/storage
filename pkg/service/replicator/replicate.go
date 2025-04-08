package replicator

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/capabilities/replica"
	"github.com/storacha/go-libstoracha/jobqueue"
	"github.com/storacha/go-ucanto/client"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/core/invocation"
	"github.com/storacha/go-ucanto/core/invocation/ran"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/message"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/receipt/fx"
	"github.com/storacha/go-ucanto/core/result"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/go-ucanto/ucan"

	cap_pdp "github.com/storacha/go-libstoracha/capabilities/pdp"

	"github.com/storacha/storage/pkg/service/capabilities"
	"github.com/storacha/storage/pkg/store/receiptstore"
)

var log = logging.Logger("replicator")

type Task struct {
	// bucket to associate with blob
	Space did.DID
	// the blob in question
	Blob blob.Blob
	// the location to replicate the blob from
	Source url.URL
	// the location to replicate the blob to
	Sink url.URL
	// invocation responsible for spawning this replication
	// should be a replica/transfer invocation
	Invocation invocation.Invocation
}

type Replicator interface {
	Replicate(context.Context, *Task) error
}

type Service struct {
	queue         *jobqueue.JobQueue[*Task]
	uploadService client.Connection
	id            principal.Signer
	capabilities  capabilities.Capabilities
	receipts      receiptstore.ReceiptStore
}

func New(
	signer principal.Signer,
	caps capabilities.Capabilities,
	rctStore receiptstore.ReceiptStore,
	uploadService client.Connection,
) (*Service, error) {

	repl := &Service{
		uploadService: uploadService,
		id:            signer,
		capabilities:  caps,
		receipts:      rctStore,
	}

	replicationQueue := jobqueue.NewJobQueue[*Task](
		jobqueue.JobHandler(repl.replicate),
		jobqueue.WithErrorHandler(func(err error) {
			log.Errorf("error while handling replication request: %s", err)
		}),
	)

	repl.queue = replicationQueue

	return repl, nil

}

func (r *Service) Replicate(ctx context.Context, task *Task) error {
	return r.queue.Queue(ctx, task)
}

func (r *Service) Start() error {
	r.queue.Startup()
	return nil
}

func (r *Service) Stop(ctx context.Context) error {
	return r.queue.Shutdown(ctx)
}

func (r *Service) replicate(ctx context.Context, task *Task) error {
	// pull the data from the source
	replicaResp, err := http.Get(task.Source.String())
	if err != nil {
		return fmt.Errorf("http get replication source (%s) failed: %w", task.Source.String(), err)
	}

	// stream the source to the sink
	req, err := http.NewRequest(http.MethodPut, task.Sink.String(), replicaResp.Body)
	if err != nil {
		return fmt.Errorf("failed to create replication sink request: %w", err)
	}
	req.Header = replicaResp.Header
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf(
			"failed http PUT to replicate blob %s from %s to %s failed: %w",
			task.Blob.Digest,
			task.Source.String(),
			task.Sink.String(),
			err,
		)
	}
	// verify status codes
	if res.StatusCode >= 300 || res.StatusCode < 200 {
		topErr := fmt.Errorf(
			"unsuccessful http PUT to replicate blob %s from %s to %s status code %d",
			task.Blob.Digest,
			task.Source.String(),
			task.Sink.String(),
			res.StatusCode,
		)
		resData, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("%s failed to read replication sink response body: %w", topErr, err)
		}
		return fmt.Errorf("%s response body: %s: %w", topErr, resData, err)
	}

	acceptResp, err := r.capabilities.BlobAccept(ctx, &capabilities.BlobAcceptRequest{
		Space: task.Space,
		Blob:  task.Blob,
		Put: blob.Promise{
			UcanAwait: blob.Await{
				Selector: ".out.ok",
				// TODO IDK what this does, or what to put here, or if we even need it?
				Link: task.Invocation.Link(),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to accept replication source blob %s: %w", task.Blob.Digest, err)
	}

	var forks []fx.Effect
	forks = append(forks, fx.FromInvocation(acceptResp.Claim))

	var pdpLink *ucan.Link
	if acceptResp.Piece != nil {
		// generate the invocation that will complete when aggregation is complete and the piece is accepted
		pieceAccept, err := cap_pdp.Accept.Invoke(
			r.id,
			// TODO validate this is the correct audience
			r.uploadService.ID().DID(),
			r.id.DID().String(),
			cap_pdp.AcceptCaveats{
				Piece: *acceptResp.Piece,
			}, delegation.WithNoExpiration())
		if err != nil {
			log.Errorw("creating piece accept invocation", "error", err)
			return fmt.Errorf("creating piece accept invocation: %w", err)
		}
		pieceAcceptLink := pieceAccept.Link()
		pdpLink = &pieceAcceptLink
		forks = append(forks, fx.FromInvocation(pieceAccept))
	}

	ok := result.Ok[replica.TransferOk, ipld.Builder](replica.TransferOk{
		Site: acceptResp.Claim.Link(),
		PDP:  pdpLink,
	})
	var opts []receipt.Option
	if len(forks) > 1 {
		opts = append(opts, receipt.WithFork(forks...))
	}
	rcpt, err := receipt.Issue(r.id, ok, ran.FromInvocation(task.Invocation), opts...)
	if err != nil {
		return fmt.Errorf("issuing receipt: %w", err)
	}
	if err := r.receipts.Put(ctx, rcpt); err != nil {
		return fmt.Errorf("failed to put transfer receipt: %w", err)
	}

	msg, err := message.Build([]invocation.Invocation{task.Invocation}, []receipt.AnyReceipt{rcpt})
	if err != nil {
		return fmt.Errorf("building message for receipt failed: %w", err)
	}

	request, err := r.uploadService.Codec().Encode(msg)
	if err != nil {
		return fmt.Errorf("failed to encode message for receipt to http request: %w", err)
	}

	response, err := r.uploadService.Channel().Request(request)
	if err != nil {
		return fmt.Errorf("failed to send request for receipt: %w", err)
	}
	if response.Status() >= 300 || response.Status() < 200 {
		topErr := fmt.Errorf("unsuccessful http POST to upload service")
		resData, err := io.ReadAll(response.Body())
		if err != nil {
			return fmt.Errorf("%s failed to read replication sink response body: %w", topErr, err)
		}
		return fmt.Errorf("%s response body: %s: %w", topErr, resData, err)
	}

	return nil
}
