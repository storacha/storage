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
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/storage/pkg/server"
	"github.com/storacha/storage/pkg/service/storage"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/urfave/cli/v2"
)

var log = logging.Logger("cmd")

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
					},
					&cli.StringFlag{
						Name:    "private-key",
						Aliases: []string{"s"},
						Usage:   "Multibase base64 encoded private key identity for the node.",
					},
					&cli.StringFlag{
						Name:    "data-dir",
						Aliases: []string{"d"},
						Usage:   "Root directory to store data in.",
					},
					&cli.StringFlag{
						Name:    "public-url",
						Aliases: []string{"u"},
						Usage:   "URL the node is publically accessible at.",
					},
				},
				Action: func(cCtx *cli.Context) error {
					var id principal.Signer
					var err error
					if cCtx.String("private-key") == "" {
						id, err = ed25519.Generate()
						if err != nil {
							return fmt.Errorf("generating ed25519 key: %w", err)
						}
						log.Errorf("Server ID is not configured, generated one for you: %s", id.DID().String())
					} else {
						id, err = ed25519.Parse(cCtx.String("private-key"))
						if err != nil {
							return fmt.Errorf("parsing private key: %w", err)
						}
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

					blobStore, err := blobstore.NewFsBlobstore(path.Join(dataDir, "blobs"))
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
						pubURLstr = fmt.Sprintf("http://localhost:%d", cCtx.Int("port"))
						log.Errorf("Public URL is not configured, using: %s", pubURLstr)
					}
					pubURL, err := url.Parse(pubURLstr)
					if err != nil {
						return fmt.Errorf("parsing public URL: %w", err)
					}

					svc, err := storage.New(
						storage.WithIdentity(id),
						storage.WithBlobstore(blobStore),
						storage.WithAllocationDatastore(allocDs),
						storage.WithClaimDatastore(claimDs),
						storage.WithPublisherDatastore(publisherDs),
						storage.WithPublicURL(*pubURL),
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
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return "", fmt.Errorf("creating directory: %s: %w", dir, err)
	}
	return dir, nil
}
