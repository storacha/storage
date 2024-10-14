package main

import (
	"os"

	logging "github.com/ipfs/go-log/v2"
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
				Usage: "Start storage node daemon.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "private-key",
						Aliases: []string{"pk"},
						Usage:   "multibase base64 encoded private key identity for the node",
					},
				},
				Action: func(cCtx *cli.Context) error {
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
