package cmd

import (
	"bytes"
	crypto_ed25519 "crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/storacha/go-ucanto/principal"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/urfave/cli/v2"

	"github.com/storacha/storage/cmd/enum"
)

// JsonKey represents the structure of a JSON-formatted key
type JsonKey struct {
	DID string `json:"did"`
	Key string `json:"key"`
}

// IdentityCmd is the main command for identity-related operations
var IdentityCmd = &cli.Command{
	Name:      "identity",
	Aliases:   []string{"id"},
	UsageText: "storage identity [generate|parse]",
	Description: `
This command provides a set of subcommands for working with decentralized identities.
Specifically for generating and managing Ed25519 keys used in DID (Decentralized Identifier) systems.
  - Generate new Ed25519 key pairs and encode them in either PEM or JSON format
  - Convert between PEM and JSON key formats
  - Extract DIDs from keys
  - Comprehensive round-trip testing to ensure format conversions maintain key integrity

Usage: 
  - Generate a new key in JSON format
    storage identity generate --type=JSON > my-key.json

  - Generate a new key in PEM format
    storage identity generate --type=PEM > my-key.pem

  - Convert from JSON to PEM
    storage identity parse my-key.json > my-key.pem

  - Convert from PEM to JSON
    storage identity parse my-key.pem > my-key.json
`,
	Usage: "Identity tools.",
	Subcommands: []*cli.Command{
		parseCmd,
		generateCmd,
	},
}

var keyFormat string

// generateCmd creates a new decentralized identity
var generateCmd = &cli.Command{
	Name:      "generate",
	Aliases:   []string{"gen"},
	UsageText: "storage identity generate [--type=<format>]",
	Description: `
Generate a new Ed25519 key pair for use with decentralized identities (DIDs).
The command will output the key in the specified format to stdout, which can be
redirected to a file. The DID is printed to stderr for convenience.

The key can be generated in two formats:
  - JSON: A JSON object containing the DID and the encoded key
  - PEM: A PEM-encoded private key followed by the public key

Examples:
  - Generate a key in JSON format:
    storage identity generate --type=JSON > my-key.json

  - Generate a key in PEM format:
    storage identity generate --type=PEM > my-key.pem

  - Generate a key with default format (JSON) and save to file:
    storage identity generate > my-key.json

Note:
  - The DID is printed to stderr to allow redirecting only the key to a file
  - The key file should be kept secure as it contains sensitive private key material
`,
	Usage: "Generate a new decentralized identity as PEM or JSON.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "type",
			Aliases:     []string{"t"},
			Usage:       fmt.Sprintf("Output format type. Accepted values: %s", enum.KeyFormats.All()),
			Value:       "JSON",
			Destination: &keyFormat,
		},
	},
	Action: func(cCtx *cli.Context) error {
		format := enum.ParseKeyFormat(strings.ToUpper(keyFormat))
		if !format.IsValid() {
			return fmt.Errorf("unknown type: '%s'. Accepted values: %s", keyFormat, enum.KeyFormats.All())
		}

		signer, out, err := CreateSignerKeyPair(format)
		if err != nil {
			return err
		}

		// Print the DID to stderr allowing the output of the command to be redirected to a file
		if _, err := fmt.Fprintf(os.Stderr, "# %s\n", signer.DID()); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}

		if n, err := fmt.Fprint(os.Stdout, string(out)); err != nil {
			return fmt.Errorf("writing output: %w", err)
		} else if n != len(out) {
			return fmt.Errorf("writing output: wrote %d of %d bytes", n, len(out))
		}

		return nil
	},
}

