package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/storacha/go-metadata"
	"github.com/storacha/go-piece/pkg/piece"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/go-ucanto/principal/signer"

	"github.com/storacha/ipni-publisher/pkg/store"
	"github.com/storacha/storage/pkg/pdp"
	"github.com/storacha/storage/pkg/pdp/aggregator"
	"github.com/storacha/storage/pkg/pdp/curio"
	"github.com/storacha/storage/pkg/pdp/pieceadder"
	"github.com/storacha/storage/pkg/pdp/piecefinder"
	"github.com/storacha/storage/pkg/presets"
	"github.com/storacha/storage/pkg/service/storage"
	"github.com/storacha/storage/pkg/store/delegationstore"
	"github.com/storacha/storage/pkg/store/receiptstore"
)

// ErrMissingSecret means that the value returned from Secrets was empty
var ErrMissingSecret = errors.New("missing value for secret")

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
	Config                         aws.Config
	S3Options                      []func(*s3.Options)
	DynamoOptions                  []func(*dynamodb.Options)
	AllocationsTableName           string
	BlobStoreBucketEndpoint        string
	BlobStoreBucketRegion          string
	BlobStoreBucketAccessKeyID     string
	BlobStoreBucketSecretAccessKey string
	BlobStoreBucket                string
	BlobStorePrefix                string
	AggregatesBucket               string
	AggregatesPrefix               string
	BufferBucket                   string
	BufferPrefix                   string
	ChunkLinksTableName            string
	MetadataTableName              string
	IPNIStoreBucket                string
	IPNIStorePrefix                string
	ClaimStoreBucket               string
	ClaimStorePrefix               string
	PublicURL                      string
	AnnounceURL                    string
	IndexingServiceDID             string
	IndexingServiceURL             string
	IndexingServiceProof           string
	IPNIPublisherAnnounceAddress   string
	BlobsPublicURL                 string
	RanLinkIndexTableName          string
	ReceiptStoreBucket             string
	ReceiptStorePrefix             string
	SQSPDPPieceAggregatorURL       string
	SQSPDPAggregateSubmitterURL    string
	SQSPDPPieceAccepterURL         string
	PDPProofSet                    uint64
	CurioURL                       string
	PrincipalMapping               map[string]string
	principal.Signer
}

func mustGetSSMParams(ctx context.Context, client *ssm.Client, names ...string) map[string]string {
	response, err := client.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          names,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		panic(fmt.Errorf("retrieving SSM parameters: %w", err))
	}
	params := map[string]string{}
	for _, name := range names {
		value := ""
		for _, p := range response.Parameters {
			if *p.Name == name {
				value = *p.Value
				break
			}
		}
		if value == "" {
			panic(ErrMissingSecret)
		}
		params[name] = value
	}
	return params
}

