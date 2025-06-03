package main

import (
	"os"

	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"

	"github.com/storacha/piri/cmd"
)

var log = logging.Logger("piri")

func main() {
	app := &cli.App{
		Name:  "piri",
		Usage: "Manage running a piri node.",
		Commands: []*cli.Command{
			cmd.StartCmd,
			cmd.IdentityCmd,
			cmd.DelegationCmd,
			cmd.ClientCmd,
			cmd.ProofSetCmd,
			cmd.VersionCmd,
			cmd.WalletCmd,
			cmd.ServeCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
