package cmd

import (
	"bytes"
	crypto_ed25519 "crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/storage/cmd/types"
	"github.com/urfave/cli/v2"
)

var keyFormat string

var IdentityCmd = &cli.Command{
	Name:    "identity",
	Aliases: []string{"id"},
	Usage:   "Identity tools.",
	Subcommands: []*cli.Command{
		{
			Name:    "generate",
			Aliases: []string{"gen"},
			Usage:   "Generate a new decentralized identity.",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "type",
					Aliases:     []string{"t"},
					Usage:       fmt.Sprintf("Output format type. Accepted values: %s", types.KeyFormats.All()),
					Value:       "JSON",
					Destination: &keyFormat,
				},
			},
			Action: func(cCtx *cli.Context) error {
				format := types.ParseKeyFormat(strings.ToUpper(keyFormat))
				if !format.IsValid() {
					return fmt.Errorf("unknown type: '%s'. Accepted values: %s", keyFormat, types.KeyFormats.All())
				}

				signer, err := ed25519.Generate()
				if err != nil {
					return fmt.Errorf("generating ed25519 key: %w", err)
				}

				var out []byte
				switch format {
				case types.KeyFormats.JSON:
					out, err = marshalJSONKey(signer)
					if err != nil {
						return fmt.Errorf("marshaling JSON: %w", err)
					}
				case types.KeyFormats.PEM:
					out, err = marshalPEMKey(signer)
					if err != nil {
						return fmt.Errorf("marshaling PEM: %w", err)
					}
				default:
					return fmt.Errorf("unknown format: %s", keyFormat)
				}

				// print the did to stderr allowing the output of the command to be redirected to a file.
				if _, err := fmt.Fprintf(os.Stderr, "# %s\n", signer.DID()); err != nil {
					return fmt.Errorf("writing output: %w", err)
				}
				if n, err := fmt.Fprintf(os.Stdout, string(out)); err != nil {
					return fmt.Errorf("writing output: %w", err)
				} else if n != len(out) {
					return fmt.Errorf("writing output: wrote %d of %d bytes", n, len(out))
				}

				return nil
			},
		},
	},
}

func marshalPEMKey(signer principal.Signer) ([]byte, error) {
	privateKey := crypto_ed25519.PrivateKey(signer.Raw())
	// Marshal and encode the private key
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("marshaling ed25519 private key: %w", err)
	}
	privateKeyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	buffer := new(bytes.Buffer)
	if err := pem.Encode(buffer, privateKeyBlock); err != nil {
		return nil, fmt.Errorf("encoding ed25519 private key: %w", err)
	}

	// Marshal and encode the public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		return nil, fmt.Errorf("marshaling ed25519 public key: %w", err)
	}
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	// Append the public key block to the same file
	if err := pem.Encode(buffer, publicKeyBlock); err != nil {
		return nil, fmt.Errorf("encoding ed25519 public key: %w", err)
	}
	return buffer.Bytes(), nil
}

func marshalJSONKey(signer principal.Signer) ([]byte, error) {
	did := signer.DID().String()
	key, err := ed25519.Format(signer)
	if err != nil {
		return nil, fmt.Errorf("formatting ed25519 key: %w", err)
	}
	out, err := json.Marshal(struct {
		DID string `json:"did"`
		Key string `json:"key"`
	}{did, key})
	if err != nil {
		return nil, fmt.Errorf("marshaling JSON: %w", err)
	}
	return out, nil
}
