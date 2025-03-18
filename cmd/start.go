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
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	ucanserver "github.com/storacha/go-ucanto/server"
	"github.com/storacha/storage/cmd/enum"
	"github.com/storacha/storage/pkg/presets"
	"github.com/storacha/storage/pkg/principalresolver"
	"github.com/storacha/storage/pkg/server"
	"github.com/storacha/storage/pkg/service/storage"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/urfave/cli/v2"
)

var StartCmd = &cli.Command{
	Name:  "start",
	Usage: "Start the storage node daemon.",
	Flags: []cli.Flag{
		KeyFileFlag,
		CurioURLFlag,
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Value:   3000,
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
		ProofSetFlag,
		&cli.StringFlag{
			Name:    "indexing-service-proof",
			Usage:   "A delegation that allows the node to cache claims with the indexing service.",
			EnvVars: []string{"STORAGE_INDEXING_SERVICE_PROOF"},
		},
	},
	Action: func(cCtx *cli.Context) error {
		id, err := PrincipalSignerFromFile(cCtx.String("key-file"))
		if err != nil {
			return err
		}

		dataDir := cCtx.String("data-dir")
		if dataDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting user home directory: %w", err)
			}

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
		receiptDir, err := mkdirp(dataDir, "receipt")
		if err != nil {
			return err
		}
		receiptDs, err := leveldb.NewDatastore(receiptDir, nil)
		if err != nil {
			return err
		}

		var pdpConfig *storage.PDPConfig
		curioURLStr := cCtx.String("curio-url")
		if curioURLStr != "" {
			curioURL, err := url.Parse(curioURLStr)
			if err != nil {
				return fmt.Errorf("parsing curio URL: %w", err)
			}
			if !cCtx.IsSet("pdp-proofset") {
				return errors.New("pdp-proofset must be set if curio is used")
			}
			proofSet := cCtx.Int64("pdp-proofset")
			pdpDir, err := mkdirp(dataDir, "pdp")
			if err != nil {
				return err
			}
			pdpDs, err := leveldb.NewDatastore(pdpDir, nil)
			if err != nil {
				return err
			}
			pdpConfig = &storage.PDPConfig{
				PDPDatastore:  pdpDs,
				CurioEndpoint: curioURL,
				ProofSet:      uint64(proofSet),
			}
		}

		port := cCtx.Int("port")
		pubURLstr := cCtx.String("public-url")
		if pubURLstr == "" {
			pubURLstr = fmt.Sprintf("http://localhost:%d", port)
			log.Errorf("Public URL is not configured, using: %s", pubURLstr)
		}
		pubURL, err := url.Parse(pubURLstr)
		if err != nil {
			return fmt.Errorf("parsing public URL: %w", err)
		}

		announceURL := *presets.AnnounceURL
		if os.Getenv("STORAGE_ANNOUNCE_URL") != "" {
			u, err := url.Parse(os.Getenv("STORAGE_ANNOUNCE_URL"))
			if err != nil {
				return fmt.Errorf("parsing announce URL: %w", err)
			}
			announceURL = *u
		}

		indexingServiceDID := presets.IndexingServiceDID
		if os.Getenv("STORAGE_INDEXING_SERVICE_DID") != "" {
			d, err := did.Parse(os.Getenv("STORAGE_INDEXING_SERVICE_DID"))
			if err != nil {
				return fmt.Errorf("parsing indexing service DID: %w", err)
			}
			indexingServiceDID = d
		}

		indexingServiceURL := *presets.IndexingServiceURL
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
		if pdpConfig != nil {
			opts = append(opts, storage.WithPDPConfig(*pdpConfig))
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

		principalMapping := presets.PrincipalMapping
		if os.Getenv("STORAGE_PRINCIPAL_MAPPING") != "" {
			var pm map[string]string
			err := json.Unmarshal([]byte(os.Getenv("STORAGE_PRINCIPAL_MAPPING")), &pm)
			if err != nil {
				return fmt.Errorf("parsing principal mapping: %w", err)
			}
			principalMapping = pm
		}
		presolv, err := principalresolver.New(principalMapping)
		if err != nil {
			return fmt.Errorf("creating principal resolver: %w", err)
		}

		go func() {
			time.Sleep(time.Millisecond * 50)
			if err == nil {
				PrintHero(id.DID())
			}
		}()

		err = server.ListenAndServe(
			fmt.Sprintf(":%d", cCtx.Int("port")),
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
