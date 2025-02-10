package cmd

import (
	"fmt"

	"github.com/storacha/storage/pkg/build"
	"github.com/urfave/cli/v2"
)

var VersionCmd = &cli.Command{
	Name:  "version",
	Usage: "Version information.",
	Action: func(cCtx *cli.Context) error {
		fmt.Println(build.Version)
		return nil
	},
}
