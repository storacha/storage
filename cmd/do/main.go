package main

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	s3ds "github.com/ipfs/go-ds-s3"
	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/did"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	ucanserver "github.com/storacha/go-ucanto/server"
	"github.com/storacha/storage/pkg/principalresolver"
	"github.com/storacha/storage/pkg/server"
	"github.com/storacha/storage/pkg/service/storage"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/urfave/cli/v2"
)

var log = logging.Logger("cmd")

var (
	AnnounceURL, _        = url.Parse("https://cid.contact/announce")
	IndexingServiceDID, _ = did.Parse("did:web:indexer.storacha.network")
	IndexingServiceURL, _ = url.Parse("https://indexer.storacha.network")
	PrincipalMapping      = map[string]string{
		"did:web:staging.upload.storacha.network": "did:key:z6MkqVThfb3PVdgT5yxumxjFFjoQ2vWd26VUQKByPuSB9N91",
		"did:web:upload.storacha.network":         "did:key:z6MkmbbLigYdv5EuU9tJMDXXUudbySwVNeHNqhQGJs7ALUsF",
	}
)

func main() {
	app := &cli.App{
		Name:  "storage",
		Usage: "Manage running a storage node.",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start the storage node daemon.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "private-key",
						Aliases:  []string{"s"},
						Usage:    "Multibase base64 encoded private key identity for the node.",
						EnvVars:  []string{"STORAGE_PRIVATE_KEY"},
						Required: true,
					},
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   3000,
						Usage:   "Port to bind the server to.",
						EnvVars: []string{"STORAGE_PORT"},
					},
					&cli.StringFlag{
						Name:    "public-url",
						Aliases: []string{"u"},
						Usage:   "URL the node is publically accessible at.",
						EnvVars: []string{"STORAGE_PUBLIC_URL"},
					},
					&cli.StringFlag{
						Name:    "indexing-service-proof",
						Usage:   "A delegation that allows the node to cache claims with the indexing service.",
						EnvVars: []string{"STORAGE_INDEXING_SERVICE_PROOF"},
					},
					&cli.StringFlag{
						Name:     "s3-region",
						Usage:    "Bucket region.",
						EnvVars:  []string{"STORAGE_S3_REGION"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "s3-endpoint",
						Usage:    "Bucket edpoint.",
						EnvVars:  []string{"STORAGE_S3_ENDPOINT"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "s3-access-key",
						Usage:    "Access key.",
						EnvVars:  []string{"STORAGE_S3_ACCESS_KEY"},
						Required: true,
					},
					&cli.StringFlag{
						Name:     "s3-secret-key",
						Usage:    "Secret key.",
						EnvVars:  []string{"STORAGE_S3_SECRET_KEY"},
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					var err error
					port := cCtx.Int("port")

					pkstr := cCtx.String("private-key")
					if pkstr == "" {
						signer, err := ed25519.Generate()
						if err != nil {
							return fmt.Errorf("generating ed25519 key: %w", err)
						}
						log.Errorf("Server ID is not configured, generated one for you: %s", signer.DID().String())
						pkstr, err = ed25519.Format(signer)
						if err != nil {
							return fmt.Errorf("formatting ed25519 key: %w", err)
						}
					}

					id, err := ed25519.Parse(pkstr)
					if err != nil {
						return fmt.Errorf("parsing private key: %w", err)
					}

					dstore, err := s3ds.NewS3Datastore(s3ds.Config{
						Region:         cCtx.String("s3-region"),
						RegionEndpoint: cCtx.String("s3-endpoint"),
						AccessKey:      cCtx.String("s3-access-key"),
						SecretKey:      cCtx.String("s3-secret-key"),
					})
					if err != nil {
						return fmt.Errorf("creating S3 bucket: %w", err)
					}

					blobStore := blobstore.NewDsBlobstore(namespace.Wrap(dstore, datastore.NewKey("blob")))
					allocDs := namespace.Wrap(dstore, datastore.NewKey("allocation"))
					claimDs := namespace.Wrap(dstore, datastore.NewKey("claim"))
					publisherDs := namespace.Wrap(dstore, datastore.NewKey("publisher"))
					receiptDs := namespace.Wrap(dstore, datastore.NewKey("receipt"))

					pubURLstr := cCtx.String("public-url")
					if pubURLstr == "" {
						pubURLstr = fmt.Sprintf("http://localhost:%d", port)
						log.Errorf("Public URL is not configured, using: %s", pubURLstr)
					}
					pubURL, err := url.Parse(pubURLstr)
					if err != nil {
						return fmt.Errorf("parsing public URL: %w", err)
					}

					announceURL := *AnnounceURL
					if os.Getenv("STORAGE_ANNOUNCE_URL") != "" {
						u, err := url.Parse(os.Getenv("STORAGE_ANNOUNCE_URL"))
						if err != nil {
							return fmt.Errorf("parsing announce URL: %w", err)
						}
						announceURL = *u
					}

					indexingServiceDID := IndexingServiceDID
					if os.Getenv("STORAGE_INDEXING_SERVICE_DID") != "" {
						d, err := did.Parse(os.Getenv("STORAGE_INDEXING_SERVICE_DID"))
						if err != nil {
							return fmt.Errorf("parsing indexing service DID: %w", err)
						}
						indexingServiceDID = d
					}

					indexingServiceURL := *IndexingServiceURL
					if os.Getenv("STORAGE_INDEXING_SERVICE_URL") != "" {
						u, err := url.Parse(os.Getenv("STORAGE_INDEXING_SERVICE_URL"))
						if err != nil {
							return fmt.Errorf("parsing indexing service URL: %w", err)
						}
						indexingServiceURL = *u
					}

					var indexingServiceProofs delegation.Proofs
					if cCtx.String("indexing-service-proof") != "" {
						dlg, err := delegation.Parse(cCtx.String("indexing-service-proof"))
						if err != nil {
							return fmt.Errorf("parsing indexing service proof: %w", err)
						}
						indexingServiceProofs = append(indexingServiceProofs, delegation.FromDelegation(dlg))
					}

					opts := []storage.Option{
						storage.WithIdentity(id),
						storage.WithBlobstore(blobStore),
						storage.WithAllocationDatastore(allocDs),
						storage.WithClaimDatastore(claimDs),
						storage.WithPublisherDatastore(publisherDs),
						storage.WithPublicURL(*pubURL),
						storage.WithPublisherDirectAnnounce(announceURL),
						storage.WithPublisherIndexingServiceConfig(indexingServiceDID, indexingServiceURL),
						storage.WithPublisherIndexingServiceProof(indexingServiceProofs...),
						storage.WithReceiptDatastore(receiptDs),
					}
					svc, err := storage.New(opts...)
					if err != nil {
						return fmt.Errorf("creating service instance: %w", err)
					}
					err = svc.Startup()
					if err != nil {
						return fmt.Errorf("starting service: %w", err)
					}

					defer svc.Close(cCtx.Context)

					presolv, err := principalresolver.New(PrincipalMapping)
					if err != nil {
						return fmt.Errorf("creating principal resolver: %w", err)
					}

					go func() {
						time.Sleep(time.Millisecond * 50)
						if err == nil {
							printHero(id.DID())
						}
					}()

					err = server.ListenAndServe(
						fmt.Sprintf(":%d", cCtx.Int("port")),
						svc,
						ucanserver.WithPrincipalResolver(presolv.ResolveDIDKey),
					)
					return err
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func printHero(id did.DID) {
	fmt.Printf(`
 00000000                                                   00                  
00      00    00                                            00                  
 000        000000   00000000   00000  0000000    0000000   00000000    0000000 
    00000     00    00     000  00           00  00     0   00    00         00 
        000   00    00      00  00     00000000  00         00    00    0000000 
000     000   00    00     000  00    000    00  000    00  00    00   00    00 
 000000000    0000   0000000    00     000000000   000000   00    00   000000000

🔥 Storage Node %s
🆔 %s
🚀 Ready!
`, "v0.0.0", id.String())
}
