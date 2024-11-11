package aws

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/storacha/go-metadata"
	"github.com/storacha/go-piece/pkg/piece"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"

	"github.com/storacha/ipni-publisher/pkg/store"
	"github.com/storacha/storage/pkg/pdp"
	"github.com/storacha/storage/pkg/pdp/aggregator"
	"github.com/storacha/storage/pkg/pdp/curio"
	"github.com/storacha/storage/pkg/pdp/pieceadder"
	"github.com/storacha/storage/pkg/pdp/piecefinder"
	"github.com/storacha/storage/pkg/service/storage"
	"github.com/storacha/storage/pkg/store/delegationstore"
	"github.com/storacha/storage/pkg/store/receiptstore"
)

// ErrNoPrivateKey means that the value returned from Secrets was empty
var ErrNoPrivateKey = errors.New("no value for private key")

func mustGetEnv(envVar string) string {
	value := os.Getenv(envVar)
	if len(value) == 0 {
		panic(fmt.Errorf("missing env var: %s", envVar))
	}
	return value
}

var ErrIndexingServiceProofsMissing = errors.New("indexing service proofs are missing")

type AWSAggregator struct {
	pieceAggregatorQueue *SQSPieceQueue
}

// AggregatePiece is the frontend to aggregation
func (aa *AWSAggregator) AggregatePiece(ctx context.Context, pieceLink piece.PieceLink) error {
	return aa.pieceAggregatorQueue.Queue(ctx, pieceLink)
}

type PDP struct {
	aggregator  *AWSAggregator
	pieceAdder  pieceadder.PieceAdder
	pieceFinder piecefinder.PieceFinder
}

// Aggregator implements pdp.PDP.
func (p *PDP) Aggregator() aggregator.Aggregator {
	return p.aggregator
}

// PieceAdder implements pdp.PDP.
func (p *PDP) PieceAdder() pieceadder.PieceAdder {
	return p.pieceAdder
}

// PieceFinder implements pdp.PDP.
func (p *PDP) PieceFinder() piecefinder.PieceFinder {
	return p.pieceFinder
}

func NewPDP(cfg Config) (*PDP, error) {
	curioURL, err := url.Parse(cfg.CurioURL)
	if err != nil {
		return nil, fmt.Errorf("parsing curio URL: %w", err)
	}
	curioAuth, err := curio.CreateCurioJWTAuthHeader("storacha", cfg.Signer)
	if err != nil {
		return nil, fmt.Errorf("generating curio JWT: %w", err)
	}
	curioClient := curio.New(http.DefaultClient, curioURL, curioAuth)
	return &PDP{
		aggregator: &AWSAggregator{
			pieceAggregatorQueue: NewSQSPieceQueue(cfg.Config, cfg.SQSPDPPieceAggregatorURL),
		},
		pieceAdder:  pieceadder.NewCurioAdder(curioClient),
		pieceFinder: piecefinder.NewCurioFinder(curioClient),
	}, nil
}

var _ pdp.PDP = (*PDP)(nil)

type Config struct {
	aws.Config
	AllocationsTableName         string
	BlobStoreBucket              string
	BlobStorePrefix              string
	ChunkLinksTableName          string
	MetadataTableName            string
	IPNIStoreBucket              string
	IPNIStorePrefix              string
	ClaimStoreBucket             string
	ClaimStorePrefix             string
	PublicURL                    string
	AnnounceURL                  string
	IndexingServiceDID           string
	IndexingServiceURL           string
	IndexingServiceProofsBucket  string
	IndexingServiceProofsPrefix  string
	IPNIPublisherAnnounceAddress string
	RanLinkIndexTableName        string
	ReceiptStoreBucket           string
	ReceiptStorePrefix           string
	SQSPDPPieceAggregatorURL     string
	CurioURL                     string
	principal.Signer
}

