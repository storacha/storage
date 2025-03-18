package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/piece/piece"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/core/ipld/hash/sha256"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/storage/pkg/client"
	"github.com/urfave/cli/v2"
)

var ErrMustBePieceLinkOrHaveSize = errors.New("passing pieceCID v1 requires a size to be present")

var ClientCmd = &cli.Command{
	Name:    "client",
	Aliases: []string{"c"},
	Usage:   "test a running storage node as a client",
	Subcommands: []*cli.Command{
		{
			Name:    "upload",
			Aliases: []string{"ba"},
			Usage:   "invoke a blob allocation",
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:     "space-did",
					Aliases:  []string{"sd"},
					Usage:    "did for the space to use",
					EnvVars:  []string{"STORAGE_CLIENT_SPACE_DID"},
					Required: true,
				},
				&cli.StringFlag{
					Name:     "blob",
					Aliases:  []string{"b"},
					Usage:    "blob to upload",
					EnvVars:  []string{"STORAGE_CLIENT_BLOB_FILE"},
					Required: true,
				},
			}, ClientSetupFlags...),
			Action: func(cCtx *cli.Context) error {
				client, err := getClient(cCtx)
				if err != nil {
					return err
				}
				blobFile, err := os.Open(cCtx.String("blob"))
				if err != nil {
					return fmt.Errorf("opening blob file: %w", err)
				}
				blobData, err := io.ReadAll(blobFile)
				if err != nil {
					return fmt.Errorf("reading blob file: %w", err)
				}
				spaceDid, err := did.Parse(cCtx.String("space-did"))
				if err != nil {
					return fmt.Errorf("parsing space did: %w", err)
				}
				digest, err := sha256.Hasher.Sum(blobData)
				if err != nil {
					return fmt.Errorf("calculating blob digest: %w", err)
				}
				address, err := client.BlobAllocate(spaceDid, digest.Bytes(), uint64(len(blobData)), cidlink.Link{Cid: cid.NewCidV1(cid.Raw, digest.Bytes())})
				if err != nil {
					return fmt.Errorf("invocing blob allocation: %w", err)
				}
				if address != nil {
					fmt.Printf("now uploading to: %s\n", address.URL.String())

					req, err := http.NewRequest(http.MethodPut, address.URL.String(), bytes.NewReader(blobData))
					if err != nil {
						return fmt.Errorf("uploading blob: %w", err)
					}
					req.Header = address.Headers
					res, err := http.DefaultClient.Do(req)
					if err != nil {
						return fmt.Errorf("sending blob: %w", err)
					}
					if res.StatusCode >= 300 || res.StatusCode < 200 {
						resData, err := io.ReadAll(res.Body)
						if err != nil {
							return fmt.Errorf("reading response body: %w", err)
						}
						return fmt.Errorf("unsuccessful put, status: %s, message: %s", res.Status, string(resData))
					}
				}
				blobResult, err := client.BlobAccept(spaceDid, digest.Bytes(), uint64(len(blobData)), cidlink.Link{Cid: cid.NewCidV1(cid.Raw, digest.Bytes())})
				if err != nil {
					return fmt.Errorf("accepting blob: %w", err)
				}
				fmt.Printf("uploaded blob available at: %s\n", blobResult.LocationCommitment.Location[0].String())
				if blobResult.PDPAccept != nil {
					fmt.Printf("submitted for PDP aggregation: %s\n", blobResult.PDPAccept.Piece.Link().String())
				}
				return nil
			},
		},
		{
			Name:    "pdp-info",
			Aliases: []string{"pi"},
			Usage:   "get piece information",
			Flags: append([]cli.Flag{
				&cli.StringFlag{
					Name:     "piece",
					Aliases:  []string{"pc"},
					Usage:    "pieceCID to get information on",
					EnvVars:  []string{"STORAGE_PIECE_CID"},
					Required: true,
				},
				&cli.Uint64Flag{
					Name:    "size",
					Aliases: []string{"ps"},
					Usage:   "optional size if passing a piece cid v1 of data",
					EnvVars: []string{"STORAGE_PIECE_SIZE"},
				},
			}, ClientSetupFlags...),
			Action: func(cCtx *cli.Context) error {
				client, err := getClient(cCtx)
				if err != nil {
					return err
				}
				pieceStr := cCtx.String("piece")
				pieceCid, err := cid.Decode(pieceStr)
				if err != nil {
					return fmt.Errorf("decoding cid: %w", err)
				}
				pieceLink, err := piece.FromLink(cidlink.Link{Cid: pieceCid})
				if err != nil {
					if !cCtx.IsSet("size") {
						return ErrMustBePieceLinkOrHaveSize
					}
					pieceLink, err = piece.FromV1LinkAndSize(cidlink.Link{Cid: pieceCid}, cCtx.Uint64("size"))
					if err != nil {
						return fmt.Errorf("parsing as pieceCID v1: %w", err)
					}
				}
				ok, err := client.PDPInfo(pieceLink)
				if err != nil {
					return fmt.Errorf("getting pdp info: %w", err)
				}
				asJSON, err := json.MarshalIndent(ok, "", "  ")
				if err != nil {
					return fmt.Errorf("marshaling info to json: %w", err)
				}
				fmt.Print(string(asJSON))
				return nil
			},
		},
	},
}

func getClient(cCtx *cli.Context) (*client.Client, error) {
	id, err := PrincipalSignerFromFile(cCtx.String("key-file"))
	if err != nil {
		return nil, err
	}
	proofFile, err := os.Open(cCtx.String("proof"))
	if err != nil {
		return nil, fmt.Errorf("opening delegation car file: %w", err)
	}
	proofData, err := io.ReadAll(proofFile)
	if err != nil {
		return nil, fmt.Errorf("reading delegation car file: %w", err)
	}
	proof, err := delegation.Extract(proofData)
	if err != nil {
		return nil, fmt.Errorf("extracting storage proof: %w", err)
	}
	nodeDID, err := did.Parse(cCtx.String("node-did"))
	if err != nil {
		return nil, fmt.Errorf("parsing node did: %w", err)
	}
	nodeURL, err := url.Parse(cCtx.String("node-url"))
	if err != nil {
		return nil, fmt.Errorf("parsing node url: %w", err)
	}
	client, err := client.NewClient(client.Config{
		ID:             id,
		StorageNodeID:  nodeDID,
		StorageNodeURL: *nodeURL,
		StorageProof:   delegation.FromDelegation(proof),
	})
	if err != nil {
		return nil, fmt.Errorf("setting up client: %w", err)
	}
	return client, nil
}
