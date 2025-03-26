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

/*
These are the record keepers:
1-day proving service: `0xd394e2504994a3369A87F4a0e8a21f914baf7263`

30-minute proving service: `0x6170dE2b09b404776197485F3dc6c968Ef948505`
# this one is wired up to dashboard(s):
- http://explore-pdp.xyz:5173/ (Wyatt owns this and uses for testing)
- https://calibration.pdp-explorer.eng.filoz.org/ (this is a productionized version)
  - @Puspendra Mahariya is the developer building the frontend. Complaints with the UX go do him.
  - Wyatt is responsible for the backend?
*/

// TODO make this a configuration parameter.
func ContractAddresses() PDPContracts {
	return PDPContracts{
		PDPVerifier: common.HexToAddress("0x5A23b7df87f59A291C26A2A1d684AD03Ce9B68DC"),
	}
}

const NumChallenges = 5

func SybilFee() *big.Int {
	return must.One(types.ParseFIL("0.1")).Int
}