// FromEnv constructs the AWS Configuration from the environment
func FromEnv(ctx context.Context) Config {
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("loading aws default config: %w", err))
	}
	ssmClient := ssm.NewFromConfig(awsConfig)
	response, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(mustGetEnv("PRIVATE_KEY")),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		panic(fmt.Errorf("retrieving private key: %w", err))
	}
	if response.Parameter == nil || response.Parameter.Value == nil {
		panic(ErrNoPrivateKey)
	}
	id, err := ed25519.Parse(*response.Parameter.Value)
	if err != nil {
		panic(fmt.Errorf("parsing private key: %s", err))
	}
	cryptoPrivKey, err := crypto.UnmarshalEd25519PrivateKey(id.Raw())
	if err != nil {
		panic(fmt.Errorf("unmarshaling private key: %w", err))
	}

	ipniStoreKeyPrefix := os.Getenv("IPNI_STORE_KEY_PREFIX")
	if len(ipniStoreKeyPrefix) == 0 {
		ipniStoreKeyPrefix = "/ipni/v1/ad/"
	}

	peer, err := peer.IDFromPrivateKey(cryptoPrivKey)
	if err != nil {
		panic(fmt.Errorf("parsing private key to peer: %w", err))
	}

	ipniPublisherAnnounceAddress := fmt.Sprintf("/dns4/%s/tcp/443/https/p2p/%s", mustGetEnv("IPNI_STORE_BUCKET_REGIONAL_DOMAIN"), peer.String())
	return Config{
		Config:                       awsConfig,
		Signer:                       id,
		ChunkLinksTableName:          mustGetEnv("CHUNK_LINKS_TABLE_NAME"),
		MetadataTableName:            mustGetEnv("METADATA_TABLE_NAME"),
		IPNIStoreBucket:              mustGetEnv("IPNI_STORE_BUCKET_NAME"),
		IPNIStorePrefix:              ipniStoreKeyPrefix,
		IPNIPublisherAnnounceAddress: ipniPublisherAnnounceAddress,
		ClaimStoreBucket:             mustGetEnv("CLAIM_STORE_BUCKET_NAME"),
		ClaimStorePrefix:             os.Getenv("CLAIM_STORE_KEY_REFIX"),
		AllocationsTableName:         mustGetEnv("ALLOCATIONS_TABLE_NAME"),
		BlobStoreBucket:              mustGetEnv("BLOB_STORE_BUCKET_NAME"),
		BlobStorePrefix:              mustGetEnv("BLOB_STORE_KEY_PREFIX"),
		AnnounceURL:                  mustGetEnv("IPNI_ENDPOINT"),
		PublicURL:                    mustGetEnv("PUBLIC_URL"),
		IndexingServiceDID:           mustGetEnv("INDEXING_SERVICE_DID"),
		IndexingServiceURL:           mustGetEnv("INDEXING_SERVICE_URL"),
		IndexingServiceProofsBucket:  mustGetEnv("INDEXING_SERVICE_PROOFS_BUCKET_NAME"),
		IndexingServiceProofsPrefix:  mustGetEnv("INDEXING_SERVICE_PROOFS_KEY_PREFIX"),
		RanLinkIndexTableName:        mustGetEnv("RAN_LINK_INDEX_TABLE_NAME"),
		ReceiptStoreBucket:           mustGetEnv("RECEIPT_STORE_BUCKET_NAME"),
		ReceiptStorePrefix:           mustGetEnv("RECEIPT_STORE_KEY_PREFIX"),
		SQSPDPPieceAggregatorURL:     os.Getenv("PIECE_AGGREGATOR_QUEUE_URL"),
		CurioURL:                     os.Getenv("CURIO_URL"),
	}
}

func Construct(cfg Config) (storage.Service, error) {

	blobStore := NewS3BlobStore(cfg.Config, cfg.BlobStoreBucket, cfg.BlobStorePrefix)
	allocationStore := NewDynamoAllocationStore(cfg.Config, cfg.AllocationsTableName)
	claimStore, err := delegationstore.NewDelegationStore(NewS3Store(cfg.Config, cfg.ClaimStoreBucket, cfg.ClaimStorePrefix))
	if err != nil {
		return nil, fmt.Errorf("constructing claim store: %w", err)
	}
	ipniStore := NewS3Store(cfg.Config, cfg.IPNIStoreBucket, cfg.IPNIStorePrefix)
	chunkLinksTable := NewDynamoProviderContextTable(cfg.Config, cfg.ChunkLinksTableName)
	metadataTable := NewDynamoProviderContextTable(cfg.Config, cfg.MetadataTableName)
	publisherStore := store.NewPublisherStore(ipniStore, chunkLinksTable, metadataTable, store.WithMetadataContext(metadata.MetadataContext))
	pubURL, err := url.Parse(cfg.PublicURL)
	if err != nil {
		return nil, fmt.Errorf("parsing public url: %w", err)
	}
	announceURL, err := url.Parse(cfg.AnnounceURL)
	if err != nil {
		return nil, fmt.Errorf("parsing announce url: %w", err)
	}
	indexingServiceDID, err := did.Parse(cfg.IndexingServiceDID)
	if err != nil {
		return nil, fmt.Errorf("parsing indexing service did: %w", err)
	}
	indexingServiceURL, err := url.Parse(cfg.IndexingServiceURL)
	if err != nil {
		return nil, fmt.Errorf("parsing indexing service url: %w", err)
	}
	indexingServiceProofs, err := NewS3IndexerProofs(cfg.Config, cfg.IndexingServiceProofsBucket, cfg.IndexingServiceProofsPrefix).Get(context.Background())
	if err != nil {
		return nil, fmt.Errorf("retrieving indexing service proofs: %w", err)
	}
	if len(indexingServiceProofs) == 0 {
		return nil, ErrIndexingServiceProofsMissing
	}
	ranLinkIndex := NewDynamoRanLinkIndex(cfg.Config, cfg.RanLinkIndexTableName)
	s3ReceiptStore := NewS3Store(cfg.Config, cfg.ReceiptStoreBucket, cfg.ReceiptStorePrefix)
	receiptStore, err := receiptstore.NewReceiptStore(s3ReceiptStore, ranLinkIndex)
	if err != nil {
		return nil, fmt.Errorf("setting up receipt store: %w", err)
	}
	opts := []storage.Option{
		storage.WithIdentity(cfg.Signer),
		storage.WithBlobstore(blobStore),
		storage.WithAllocationStore(allocationStore),
		storage.WithClaimStore(claimStore),
		storage.WithPublisherStore(publisherStore),
		storage.WithPublicURL(*pubURL),
		storage.WithPublisherDirectAnnounce(*announceURL),
		storage.WithPublisherAnnounceAddr(cfg.IPNIPublisherAnnounceAddress),
		storage.WithPublisherIndexingServiceConfig(indexingServiceDID, *indexingServiceURL),
		storage.WithPublisherIndexingServiceProof(indexingServiceProofs...),
		storage.WithReceiptStore(receiptStore),
	}

	if cfg.SQSPDPPieceAggregatorURL != "" && cfg.CurioURL != "" {
		pdp, err := NewPDP(cfg)
		if err != nil {
			return nil, fmt.Errorf("setting up PDP: %w", err)
		}

		opts = append(opts, storage.WithPDPConfig(storage.PDPConfig{PDPService: pdp}))
	}
	return storage.New(opts...)
}
