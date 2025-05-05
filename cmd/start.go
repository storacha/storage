package cmd

import (
	crypto_ed25519 "crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/ipni/go-libipni/maurl"
	"github.com/multiformats/go-multiaddr"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	ucanserver "github.com/storacha/go-ucanto/server"
	"github.com/urfave/cli/v2"

	"github.com/storacha/storage/cmd/enum"
	"github.com/storacha/storage/pkg/config"
	"github.com/storacha/storage/pkg/pdp/aggregator/jobqueue"
	"github.com/storacha/storage/pkg/principalresolver"
	"github.com/storacha/storage/pkg/server"
	"github.com/storacha/storage/pkg/service/storage"
	"github.com/storacha/storage/pkg/store/blobstore"
)

var StartCmd = &cli.Command{
	Name:  "start",
	Usage: "Start the storage node daemon.",
	Flags: []cli.Flag{
		CurioURLFlag,
		ProofSetFlag,
		&cli.PathFlag{
			Name:      "key-file",
			Usage:     "Path to a file containing ed25519 private key, typically created by the id gen command.",
			EnvVars:   []string{"STORAGE_PRIVATE_KEY"},
			TakesFile: true,
		},
		&cli.StringFlag{
			Name:    "config",
			Usage:   "Path to configuration file.",
			EnvVars: []string{"STORAGE_CONFIG"},
		},
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Value:   config.DefaultServicePort,
			Usage:   "Port to bind the server to.",
			EnvVars: []string{"STORAGE_PORT"},
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
		&cli.StringFlag{
			Name:    "indexing-service-url",
			Usage:   "URL of the indexing service.",
			EnvVars: []string{"STORAGE_INDEXING_SERVICE_URL"},
		},
		&cli.StringFlag{
			Name:    "indexing-service-did",
			Usage:   "DID of the indexing service.",
			EnvVars: []string{"STORAGE_INDEXING_SERVICE_DID"},
		},
		&cli.StringFlag{
			Name:    "announce-url",
			Usage:   "URL to announce storage advertisements.",
			EnvVars: []string{"STORAGE_ANNOUNCE_URL"},
		},
		&cli.StringFlag{
			Name:    "upload-service-url",
			Usage:   "URL of the upload service.",
			EnvVars: []string{"STORAGE_UPLOAD_SERVICE_URL"},
		},
		&cli.StringFlag{
			Name:    "upload-service-did",
			Usage:   "DID of the upload service.",
			EnvVars: []string{"STORAGE_UPLOAD_SERVICE_DID"},
		},
	},
	Action: func(cCtx *cli.Context) error {
		cfg, err := config.LoadConfig(cCtx)
		if err != nil {
			return err
		}

		// load identity from key file
		id, err := PrincipalSignerFromFile(cfg.Core.KeyFilePath)
		if err != nil {
			return err
		}

		// Create file-based blob storage
		blobStore, err := blobstore.NewFsBlobstore(
			path.Join(cfg.Directories.DataDir, "blobs"),
			path.Join(cfg.Directories.TempDir, "blobs"),
		)
		if err != nil {
			return fmt.Errorf("creating blob storage: %w", err)
		}

		// Set up datastores
		allocsDir, err := mkdirp(cfg.Directories.DataDir, "allocation")
		if err != nil {
			return err
		}
		allocDs, err := leveldb.NewDatastore(allocsDir, nil)
		if err != nil {
			return err
		}

		claimsDir, err := mkdirp(cfg.Directories.DataDir, "claim")
		if err != nil {
			return err
		}
		claimDs, err := leveldb.NewDatastore(claimsDir, nil)
		if err != nil {
			return err
		}

		publisherDir, err := mkdirp(cfg.Directories.DataDir, "publisher")
		if err != nil {
			return err
		}
		publisherDs, err := leveldb.NewDatastore(publisherDir, nil)
		if err != nil {
			return err
		}

		receiptDir, err := mkdirp(cfg.Directories.DataDir, "receipt")
		if err != nil {
			return err
		}
		receiptDs, err := leveldb.NewDatastore(receiptDir, nil)
		if err != nil {
			return err
		}

		// Parse necessary URLs
		pubURL, err := url.Parse(cfg.Core.PublicURL)
		if err != nil {
			return fmt.Errorf("parsing public URL: %w", err)
		}

		announceURL, err := url.Parse(cfg.Indexing.AnnounceURL)
		if err != nil {
			return fmt.Errorf("parsing announce URL: %w", err)
		}

		indexingServiceURL, err := url.Parse(cfg.Indexing.ServiceURL)
		if err != nil {
			return fmt.Errorf("parsing indexing service URL: %w", err)
		}

		uploadServiceURL, err := url.Parse(cfg.Upload.ServiceURL)
		if err != nil {
			return fmt.Errorf("parsing upload service URL: %w", err)
		}

		// Parse DIDs
		indexingServiceDID, err := did.Parse(cfg.Indexing.ServiceDID)
		if err != nil {
			return fmt.Errorf("parsing indexing service DID: %w", err)
		}

		uploadServiceDID, err := did.Parse(cfg.Upload.ServiceDID)
		if err != nil {
			return fmt.Errorf("parsing upload service DID: %w", err)
		}

		// Parse delegation proofs if present
		var indexingServiceProofs delegation.Proofs
		if cfg.Indexing.StorageProof != "" {
			dlg, err := delegation.Parse(cfg.Indexing.StorageProof)
			if err != nil {
				return fmt.Errorf("parsing indexing service proof: %w", err)
			}
			indexingServiceProofs = append(indexingServiceProofs, delegation.FromDelegation(dlg))
		}

		// Configure PDP if enabled
		var pdpConfig *storage.PDPConfig
		var blobAddr multiaddr.Multiaddr
		if cfg.PDP != nil && cfg.PDP.ServerURL != "" {
			curioURL, err := url.Parse(cfg.PDP.ServerURL)
			if err != nil {
				return fmt.Errorf("parsing curio URL: %w", err)
			}

			if cfg.PDP.ProofSet == 0 {
				return errors.New("pdp-proofset must be set if curio is used")
			}

			pdpDir, err := mkdirp(cfg.Directories.DataDir, "pdp")
			if err != nil {
				return err
			}
			pdpDs, err := leveldb.NewDatastore(pdpDir, nil)
			if err != nil {
				return err
			}

			pdpDB, err := jobqueue.NewInMemoryDB()
			if err != nil {
				return err
			}

			pdpConfig = &storage.PDPConfig{
				PDPDatastore:  pdpDs,
				CurioEndpoint: curioURL,
				ProofSet:      cfg.PDP.ProofSet,
				Database:      pdpDB,
			}
			curioAddr, err := maurl.FromURL(curioURL)
			if err != nil {
				return err
			}
			pieceAddr, err := multiaddr.NewMultiaddr("/http-path/" + url.PathEscape("piece/{blobCID}"))
			if err != nil {
				return err
			}
			blobAddr = multiaddr.Join(curioAddr, pieceAddr)
		}

		// Configure storage service options
		opts := []storage.Option{
			storage.WithIdentity(id),
			storage.WithBlobstore(blobStore),
			storage.WithAllocationDatastore(allocDs),
			storage.WithClaimDatastore(claimDs),
			storage.WithPublisherDatastore(publisherDs),
			storage.WithPublicURL(*pubURL),
			storage.WithPublisherDirectAnnounce(*announceURL),
			storage.WithUploadServiceConfig(uploadServiceDID, *uploadServiceURL),
			storage.WithPublisherIndexingServiceConfig(indexingServiceDID, *indexingServiceURL),
			storage.WithPublisherIndexingServiceProof(indexingServiceProofs...),
			storage.WithReceiptDatastore(receiptDs),
		}

		if pdpConfig != nil {
			opts = append(opts, storage.WithPDPConfig(*pdpConfig))
		}
		if blobAddr != nil {
			opts = append(opts, storage.WithPublisherBlobAddress(blobAddr))
		}

		// Create and start the service
		svc, err := storage.New(opts...)
		if err != nil {
			return fmt.Errorf("creating service instance: %w", err)
		}

		err = svc.Startup(cCtx.Context)
		if err != nil {
			return fmt.Errorf("starting service: %w", err)
		}

		defer svc.Close(cCtx.Context)

		// Configure principal resolver
		presolv, err := principalresolver.New(cfg.Principals)
		if err != nil {
			return fmt.Errorf("creating principal resolver: %w", err)
		}

		// Display the service ID
		go func() {
			time.Sleep(time.Millisecond * 50)
			if err == nil {
				PrintHero(id.DID())
			}
		}()

		// Start the HTTP server
		err = server.ListenAndServe(
			fmt.Sprintf(":%d", cfg.Core.ServerPort),
			svc,
			ucanserver.WithPrincipalResolver(presolv.ResolveDIDKey),
		)
		return err
	},
}