// parseCmd converts between PEM and JSON formats
var parseCmd = &cli.Command{
	Name:      "parse",
	UsageText: "storage identity parse <file>",
	Description: `
Convert between PEM and JSON key formats. The command automatically detects the input
format based on the file extension and converts to the opposite format.

The conversion process preserves the original key material and DID. This allows
for a complete round-trip conversion between formats without loss of information.

Supported conversions:
  - .json file -> PEM format
  - .pem file -> JSON format

Examples:
  - Convert a JSON key to PEM format:
    storage identity parse my-key.json > my-key.pem

  - Convert a PEM key to JSON format:
    storage identity parse my-key.pem > my-key.json

Note:
  - The file extension must match the actual format (.json or .pem)
  - The output is printed to stdout and can be redirected to a file
  - Both formats contain the same cryptographic material, just encoded differently
`,
	Args:  true,
	Usage: "Convert between PEM and JSON formats",
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 1 {
			return fmt.Errorf("usage: parse <file>")
		}

		keyFile := cctx.Args().First()
		keyFileExt := strings.Trim(filepath.Ext(keyFile), ".")
		keyFormat := enum.ParseKeyFormat(keyFileExt)

		var (
			out []byte
			err error
		)

		switch keyFormat {
		case enum.KeyFormats.PEM:
			out, err = PEMFileToJSON(keyFile)
			if err != nil {
				return err
			}
		case enum.KeyFormats.JSON:
			out, err = JSONFileToPEM(keyFile)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown file type: '%s'. Accepted values: %s", keyFileExt, enum.KeyFormats.All())
		}

		fmt.Println(string(out))
		return nil
	},
}

// CreateSignerKeyPair generates a new key pair and returns it in the requested format
func CreateSignerKeyPair(format enum.KeyFormat) (principal.Signer, []byte, error) {
	signer, err := ed25519.Generate()
	if err != nil {
		return nil, nil, fmt.Errorf("generating ed25519 key: %w", err)
	}

	var out []byte
	switch format {
	case enum.KeyFormats.JSON:
		out, err = MarshalJSONKey(signer)
		if err != nil {
			return nil, nil, fmt.Errorf("marshaling JSON: %w", err)
		}
	case enum.KeyFormats.PEM:
		out, err = MarshalPEMKey(signer)
		if err != nil {
			return nil, nil, fmt.Errorf("marshaling PEM: %w", err)
		}
	default:
		return nil, nil, fmt.Errorf("unknown format: %s", keyFormat)
	}

	return signer, out, nil
}

// JSONFileToPEM converts a JSON key file to PEM format
func JSONFileToPEM(path string) ([]byte, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer jsonFile.Close()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var jsonKey JsonKey
	if err := json.Unmarshal(jsonData, &jsonKey); err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	ed25519Sk, err := ed25519.Parse(jsonKey.Key)
	if err != nil {
		return nil, fmt.Errorf("parsing key: %w", err)
	}

	return MarshalPEMKey(ed25519Sk)
}

// PEMFileToJSON converts a PEM key file to JSON format
func PEMFileToJSON(path string) ([]byte, error) {
	pemFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file %s: %w", path, err)
	}
	defer pemFile.Close()

	pemData, err := io.ReadAll(pemFile)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	blk, _ := pem.Decode(pemData)
	if blk == nil {
		return nil, fmt.Errorf("no PEM block found")
	}

	sk, err := x509.ParsePKCS8PrivateKey(blk.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing PKCS#8 private key: %w", err)
	}

	ed25519SK, ok := sk.(crypto_ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("PKCS#8 private key does not implement ed25519")
	}

	key, err := ed25519.FromRaw(ed25519SK)
	if err != nil {
		return nil, fmt.Errorf("decoding ed25519 private key: %w", err)
	}

	return MarshalJSONKey(key)
}

// MarshalPEMKey encodes a signer to PEM format
func MarshalPEMKey(signer principal.Signer) ([]byte, error) {
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

// MarshalJSONKey encodes a signer to JSON format
func MarshalJSONKey(signer principal.Signer) ([]byte, error) {
	did := signer.DID().String()
	key, err := ed25519.Format(signer)
	if err != nil {
		return nil, fmt.Errorf("formatting ed25519 key: %w", err)
	}

	out, err := json.Marshal(JsonKey{
		DID: did,
		Key: key,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling JSON: %w", err)
	}

	return out, nil
}
