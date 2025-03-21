package contract

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/snadrus/must"
)

type PDPContracts struct {
	PDPVerifier common.Address
}

// TODO make this a configuration parameter.
func ContractAddresses() PDPContracts {
	return PDPContracts{
		PDPVerifier: common.HexToAddress("0x38529187C03de8d60C8489af063c675b0892CCD9"),
	}
}

const NumChallenges = 5

func SybilFee() *big.Int {
	return must.One(types.ParseFIL("0.1")).Int
}
