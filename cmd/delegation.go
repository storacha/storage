package main

import (
	"fmt"
	"io"
	"os"

	"github.com/storacha/go-capabilities/pkg/blob"
	"github.com/storacha/go-capabilities/pkg/pdp"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/did"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/urfave/cli/v2"
)

var delegationCmd = &cli.Command{
	Name:    "delegation",
	Aliases: []string{"dg"},
	Usage:   "Delegation tools.",
	Subcommands: []*cli.Command{
		{
			Name:    "generate",
			Aliases: []string{"gen"},
			Usage:   "Generate a new storage delegation",
			Flags: []cli.Flag{
				PrivateKeyFlag,
				&cli.StringFlag{
					Name:     "client-did",
					Aliases:  []string{"d"},
					Usage:    "did for a client",
					EnvVars:  []string{"STORAGE_CLIENT_DID"},
					Required: true,
				},
			},
			Action: func(cCtx *cli.Context) error {
				id, err := ed25519.Parse(cCtx.String("private-key"))
				if err != nil {
					return fmt.Errorf("parsing private key: %w", err)
				}
				clientDid, err := did.Parse(cCtx.String("client-did"))
				if err != nil {
					return fmt.Errorf("parsing client-did: %w", err)
				}
				delegation, err := delegation.Delegate(
					id,
					clientDid,
					[]ucan.Capability[ucan.NoCaveats]{
						ucan.NewCapability(
							blob.AllocateAbility,
							id.DID().String(),
							ucan.NoCaveats{},
						),
						ucan.NewCapability(
							blob.AcceptAbility,
							id.DID().String(),
							ucan.NoCaveats{},
						),
						ucan.NewCapability(
							pdp.InfoAbility,
							id.DID().String(),
							ucan.NoCaveats{},
						),
					},
					delegation.WithNoExpiration(),
				)
				if err != nil {
					return fmt.Errorf("generating delegation: %w", err)
				}
				_, err = io.Copy(os.Stdout, delegation.Archive())
				return err
			},
		},
	},
}
