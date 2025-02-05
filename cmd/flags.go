package cmd

import "github.com/urfave/cli/v2"

func RequiredStringFlag(strFlag *cli.StringFlag) *cli.StringFlag {
	copy := *strFlag
	copy.Required = true
	return &copy
}

func RequiredIntFlag(strFlag *cli.Int64Flag) *cli.Int64Flag {
	copy := *strFlag
	copy.Required = true
	return &copy
}

var CurioURLFlag = &cli.StringFlag{
	Name:    "curio-url",
	Aliases: []string{"c"},
	Usage:   "URL of a running instance of curio",
	EnvVars: []string{"STORAGE_CURIO_URL"},
}

var PrivateKeyFlag = &cli.StringFlag{
	Name:     "private-key",
	Aliases:  []string{"s"},
	Usage:    "Multibase base64 encoded private key identity for the node.",
	EnvVars:  []string{"STORAGE_PRIVATE_KEY"},
	Required: true,
}

var ClientKeyFlag = &cli.StringFlag{
	Name:     "client-key",
	Aliases:  []string{"s"},
	Usage:    "Multibase base64 encoded private key identity for the client",
	EnvVars:  []string{"STORAGE_CLIENT_KEY"},
	Required: true,
}
var NodeDIDFlag = &cli.StringFlag{
	Name:     "node-did",
	Aliases:  []string{"nd"},
	Usage:    "did for the storage node",
	EnvVars:  []string{"STORAGE_NODE_DID"},
	Required: true,
}
var NodeURLFlag = &cli.StringFlag{
	Name:     "node-url",
	Aliases:  []string{"nu"},
	Usage:    "url for the storage node",
	EnvVars:  []string{"STORAGE_NODE_URL"},
	Required: true,
}
var ProofFlag = &cli.StringFlag{
	Name:     "proof",
	Aliases:  []string{"p"},
	Usage:    "CAR file containing a storage proof delegation",
	EnvVars:  []string{"STORAGE_CLIENT_PROOF"},
	Required: true,
}

var ClientSetupFlags = []cli.Flag{
	ClientKeyFlag,
	NodeDIDFlag,
	NodeURLFlag,
	ProofFlag,
}

var ProofSetFlag = &cli.Int64Flag{
	Name:    "pdp-proofset",
	Aliases: []string{"pdp"},
	Usage:   "Proofset to use with PDP",
	EnvVars: []string{"STORAGE_PDP_PROOFSET"},
}