// FromEnv constructs the AWS Configuration from the environment
func FromEnv(ctx context.Context) Config {
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("loading aws default config: %w", err))
	}

	ssmClient := ssm.NewFromConfig(awsConfig)
	secretNames := []string{mustGetEnv("PRIVATE_KEY")}
	for _, n := range []string{
		"BLOB_STORE_BUCKET_ACCESS_KEY_ID",
		"BLOB_STORE_BUCKET_SECRET_ACCESS_KEY",
	} {
		if os.Getenv(n) != "" {
			secretNames = append(secretNames, os.Getenv(n))
		}
	}
	secrets := mustGetSSMParams(ctx, ssmClient, secretNames...)

	id, err := ed25519.Parse(secrets[mustGetEnv("PRIVATE_KEY")])
	if err != nil {
		panic(fmt.Errorf("parsing private key: %s", err))
	}

	if len(os.Getenv("DID")) != 0 {
		d, err := did.Parse(os.Getenv("DID"))
		if err != nil {
			panic(fmt.Errorf("parsing DID: %w", err))
		}
		id, err = signer.Wrap(id, d)
		if err != nil {
			panic(fmt.Errorf("wrapping server DID: %w", err))
		}
	}

	cryptoPrivKey, err := crypto.UnmarshalEd25519PrivateKey(id.Raw())
	if err != nil {
		panic(fmt.Errorf("unmarshaling private key: %w", err))
	}

	ipniStoreKeyPrefix := os.Getenv("IPNI_STORE_KEY_PREFIX")
	if len(ipniStoreKeyPrefix) == 0 {
		ipniStoreKeyPrefix = "ipni/v1/ad/"
	}

	peer, err := peer.IDFromPrivateKey(cryptoPrivKey)
	if err != nil {
		panic(fmt.Errorf("parsing private key to peer: %w", err))
	}

	ipniPublisherAnnounceAddress := fmt.Sprintf("/dns4/%s/tcp/443/https/p2p/%s", mustGetEnv("IPNI_STORE_BUCKET_REGIONAL_DOMAIN"), peer.String())
	blobsPublicURL := "https://" + mustGetEnv("BLOB_STORE_BUCKET_REGIONAL_DOMAIN")
	proofSetString := os.Getenv("PDP_PROOFSET")
	var proofSet uint64
	if len(proofSetString) > 0 {
		proofSet, err = strconv.ParseUint(proofSetString, 10, 64)
		if err != nil {
			panic(fmt.Errorf("parsing pdp proofset: %w", err))
		}
	}

	var principalMapping map[string]string
	if os.Getenv("PRINCIPAL_MAPPING") != "" {
		principalMapping = map[string]string{}
		maps.Copy(principalMapping, presets.PrincipalMapping)
		var pm map[string]string
		err := json.Unmarshal([]byte(os.Getenv("PRINCIPAL_MAPPING")), &pm)
		if err != nil {
			panic(fmt.Errorf("parsing principal mapping: %w", err))
		}
		maps.Copy(principalMapping, pm)
	} else {
		principalMapping = presets.PrincipalMapping
	}

	return Config{
		Config:                         awsConfig,
		Signer:                         id,
		ChunkLinksTableName:            mustGetEnv("CHUNK_LINKS_TABLE_NAME"),
		MetadataTableName:              mustGetEnv("METADATA_TABLE_NAME"),
		IPNIStoreBucket:                mustGetEnv("IPNI_STORE_BUCKET_NAME"),
		IPNIStorePrefix:                ipniStoreKeyPrefix,
		IPNIPublisherAnnounceAddress:   ipniPublisherAnnounceAddress,
		BlobsPublicURL:                 blobsPublicURL,
		ClaimStoreBucket:               mustGetEnv("CLAIM_STORE_BUCKET_NAME"),
		ClaimStorePrefix:               os.Getenv("CLAIM_STORE_KEY_REFIX"),
		AllocationsTableName:           mustGetEnv("ALLOCATIONS_TABLE_NAME"),
		BlobStoreBucketEndpoint:        os.Getenv("BLOB_STORE_BUCKET_ENDPOINT"),
		BlobStoreBucketRegion:          os.Getenv("BLOB_STORE_BUCKET_REGION"),
		BlobStoreBucketAccessKeyID:     secrets[os.Getenv("BLOB_STORE_BUCKET_ACCESS_KEY_ID")],
		BlobStoreBucketSecretAccessKey: secrets[os.Getenv("BLOB_STORE_BUCKET_SECRET_ACCESS_KEY")],
		BlobStoreBucket:                mustGetEnv("BLOB_STORE_BUCKET_NAME"),
		BlobStorePrefix:                os.Getenv("BLOB_STORE_KEY_PREFIX"),
		BufferBucket:                   os.Getenv("BUFFER_BUCKET_NAME"),
		BufferPrefix:                   os.Getenv("BUFFER_KEY_PREFIX"),
		AggregatesBucket:               os.Getenv("AGGREGATES_BUCKET_NAME"),
		AggregatesPrefix:               os.Getenv("AGGREGATES_KEY_PREFIX"),
		AnnounceURL:                    mustGetEnv("IPNI_ENDPOINT"),
		PublicURL:                      mustGetEnv("PUBLIC_URL"),
		IndexingServiceDID:             mustGetEnv("INDEXING_SERVICE_DID"),
		IndexingServiceURL:             mustGetEnv("INDEXING_SERVICE_URL"),
		IndexingServiceProof:           mustGetEnv("INDEXING_SERVICE_PROOF"),
		RanLinkIndexTableName:          mustGetEnv("RAN_LINK_INDEX_TABLE_NAME"),
		ReceiptStoreBucket:             mustGetEnv("RECEIPT_STORE_BUCKET_NAME"),
		ReceiptStorePrefix:             os.Getenv("RECEIPT_STORE_KEY_PREFIX"),
		SQSPDPPieceAggregatorURL:       os.Getenv("PIECE_AGGREGATOR_QUEUE_URL"),
		SQSPDPAggregateSubmitterURL:    os.Getenv("AGGREGATE_SUBMITTER_QUEUE_URL"),
		SQSPDPPieceAccepterURL:         os.Getenv("PIECE_ACCEPTER_QUEUE_URL"),
		PDPProofSet:                    proofSet,
		CurioURL:                       os.Getenv("CURIO_URL"),
		PrincipalMapping:               principalMapping,
	}
}

