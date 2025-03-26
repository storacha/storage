package ethereum

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type Sender interface {
	Send(ctx context.Context, fromAddress common.Address, tx *ethtypes.Transaction, reason string) (common.Hash, error)
}
