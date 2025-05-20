package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"

	"github.com/storacha/storage/pkg/pdp"
)

var ServeCmd = &cli.Command{
	Name: "serve",
	Subcommands: []*cli.Command{
		pdpCmd,
	},
}

var pdpCmd = &cli.Command{
	Name:  "pdp",
	Usage: "TODO",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Value:   3001,
			Usage:   "Port to bind the server to.",
			EnvVars: []string{"PDP_SERVER_PORT"},
			Action: func(c *cli.Context, v int) error {
				if v <= 0 || v > 65535 {
					return fmt.Errorf("invalid port: must be between 1 and 65535")
				}
				return nil
			},
		},
		&cli.StringFlag{
			Name:    "data-dir",
			Aliases: []string{"d"},
			Usage:   "Root directory to store data in.",
			EnvVars: []string{"PDP_STORAGE_DATA_DIR"},
		},
		&cli.StringFlag{
			Name:    "tmp-dir",
			Aliases: []string{"t"},
			Usage:   "Temporary directory data is uploaded to before being moved to data-dir.",
			EnvVars: []string{"PDP_STORAGE_TMP_DIR"},
		},
		// TODO: these were the default values from testing, and they are reused
		// here for convince, TODO here is to figure out how to just use one
		// API for both lotus and ethereum. iirc Lotus api should support both
		// with some modifications to the lotus config.
		&cli.StringFlag{
			Name:     "lotus-client-host",
			Usage:    "A websock api address of a lotus node",
			Value:    "ws://127.0.0.1:1234/rpc/v1",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "eth-client-host",
			Usage:    "An api address of a eth node",
			Value:    "https://api.calibration.node.glif.io/rpc/v1",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "pdp-address",
			Usage:    "A hex encoded delegate address for interacting with PDP contact",
			Required: true,
			Action: func(context *cli.Context, s string) error {
				if !common.IsHexAddress(s) {
					return fmt.Errorf("invalid address %s", s)
				}
				return nil
			},
		},
	},
	Action: func(cctx *cli.Context) error {
		logging.SetLogLevel("*", "INFO")
		rootDir := cctx.String("data-dir")
		if rootDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting user home directory: %w", err)
			}

			dir, err := mkdirp(homeDir, ".storacha")
			if err != nil {
				return err
			}
			log.Warnf("Data directory is not configured, using default: %s", dir)
			rootDir = dir
		}

		ctx := cctx.Context
		port := cctx.Int("port")
		lotusURL := cctx.String("lotus-client-host")
		ethURL := cctx.String("eth-client-host")
		addrStr := cctx.String("pdp-address")

		dataDir, err := mkdirp(rootDir, "local-pdp")
		if err != nil {
			return err
		}

		svr, err := pdp.NewServer(
			ctx,
			dataDir,
			port,
			lotusURL,
			ethURL,
			"pdp-server.db",
			common.HexToAddress(addrStr),
		)
		if err != nil {
			return fmt.Errorf("creating pdp server: %w", err)
		}

		if err := svr.Start(ctx); err != nil {
			return fmt.Errorf("starting pdp server: %w", err)
		}

		<-ctx.Done()

		stopCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := svr.Stop(stopCtx); err != nil {
			return fmt.Errorf("stopping pdp server: %w", err)
		}
		return nil
	},
}