func Construct(cfg Config) (storage.Service, error) {
	blobStoreOpts := cfg.S3Options
	if cfg.BlobStoreBucketAccessKeyID != "" && cfg.BlobStoreBucketSecretAccessKey != "" {
		blobStoreOpts = append(blobStoreOpts, func(opts *s3.Options) {
			opts.Region = cfg.BlobStoreBucketRegion
			opts.Credentials = credentials.NewStaticCredentialsProvider(
				cfg.BlobStoreBucketAccessKeyID,
				cfg.BlobStoreBucketSecretAccessKey,
				"",
			)
			if cfg.BlobStoreBucketEndpoint != "" {
				opts.BaseEndpoint = &cfg.BlobStoreBucketEndpoint
				opts.UsePathStyle = true
			}
		})
	}
	blobStore := NewS3BlobStore(cfg.Config, cfg.BlobStoreBucket, cfg.BlobStorePrefix, blobStoreOpts...)
	allocationStore := NewDynamoAllocationStore(cfg.Config, cfg.AllocationsTableName, cfg.DynamoOptions...)
	claimStore, err := delegationstore.NewDelegationStore(NewS3Store(cfg.Config, cfg.ClaimStoreBucket, cfg.ClaimStorePrefix, cfg.S3Options...))
	if err != nil {
		return nil, fmt.Errorf("constructing claim store: %w", err)
	}
	ipniStore := NewS3Store(cfg.Config, cfg.IPNIStoreBucket, cfg.IPNIStorePrefix, cfg.S3Options...)
	chunkLinksTable := NewDynamoProviderContextTable(cfg.Config, cfg.ChunkLinksTableName, cfg.DynamoOptions...)
	metadataTable := NewDynamoProviderContextTable(cfg.Config, cfg.MetadataTableName, cfg.DynamoOptions...)
	publisherStore := store.NewPublisherStore(ipniStore, chunkLinksTable, metadataTable, store.WithMetadataContext(metadata.MetadataContext))
	pubURL, err := url.Parse(cfg.PublicURL)
	if err != nil {
		return nil, fmt.Errorf("parsing public url: %w", err)
	}
	blobsPublicURL, err := url.Parse(cfg.BlobsPublicURL)
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
	var indexingServiceProofs delegation.Proofs
	proof, err := delegation.Parse(cfg.IndexingServiceProof)
	if err != nil {
		return nil, fmt.Errorf("parsing indexing service proof")
	}
	indexingServiceProofs = append(indexingServiceProofs, delegation.FromDelegation(proof))
	if len(indexingServiceProofs) == 0 {
		return nil, ErrIndexingServiceProofsMissing
	}
	ranLinkIndex := NewDynamoRanLinkIndex(cfg.Config, cfg.RanLinkIndexTableName, cfg.DynamoOptions...)
	s3ReceiptStore := NewS3Store(cfg.Config, cfg.ReceiptStoreBucket, cfg.ReceiptStorePrefix, cfg.S3Options...)
	receiptStore, err := receiptstore.NewReceiptStore(s3ReceiptStore, ranLinkIndex)
	if err != nil {
		return nil, fmt.Errorf("setting up receipt store: %w", err)
	}
	announceAddr, err := multiaddr.NewMultiaddr(cfg.IPNIPublisherAnnounceAddress)
	if err != nil {
		return nil, fmt.Errorf("parsing announce multiaddr: %w", err)
	}
	opts := []storage.Option{
		storage.WithIdentity(cfg.Signer),
		storage.WithBlobstore(blobStore),
		storage.WithAllocationStore(allocationStore),
		storage.WithClaimStore(claimStore),
		storage.WithPublisherStore(publisherStore),
		storage.WithPublicURL(*pubURL),
		storage.WithPublisherDirectAnnounce(*announceURL),
		storage.WithPublisherAnnounceAddress(announceAddr),
		storage.WithPublisherIndexingServiceConfig(indexingServiceDID, *indexingServiceURL),
		storage.WithPublisherIndexingServiceProof(indexingServiceProofs...),
		storage.WithReceiptStore(receiptStore),
		storage.WithBlobsPublicURL(*blobsPublicURL),
		storage.WithBlobsPresigner(blobStore.PresignClient()),
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