func PrincipalSignerFromFile(path string) (principal.Signer, error) {
	// open the file
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	// ensure its either json or pem
	baseName := strings.Trim(filepath.Ext(path), ".")
	typ := enum.ParseKeyFormat(strings.ToUpper(baseName))
	if !typ.IsValid() {
		return nil, fmt.Errorf("unsupported file type: %s, expected .json or .pem", baseName)
	}

	// read accordingly
	switch typ {
	case enum.KeyFormats.PEM:
		return readPrivateKeyFromPEM(f)
	case enum.KeyFormats.JSON:
		return readPrivateKeyFromJSON(f)
	}

	return nil, fmt.Errorf("unsupported file type: %s, expected .json or .pem", baseName)
}

func readPrivateKeyFromJSON(f io.Reader) (principal.Signer, error) {
	jsonData, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading private key file: %w", err)
	}

	var key struct {
		DID string `json:"did"`
		Key string `json:"key"`
	}
	if err := json.Unmarshal(jsonData, &key); err != nil {
		return nil, fmt.Errorf("unmarshaling private key file to json: %w", err)
	}

	return ed25519.Parse(key.Key)
}

func readPrivateKeyFromPEM(f io.Reader) (principal.Signer, error) {
	pemData, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}

	var privateKey *crypto_ed25519.PrivateKey
	rest := pemData

	// Loop until no more blocks
	for {
		block, remaining := pem.Decode(rest)
		if block == nil {
			// No more PEM blocks
			break
		}
		rest = remaining

		// Look for "PRIVATE KEY"
		if block.Type == "PRIVATE KEY" {
			parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse PKCS#8 private key: %w", err)
			}

			// We expect a ed25519 private key, cast it
			key, ok := parsedKey.(crypto_ed25519.PrivateKey)
			if !ok {
				return nil, fmt.Errorf("the parsed key is not an ED25519 private key")
			}
			privateKey = &key
			break
		}
	}

	if privateKey == nil {
		return nil, fmt.Errorf("could not find a PRIVATE KEY block in the PEM file")
	}
	return ed25519.FromRaw(*privateKey)
}
