package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/storage/pkg/pdp/curio"
	"github.com/urfave/cli/v2"
)

var ProofSetCmd = &cli.Command{
	Name:    "proofset",
	Aliases: []string{"ps"},
	Usage:   "proofset tools.",
	Subcommands: []*cli.Command{
		{
			Name:    "create",
			Aliases: []string{"c"},
			Usage:   "Generate a new proofset",
			Flags: []cli.Flag{
				RequiredStringFlag(CurioURLFlag),
				PrivateKeyFlag,
				&cli.StringFlag{
					Name:     "record-keeper",
					Aliases:  []string{"rk"},
					Usage:    "Hex address of the record keeper",
					EnvVars:  []string{"STORAGE_RECORD_KEEPER_CONTRACT"},
					Required: true,
				},
			},
			Action: func(cCtx *cli.Context) error {
				curioURLStr := cCtx.String("curio-url")
				curioURL, err := url.Parse(curioURLStr)
				if err != nil {
					return fmt.Errorf("parsing curio URL: %w", err)
				}
				id, err := ed25519.Parse(cCtx.String("private-key"))
				if err != nil {
					return fmt.Errorf("parsing private key: %w", err)
				}
				curioAuth, err := curio.CreateCurioJWTAuthHeader("storacha", id)
				if err != nil {
					return fmt.Errorf("generating curio jwt: %w", err)
				}

				client := curio.New(http.DefaultClient, curioURL, curioAuth)
				statusRef, err := client.CreateProofSet(cCtx.Context, curio.CreateProofSet{
					RecordKeeper: cCtx.String("record-keeper"),
				})
				if err != nil {
					return fmt.Errorf("creating proof set: %w", err)
				}
				fmt.Printf("proof set being created, check status at %s\n", statusRef.URL)
				return nil
			},
		},
		{
			Name:    "status",
			Aliases: []string{"cs"},
			Usage:   "check on progress creating a proofset",
			Flags: []cli.Flag{
				PrivateKeyFlag,
				RequiredStringFlag(CurioURLFlag),
				&cli.StringFlag{
					Name:     "ref-url",
					Aliases:  []string{"ru"},
					Usage:    "Ref URL from create command",
					EnvVars:  []string{"STORAGE_REF_URL"},
					Required: true,
				},
			},
			Action: func(cCtx *cli.Context) error {
				curioURLStr := cCtx.String("curio-url")
				curioURL, err := url.Parse(curioURLStr)
				if err != nil {
					return fmt.Errorf("parsing curio URL: %w", err)
				}
				id, err := ed25519.Parse(cCtx.String("private-key"))
				if err != nil {
					return fmt.Errorf("parsing private key: %w", err)
				}
				curioAuth, err := curio.CreateCurioJWTAuthHeader("storacha", id)
				if err != nil {
					return fmt.Errorf("generating curio jwt: %w", err)
				}

				client := curio.New(http.DefaultClient, curioURL, curioAuth)
				status, err := client.ProofSetCreationStatus(cCtx.Context, curio.StatusRef{
					URL: cCtx.String("ref-url"),
				})
				if err != nil {
					return fmt.Errorf("getting proof set status: %w", err)
				}
				jsonStatus, err := json.MarshalIndent(status, "", "  ")
				if err != nil {
					return fmt.Errorf("rendering json: %w", err)
				}
				fmt.Print(string(jsonStatus))
				return nil
			},
		},
		{
			Name:    "get",
			Aliases: []string{"g"},
			Usage:   "get a proofs set",
			Flags: []cli.Flag{
				PrivateKeyFlag,
				RequiredStringFlag(CurioURLFlag),
				RequiredIntFlag(ProofSetFlag),
			},
			Action: func(cCtx *cli.Context) error {
				curioURLStr := cCtx.String("curio-url")
				curioURL, err := url.Parse(curioURLStr)
				if err != nil {
					return fmt.Errorf("parsing curio URL: %w", err)
				}
				id, err := ed25519.Parse(cCtx.String("private-key"))
				if err != nil {
					return fmt.Errorf("parsing private key: %w", err)
				}
				curioAuth, err := curio.CreateCurioJWTAuthHeader("storacha", id)
				if err != nil {
					return fmt.Errorf("generating curio jwt: %w", err)
				}

				client := curio.New(http.DefaultClient, curioURL, curioAuth)
				proofSet, err := client.GetProofSet(cCtx.Context, cCtx.Uint64("pdp-proofset"))
				if err != nil {
					return fmt.Errorf("getting proof set status: %w", err)
				}
				jsonProofSet, err := json.MarshalIndent(proofSet, "", "  ")
				if err != nil {
					return fmt.Errorf("rendering json: %w", err)
				}
				fmt.Print(string(jsonProofSet))
				return nil
			},
		},
	},
}
