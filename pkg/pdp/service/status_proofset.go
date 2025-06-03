package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"

	"github.com/storacha/piri/pkg/pdp/service/models"
)

type ProofSetStatus struct {
	CreateMessageHash string
	ProofsetCreated   bool
	Service           string
	OK                bool
	TxStatus          string
	ProofSetId        int64
}

func (p *PDPService) ProofSetStatus(ctx context.Context, txHash common.Hash) (*ProofSetStatus, error) {
	var proofSetCreate models.PDPProofsetCreate
	if err := p.db.WithContext(ctx).
		Where("create_message_hash = ?", txHash.Hex()).
		First(&proofSetCreate).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("proof set creation not for for given txHash")
		}
		return nil, fmt.Errorf("failed to retrieve proof set creation: %w", err)
	}

	if proofSetCreate.Service != p.name {
		return nil, fmt.Errorf("proof set creation not for given service")
	}

	response := &ProofSetStatus{
		CreateMessageHash: proofSetCreate.CreateMessageHash,
		ProofsetCreated:   proofSetCreate.ProofsetCreated,
		Service:           proofSetCreate.Service,
	}
	if proofSetCreate.Ok != nil {
		response.OK = *proofSetCreate.Ok
	}

	// Now get the tx_status from message_waits_eth
	var ethMsgWait models.MessageWaitsEth
	if err := p.db.WithContext(ctx).
		Where("signed_tx_hash = ?", txHash.Hex()).
		First(&ethMsgWait).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("proof set creation not for for given txHash")
		}
		return nil, fmt.Errorf("failed to query proof set creation: %w", err)
	}

	response.TxStatus = ethMsgWait.TxStatus

	if proofSetCreate.ProofsetCreated {
		// The proof set has been created, get the proofSetId from pdp_proof_sets
		var proofSet models.PDPProofSet
		if err := p.db.WithContext(ctx).
			Where("create_message_hash = ?", txHash.Hex()).
			Find(&proofSet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("proof set not found despite proofset_created = true")
			}
			return nil, fmt.Errorf("failed to retrieve proof set: %w", err)
		}
		response.ProofSetId = proofSet.ID

	}

	return response, nil
}
