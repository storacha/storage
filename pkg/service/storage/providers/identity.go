package providers

import (
	crypto_ed25519 "crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"

	logging "github.com/ipfs/go-log"
	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"go.uber.org/fx"

	"github.com/storacha/piri/pkg/config"
)

var log = logging.Logger("storage")

// IdentityParams are the dependencies for creating an identity
type IdentityParams struct {
	fx.In
	Config config.UCANServer
}

// NewIdentity creates or loads a principal signer based on configuration
func NewIdentity(params IdentityParams) (principal.Signer, error) {
	// Try to load from file first
	if params.Config.KeyFile != "" {
		signer, err := readPrivateKeyFromPEM(params.Config.KeyFile)
		if err == nil {
			log.Infof("Loaded identity from %s: %s", signer.DID().String(), signer.DID())
			return signer, nil
		}
		log.Warnf("Failed to load identity from %s: %v", params.Config.KeyFile, err)
	}
	return nil, fmt.Errorf("keyfile not specified")
}

func readPrivateKeyFromPEM(path string) (principal.Signer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
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

func saveIdentityToFile(signer principal.Signer, path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	// Save with restricted permissions
	return os.WriteFile(path, []byte(signer.Encode()), 0600)
}

// IdentityModule provides identity-related dependencies
var IdentityModule = fx.Module("identity",
	fx.Provide(NewIdentity),
)
