package cmd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	leveldb "github.com/ipfs/go-ds-leveldb"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/storacha/storage/pkg/store/keystore"
	"github.com/storacha/storage/pkg/wallet"
)

const WalletDir = "wallet"

var WalletCmd = &cli.Command{
	Name:  "wallet",
	Usage: "Manage wallet",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "data-dir",
			Aliases: []string{"d"},
			Usage:   "Root directory to store data in.",
			EnvVars: []string{"STORAGE_DATA_DIR"},
		},
	},
	Subcommands: []*cli.Command{
		walletImport,
		walletList,
	},
}

var walletImport = &cli.Command{
	Name:  "import",
	Usage: "import keys",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "format",
			Usage: "specify input format for key",
			Value: "hex",
		},
	},

	Action: func(cctx *cli.Context) error {
		inpdata, err := os.ReadFile(cctx.Args().First())
		if err != nil {
			return err
		}

		var ki struct {
			Type       string
			PrivateKey []byte
		}
		switch cctx.String("format") {
		case "hex":
			data, err := hex.DecodeString(strings.TrimSpace(string(inpdata)))
			if err != nil {
				return err
			}

			if err := json.Unmarshal(data, &ki); err != nil {
				return err
			}
		case "json":
			if err := json.Unmarshal(inpdata, &ki); err != nil {
				return err
			}
		case "gfc":
			var f struct {
				KeyInfo []struct {
					PrivateKey []byte
					SigType    int
				}
			}
			if err := json.Unmarshal(inpdata, &f); err != nil {
				return xerrors.Errorf("failed to parse go-filecoin key: %s", err)
			}

			gk := f.KeyInfo[0]
			ki.PrivateKey = gk.PrivateKey
			switch gk.SigType {
			// NB(forrest): for now we only accept delegated address types as it's the required type for interacting with the PDP Contract.
			/*
				case 1:
					ki.Type = types.KTSecp256k1
				case 2:
					ki.Type = types.KTBLS

			*/
			case 3:
				ki.Type = "delegated"
			default:
				return fmt.Errorf("unrecognized key type: %d", gk.SigType)
			}
		default:
			return fmt.Errorf("unrecognized format: %s", cctx.String("format"))
		}

		dataDir := cctx.String("data-dir")
		if dataDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting user home directory: %w", err)
			}

			dir, err := mkdirp(homeDir, ".storacha")
			if err != nil {
				return err
			}
			log.Warnf("Data directory is not configured, using default: %s", dir)
			dataDir = dir
		}

		walletDir, err := mkdirp(dataDir, WalletDir)
		if err != nil {
			return err
		}

		walletDs, err := leveldb.NewDatastore(walletDir, nil)
		if err != nil {
			return err
		}

		keyStore, err := keystore.NewKeyStore(walletDs)
		if err != nil {
			return err
		}

		wlt, err := wallet.NewWallet(keyStore)
		if err != nil {
			return err
		}

		addr, err := wlt.Import(cctx.Context, &keystore.KeyInfo{PrivateKey: ki.PrivateKey})
		if err != nil {
			return err
		}

		fmt.Printf("imported wallet %s successfully!\n", addr)
		return nil
	},
}

var walletList = &cli.Command{
	Name:  "list",
	Usage: "List wallet address",
	Action: func(cctx *cli.Context) error {
		dataDir := cctx.String("data-dir")
		if dataDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting user home directory: %w", err)
			}

			dir, err := mkdirp(homeDir, ".storacha")
			if err != nil {
				return err
			}
			log.Warnf("Data directory is not configured, using default: %s", dir)
			dataDir = dir
		}
		pdpDir, err := mkdirp(dataDir, WalletDir)
		if err != nil {
			return err
		}

		pdpDs, err := leveldb.NewDatastore(pdpDir, nil)
		if err != nil {
			return err
		}

		keyStore, err := keystore.NewKeyStore(pdpDs)
		if err != nil {
			return err
		}

		wlt, err := wallet.NewWallet(keyStore)
		if err != nil {
			return err
		}

		kis, err := wlt.List(cctx.Context)
		if err != nil {
			return err
		}

		for _, k := range kis {
			fmt.Println("Address: ", k.Address.String())
		}

		return nil
	},
}
