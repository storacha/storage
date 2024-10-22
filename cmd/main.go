package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"time"

	leveldb "github.com/ipfs/go-ds-leveldb"
	logging "github.com/ipfs/go-log/v2"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/did"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
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
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   3000,
						Usage:   "Port to bind the server to.",
						EnvVars: []string{"STORAGE_PORT"},
					},
					&cli.StringFlag{
						Name:    "private-key",
						Aliases: []string{"s"},
						Usage:   "Multibase base64 encoded private key identity for the node.",
						EnvVars: []string{"STORAGE_PRIVATE_KEY"},
					},
					&cli.StringFlag{
						Name:    "data-dir",
						Aliases: []string{"d"},
						Usage:   "Root directory to store data in.",
						EnvVars: []string{"STORAGE_DATA_DIR"},
					},
					&cli.StringFlag{
						Name:    "tmp-dir",
						Aliases: []string{"t"},
						Usage:   "Temporary directory data is uploaded to before being moved to data-dir.",
						EnvVars: []string{"STORAGE_TMP_DIR"},
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

					homeDir, err := os.UserHomeDir()
					if err != nil {
						return fmt.Errorf("getting user home directory: %w", err)
					}

					dataDir := cCtx.String("data-dir")
					if dataDir == "" {
						dir, err := mkdirp(homeDir, ".storacha")
						if err != nil {
							return err
						}
						log.Errorf("Data directory is not configured, using default: %s", dir)
						dataDir = dir
					}

					tmpDir := cCtx.String("tmp-dir")
					if tmpDir == "" {
						dir, err := mkdirp(path.Join(os.TempDir(), "storage"))
						if err != nil {
							return err
						}
						log.Warnf("Tmp directory is not configured, using default: %s", dir)
						tmpDir = dir
					}

					blobStore, err := blobstore.NewFsBlobstore(path.Join(dataDir, "blobs"), path.Join(tmpDir, "blobs"))
					if err != nil {
						return fmt.Errorf("creating blob storage: %w", err)
					}

					allocsDir, err := mkdirp(dataDir, "allocation")
					if err != nil {
						return err
					}
					allocDs, err := leveldb.NewDatastore(allocsDir, nil)
					if err != nil {
						return err
					}
					claimsDir, err := mkdirp(dataDir, "claim")
					if err != nil {
						return err
					}
					claimDs, err := leveldb.NewDatastore(claimsDir, nil)
					if err != nil {
						return err
					}
					publisherDir, err := mkdirp(dataDir, "publisher")
					if err != nil {
						return err
					}
					publisherDs, err := leveldb.NewDatastore(publisherDir, nil)
					if err != nil {
						return err
					}

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
					if os.Getenv("ANNOUNCE_URL") != "" {
						u, err := url.Parse(os.Getenv("ANNOUNCE_URL"))
						if err != nil {
							return fmt.Errorf("parsing announce URL: %w", err)
						}
						announceURL = *u
					}

					indexingServiceDID := IndexingServiceDID
					if os.Getenv("INDEXING_SERVICE_DID") != "" {
						d, err := did.Parse(os.Getenv("INDEXING_SERVICE_DID"))
						if err != nil {
							return fmt.Errorf("parsing indexing service DID: %w", err)
						}
						indexingServiceDID = d
					}

					indexingServiceURL := *IndexingServiceURL
					if os.Getenv("INDEXING_SERVICE_URL") != "" {
						u, err := url.Parse(os.Getenv("INDEXING_SERVICE_URL"))
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

					svc, err := storage.New(
						storage.WithIdentity(id),
						storage.WithBlobstore(blobStore),
						storage.WithAllocationDatastore(allocDs),
						storage.WithClaimDatastore(claimDs),
						storage.WithPublisherDatastore(publisherDs),
						storage.WithPublicURL(*pubURL),
						storage.WithPublisherDirectAnnounce(announceURL),
						storage.WithPublisherIndexingServiceConfig(indexingServiceDID, indexingServiceURL),
						storage.WithPublisherIndexingServiceProof(indexingServiceProofs...),
					)
					if err != nil {
						return fmt.Errorf("creating service instance: %w", err)
					}
					defer svc.Close()

					go func() {
						time.Sleep(time.Millisecond * 50)
						if err == nil {
							printHero(id.DID())
						}
					}()
					err = server.ListenAndServe(fmt.Sprintf(":%d", cCtx.Int("port")), svc)
					return err
				},
			},
			{
				Name:    "identity",
				Aliases: []string{"id"},
				Usage:   "Identity tools.",
				Subcommands: []*cli.Command{
					{
						Name:    "generate",
						Aliases: []string{"gen"},
						Usage:   "Generate a new decentralized identity.",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "json",
								Usage: "output JSON",
							},
						},
						Action: func(cCtx *cli.Context) error {
							signer, err := ed25519.Generate()
							if err != nil {
								return fmt.Errorf("generating ed25519 key: %w", err)
							}
							did := signer.DID().String()
							key, err := ed25519.Format(signer)
							if err != nil {
								return fmt.Errorf("formatting ed25519 key: %w", err)
							}
							if cCtx.Bool("json") {
								out, err := json.Marshal(struct {
									DID string `json:"did"`
									Key string `json:"key"`
								}{did, key})
								if err != nil {
									return fmt.Errorf("marshaling JSON: %w", err)
								}
								fmt.Println(string(out))
							} else {
								fmt.Printf("# %s\n", did)
								fmt.Println(key)
							}
							return nil
						},
					},
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

func mkdirp(dirpath ...string) (string, error) {
	dir := path.Join(dirpath...)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return "", fmt.Errorf("creating directory: %s: %w", dir, err)
	}
	return dir, nil
}
