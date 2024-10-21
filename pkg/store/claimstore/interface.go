package claimstore

import (
	"github.com/storacha/storage/pkg/store/delegationstore"
)

type ClaimStore interface {
	delegationstore.DelegationStore
}
