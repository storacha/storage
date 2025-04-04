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
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/receipt/fx"
	"github.com/storacha/go-ucanto/core/result"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/principal"
	http2 "github.com/storacha/go-ucanto/transport/http"
	"github.com/storacha/go-ucanto/ucan"

	cap_pdp "github.com/storacha/go-libstoracha/capabilities/pdp"

	"github.com/storacha/storage/pkg/pdp"
	"github.com/storacha/storage/pkg/service/blobs"
	"github.com/storacha/storage/pkg/service/claims"
	"github.com/storacha/storage/pkg/service/storage"
	"github.com/storacha/storage/pkg/store/receiptstore"
)

var log = logging.Logger("replicator")

var (
	UploadServiceDID, _ = did.Parse("did:web:upload.storacha.network")
	UploadServiceURL, _ = url.Parse("https://upload.storacha.network")
)

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

type LocalReplicator struct {
	queue *jobqueue.JobQueue[*Task]
	r     *SimpleReplicator
}

func NewLocalReplicator(c claims.Claims, b blobs.Blobs, p pdp.PDP, r receiptstore.ReceiptStore, i principal.Signer) (*LocalReplicator, error) {
	sr, err := NewSimpleReplicator(c, b, p, r, i)
	if err != nil {
		return nil, err
	}

	replicationQueue := jobqueue.NewJobQueue[*Task](
		jobqueue.JobHandler(sr.replicate),
		jobqueue.WithErrorHandler(func(err error) {
			log.Errorf("error while handling replication request: %s", err)
		}),
		jobqueue.WithBuffer(10),
	)

	return &LocalReplicator{
		queue: replicationQueue,
		r:     sr,
	}, nil
}

func (r *LocalReplicator) Enqueue(ctx context.Context, task *Task) error {
	return r.queue.Queue(ctx, task)
}

func (r *LocalReplicator) Start(ctx context.Context) error {
	r.queue.Startup()
	return nil
}

func (r *LocalReplicator) Stop(ctx context.Context) error {
	return r.queue.Shutdown(ctx)
}

type storageService struct {
	claims   claims.Claims
	blobs    blobs.Blobs
	pdp      pdp.PDP
	receipts receiptstore.ReceiptStore
	id       principal.Signer
}

func (s *storageService) ID() principal.Signer {
	return s.id
}

func (s *storageService) PDP() pdp.PDP {
	return s.pdp
}

func (s *storageService) Blobs() blobs.Blobs {
	return s.blobs
}

func (s *storageService) Claims() claims.Claims {
	return s.claims
}

func (s *storageService) Receipts() receiptstore.ReceiptStore {
	return s.receipts
}

func NewSimpleReplicator(c claims.Claims, b blobs.Blobs, p pdp.PDP, r receiptstore.ReceiptStore, i principal.Signer) (*SimpleReplicator, error) {
	channel := http2.NewHTTPChannel(UploadServiceURL)
	conn, err := client.NewConnection(UploadServiceDID, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to upload service: %w", err)
	}

	return &SimpleReplicator{
		uploadServiceConn: conn,
		storageService: &storageService{
			claims:   c,
			blobs:    b,
			pdp:      p,
			receipts: r,
			id:       i,
		}}, nil
}

type SimpleReplicator struct {
	storageService    storage.Service
	uploadServiceConn client.Connection
}

func (r *SimpleReplicator) replicate(ctx context.Context, task Task) error {
	// pull the data from the source
	replicaResp, err := http.Get(task.Source.String())
	if err != nil {
		return fmt.Errorf("http get replication source (%s) failed: %w", task.Source, err)
	}

	// stream the source to the sink
	req, err := http.NewRequest(http.MethodPut, task.Sink.String(), replicaResp.Body)
	if err != nil {
		return fmt.Errorf("failed to create replication sink request: %w", err)
	}
	req.Header = replicaResp.Header
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed http PUT to replicate blob %s from %s to %s failed: %w", task.Blob.Digest, task.Source, task.Sink, err)
	}
	// verify status codes
	if res.StatusCode >= 300 || res.StatusCode < 200 {
		err := fmt.Errorf("unsuccessful http PUT to replicate blob %s from %s to %s status code %d", task.Blob.Digest, task.Source, task.Sink, res.StatusCode)
		resData, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read replication sink response body: %w", err)
		}
		return fmt.Errorf("response body: %s: %w", resData, err)
	}

	// TODO this is a really gross way to have a dep, but can refactor later
	acceptResp, err := storage.BlobAccept(ctx, r.storageService, &storage.BlobAcceptRequest{
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
			r.storageService.ID(),
			// TODO validate this is the correct audience
			UploadServiceDID,
			r.storageService.ID().DID().String(),
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

	ok := result.Ok[replica.TransferOK, ipld.Builder](replica.TransferOK{
		Site: acceptResp.Claim.Link(),
		PDP:  pdpLink,
	})
	rcpt, err := receipt.Issue(r.storageService.ID(), ok, ran.FromInvocation(task.Invocation))
	if err != nil {
		return fmt.Errorf("issuing receipt: %w", err)
	}
	if err := r.storageService.Receipts().Put(ctx, rcpt); err != nil {
		return fmt.Errorf("failed to put transfer receipt: %w", err)
	}
	// TODO how does one send a receipt to the indexing service
	client.Execute(rcpt, r.uploadServiceConn)
	/*
		receipt := replica.TransferOK{
			Site: acceptResp.Claim.Link(),
			PDP:  pdpLink,
		}
	*/
}
