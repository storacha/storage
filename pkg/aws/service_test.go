package aws

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/alanshaw/storetheindex/config"
	"github.com/alanshaw/storetheindex/ingest"
	"github.com/alanshaw/storetheindex/registry"
	httpfind "github.com/alanshaw/storetheindex/server/find"
	httpingest "github.com/alanshaw/storetheindex/server/ingest"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/ipni/go-indexer-core/engine"
	"github.com/ipni/go-indexer-core/store/memory"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/storage/pkg/internal/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcdynamodb "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

func TestConstruct(t *testing.T) {
	id, err := signer.Generate()
	require.NoError(t, err)

	cryptoPrivKey, err := crypto.UnmarshalEd25519PrivateKey(id.Raw())
	require.NoError(t, err)

	peer, err := peer.IDFromPrivateKey(cryptoPrivKey)
	require.NoError(t, err)

	s3Endpoint := createS3(t)
	dynamoEndpoint := createDynamo(t)

	ipniFindURL := testutil.RandomLocalURL(t)
	ipniAnnounceURL := testutil.RandomLocalURL(t)

	closeIPNI := startIPNIService(t, ipniFindURL, ipniAnnounceURL)
	defer closeIPNI()

	cfg := Config{
		Signer: id,
		Config: aws.Config{},
		S3Options: []func(*s3.Options){
			func(o *s3.Options) {
				o.Credentials = credentials.NewStaticCredentialsProvider("minioadmin", "minioadmin", "")
				o.UsePathStyle = true
				o.Region = "us-east-1"
				o.BaseEndpoint = aws.String(s3Endpoint.String())
			},
		},
		DynamoOptions: []func(*dynamodb.Options){
			func(o *dynamodb.Options) {
				o.Credentials = credentials.NewStaticCredentialsProvider("DUMMYIDEXAMPLE", "DUMMYEXAMPLEKEY", "")
				o.Region = "us-east-1"
				o.BaseEndpoint = aws.String(dynamoEndpoint.String())
			},
		},
		ChunkLinksTableName:          fmt.Sprintf("chunk-links-%s", testutil.RandomCID(t)),
		MetadataTableName:            fmt.Sprintf("metadata-%s", testutil.RandomCID(t)),
		IPNIStoreBucket:              fmt.Sprintf("ipni-store-%s", testutil.RandomCID(t)),
		IPNIPublisherAnnounceAddress: fmt.Sprintf("/dns4/%s/tcp/%s/p2p/%s", s3Endpoint.Host, s3Endpoint.Port(), peer),
		BlobsPublicURL:               s3Endpoint.String(),
		ClaimStoreBucket:             fmt.Sprintf("ipni-store-%s", testutil.RandomCID(t)),
		AllocationsTableName:         fmt.Sprintf("allocations-%s", testutil.RandomCID(t)),
		BlobStoreBucket:              fmt.Sprintf("blobs-%s", testutil.RandomCID(t)),
		BlobStorePrefix:              "/blob/",
		AnnounceURL:                  mustGetEnv("IPNI_ENDPOINT"),
		PublicURL:                    mustGetEnv("PUBLIC_URL"),
		IndexingServiceDID:           mustGetEnv("INDEXING_SERVICE_DID"),
		IndexingServiceURL:           mustGetEnv("INDEXING_SERVICE_URL"),
		IndexingServiceProof:         mustGetEnv("INDEXING_SERVICE_PROOF"),
		RanLinkIndexTableName:        mustGetEnv("RAN_LINK_INDEX_TABLE_NAME"),
		ReceiptStoreBucket:           mustGetEnv("RECEIPT_STORE_BUCKET_NAME"),
		ReceiptStorePrefix:           os.Getenv("RECEIPT_STORE_KEY_PREFIX"),
		SQSPDPPieceAggregatorURL:     os.Getenv("PIECE_AGGREGATOR_QUEUE_URL"),
		SQSPDPAggregateSubmitterURL:  os.Getenv("AGGREGATE_SUBMITTER_QUEUE_URL"),
		SQSPDPPieceAccepterURL:       os.Getenv("PIECE_ACCEPTER_QUEUE_URL"),
		PDPProofSet:                  proofSet,
		CurioURL:                     os.Getenv("CURIO_URL"),
		PrincipalMapping:             principalMapping,
	}
	s, err := Construct(cfg)
	require.NoError(t, err)
}

func createDynamo(t *testing.T) *url.URL {
	ctx := context.Background()
	container, err := tcdynamodb.Run(ctx, "amazon/dynamodb-local:latest")
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	endpoint, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	return testutil.Must(url.Parse("http://" + endpoint))(t)
}

func createS3(t *testing.T) *url.URL {
	ctx := context.Background()
	container, err := tcminio.Run(ctx, "minio/minio:latest")
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	endpoint, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	return testutil.Must(url.Parse("http://" + endpoint))(t)
}

func startIPNIService(
	t *testing.T,
	findURL url.URL,
	announceURL url.URL,
) func() {
	indexerCore := engine.New(memory.New())

	reg, err := registry.New(
		context.Background(),
		config.NewDiscovery(),
		dssync.MutexWrap(datastore.NewMapDatastore()),
	)
	require.NoError(t, err)

	p2pHost, err := libp2p.New()
	require.NoError(t, err)

	ingConfig := config.NewIngest()
	ingConfig.PubSubTopic = "/storacha/indexer/ingest/testnet"
	ing, err := ingest.NewIngester(
		ingConfig,
		p2pHost,
		indexerCore,
		reg,
		dssync.MutexWrap(datastore.NewMapDatastore()),
		dssync.MutexWrap(datastore.NewMapDatastore()),
	)
	require.NoError(t, err)

	announceAddr := fmt.Sprintf("%s:%s", announceURL.Hostname(), announceURL.Port())
	ingSvr, err := httpingest.New(announceAddr, indexerCore, ing, reg)
	require.NoError(t, err)

	var ingStartErr error
	go func() {
		ingStartErr = ingSvr.Start()
	}()

	findAddr := fmt.Sprintf("%s:%s", findURL.Hostname(), findURL.Port())
	findSvr, err := httpfind.New(findAddr, indexerCore, reg)
	require.NoError(t, err)

	var findStartErr error
	go func() {
		findStartErr = findSvr.Start()
	}()

	time.Sleep(time.Millisecond * 100)
	require.NoError(t, ingStartErr)
	require.NoError(t, findStartErr)

	return func() {
		ingSvr.Close()
		ing.Close()
		findSvr.Close()
		reg.Close()
		indexerCore.Close()
		p2pHost.Close()
	}
}
