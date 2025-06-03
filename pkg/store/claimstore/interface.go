package claimstore

import (
	"github.com/storacha/piri/pkg/store/delegationstore"
)

type ClaimStore interface {
	delegationstore.DelegationStore
}
