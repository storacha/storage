package contract

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/snadrus/must"

	"github.com/storacha/piri/pkg/build"
)

type PDPContracts struct {
	PDPVerifier common.Address
}

func Addresses() PDPContracts {
	// addresses here based on https://github.com/FilOzone/pdp/?tab=readme-ov-file#contracts
	switch build.BuildType {
	case build.BuildCalibnet:
		return PDPContracts{
			PDPVerifier: common.HexToAddress("0x5A23b7df87f59A291C26A2A1d684AD03Ce9B68DC"),
		}
	case build.BuildMainnet:
		return PDPContracts{
			PDPVerifier: common.HexToAddress("0x9C65E8E57C98cCc040A3d825556832EA1e9f4Df6"),
		}
	default:
		panic("pdp contracts unknown for this network")
	}
}

const NumChallenges = 5

func SybilFee() *big.Int {
	return must.One(types.ParseFIL("0.1")).Int
}
