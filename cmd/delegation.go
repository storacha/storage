package cmd

import (
	"fmt"

	"github.com/storacha/go-libstoracha/capabilities/blob"
	"github.com/storacha/go-libstoracha/capabilities/blob/replica"
	"github.com/storacha/go-libstoracha/capabilities/pdp"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/did"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/urfave/cli/v2"
)

var DelegationCmd = &cli.Command{
	Name:    "delegation",
	Aliases: []string{"dg"},
	Usage:   "Delegation tools.",
	Subcommands: []*cli.Command{
		{
			Name:    "generate",
			Aliases: []string{"gen"},
			Usage:   "Generate a new storage delegation",
			Flags: []cli.Flag{
				KeyFileFlag,
				&cli.StringFlag{
					Name:     "client-did",
					Aliases:  []string{"d"},
					Usage:    "did for a client",
					EnvVars:  []string{"PIRI_CLIENT_DID"},
					Required: true,
				},
			},
			Action: func(cCtx *cli.Context) error {
				id, err := PrincipalSignerFromFile(cCtx.String("key-file"))
				if err != nil {
					return fmt.Errorf("parsing private key: %w", err)
				}
				clientDid, err := did.Parse(cCtx.String("client-did"))
				if err != nil {
					return fmt.Errorf("parsing client-did: %w", err)
				}
				dlg, err := delegation.Delegate(
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
						ucan.NewCapability(
							replica.AllocateAbility,
							id.DID().String(),
							ucan.NoCaveats{},
						),
					},
					delegation.WithNoExpiration(),
				)
				if err != nil {
					return fmt.Errorf("generating delegation: %w", err)
				}
				dlgStr, err := delegation.Format(dlg)
				if err != nil {
					return fmt.Errorf("formatting delegation: %w", err)
				}
				fmt.Println(dlgStr)
				return nil
			},
		},
	},
}
