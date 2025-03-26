package main

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/curio/pdp/contract"
)

const data = `{"logs": [{"data": "0x0000000000000000000000000000000000000000000000000000000000266bec0000000000000000000000000000000000000000000000000000000000000001", "topics": ["0xb3bcc997e36eca7b534a754795c8c1277f79fb6b97ac09855fb2c3f58a83cd3a", "0x0000000000000000000000000000000000000000000000000000000000000015"], "address": "0xb1b1df5c1eb5338e32a7ee6b5e47980fb892bb9f", "removed": false, "logIndex": "0x4", "blockHash": "0xea76d952c8cc397c4bedbf1a7d937c3e2945e9ac4f04f30449d8a69a0f93318d", "blockNumber": "0x266bec", "transactionHash": "0x7c1f892a2ef4849ba9c8c06efd9ecb30a40740f7044f1a4af98bb34194c35326", "transactionIndex": "0x2"}, {"data": "0x", "topics": ["0x5979d495e336598dba8459e44f8eb2a1c957ce30fcc10cabea4bb0ffe969df6a", "0x0000000000000000000000000000000000000000000000000000000000000015"], "address": "0x58b1b601ee88044f5a7f56b3abec45faa7e7681b", "removed": false, "logIndex": "0x5", "blockHash": "0xea76d952c8cc397c4bedbf1a7d937c3e2945e9ac4f04f30449d8a69a0f93318d", "blockNumber": "0x266bec", "transactionHash": "0x7c1f892a2ef4849ba9c8c06efd9ecb30a40740f7044f1a4af98bb34194c35326", "transactionIndex": "0x2"}], "root": "0x0000000000000000000000000000000000000000000000000000000000000000", "status": "0x1", "gasUsed": "0x4b9917a", "blockHash": "0xea76d952c8cc397c4bedbf1a7d937c3e2945e9ac4f04f30449d8a69a0f93318d", "logsBloom": "0x00000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000040400000000000000000000000004000000000000000000000000000000000000000000000000000000000000000000000001000000204000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000400020800000000000000000000000000002000001000000000000000000000000000000000010000000000000000000000", "blockNumber": "0x266bec", "contractAddress": "0x0000000000000000000000000000000000000000", "transactionHash": "0x7c1f892a2ef4849ba9c8c06efd9ecb30a40740f7044f1a4af98bb34194c35326", "transactionIndex": "0x2", "cumulativeGasUsed": "0x0", "effectiveGasPrice": "0x4a7eac32"}`

func main() {

	var txReceipt types.Receipt
	err := json.Unmarshal([]byte(data), &txReceipt)
	if err != nil {
		panic(err)
	}
	fmt.Printf("tx receipt:\n%+v\n", txReceipt.Logs)

	// Parse the logs to extract the proofSetId
	proofSetId, err := extractProofSetIdFromReceipt(&txReceipt)
	if err != nil {
		panic(err)
	}
	fmt.Println(proofSetId)
}

func extractProofSetIdFromReceipt(receipt *types.Receipt) (uint64, error) {
	pdpABI, err := contract.PDPVerifierMetaData.GetAbi()
	if err != nil {
		return 0, xerrors.Errorf("failed to get PDP ABI: %w", err)
	}

	for _, event := range pdpABI.Events {
		fmt.Println(event.Name, event.ID)
	}

	event, exists := pdpABI.Events["ProofSetCreated"]
	if !exists {
		return 0, xerrors.Errorf("ProofSetCreated event not found in ABI")
	}

	newHash := common.HexToHash("0x5979d495e336598dba8459e44f8eb2a1c957ce30fcc10cabea4bb0ffe969df6a")
	fmt.Println("event:", event.ID)
	for _, vLog := range receipt.Logs {
		for _, t := range vLog.Topics {
			fmt.Println(t.String())
		}
		if len(vLog.Topics) > 0 && vLog.Topics[0] == newHash {
			if len(vLog.Topics) < 2 {
				return 0, xerrors.Errorf("log does not contain setId topic")
			}

			setIdBigInt := new(big.Int).SetBytes(vLog.Topics[1].Bytes())
			return setIdBigInt.Uint64(), nil
		}
	}

	return 0, xerrors.Errorf("ProofSetCreated event not found in receipt")
}
