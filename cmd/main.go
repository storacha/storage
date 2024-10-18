package main

import (
	"encoding/json"
	"fmt"
	"os"

	logging "github.com/ipfs/go-log/v2"
	ed25519 "github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/urfave/cli/v2"
)

var log = logging.Logger("cmd")

func main() {
	logging.SetLogLevel("*", "info")

	app := &cli.App{
		Name:  "storage",
		Usage: "Manage running a storage node.",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start the storage node daemon.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "private-key",
						Aliases: []string{"pk"},
						Usage:   "multibase base64 encoded private key identity for the node",
					},
				},
				Action: func(cCtx *cli.Context) error {
					fmt.Printf(`
 00000000                                                   00                  
00      00    00                                            00                  
 000        000000   00000000   00000  0000000    0000000   00000000    0000000 
    00000     00    00     000  00           00  00     0   00    00         00 
        000   00    00      00  00     00000000  00         00    00    0000000 
000     000   00    00     000  00    000    00  000    00  00    00   00    00 
 000000000    0000   0000000    00     000000000   000000   00    00   000000000

🔥 Storage Node %s

`, "v0.0.0")
					return nil
				},
			},
			{
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
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
