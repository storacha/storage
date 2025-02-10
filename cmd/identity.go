package cmd

import (
	"encoding/json"
	"fmt"

	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/urfave/cli/v2"
)

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
}
