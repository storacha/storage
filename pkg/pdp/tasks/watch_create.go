package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/xerrors"
	"gorm.io/gorm"

	chaintypes "github.com/filecoin-project/lotus/chain/types"

	"github.com/storacha/storage/pkg/pdp/scheduler"
	"github.com/storacha/storage/pkg/pdp/service/contract"
	"github.com/storacha/storage/pkg/pdp/service/models"
)

type ProofSetCreate struct {
	CreateMessageHash string `db:"create_message_hash"`
	Service           string `db:"service"`
}

func NewWatcherCreate(
	db *gorm.DB,
	ethClient bind.ContractBackend,
	contractClient contract.PDP,
	pcs *scheduler.Chain,
) error {
	if err := pcs.AddHandler(func(ctx context.Context, revert, apply *chaintypes.TipSet) error {
		err := processPendingProofSetCreates(ctx, db, ethClient, contractClient)
		if err != nil {
			log.Warnf("Failed to process pending proof set creates: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func processPendingProofSetCreates(
	ctx context.Context,
	db *gorm.DB,
	ethClient bind.ContractBackend,
	contractClient contract.PDP,
) error {
	// Query for pdp_proofset_creates entries where ok = TRUE and proofset_created = FALSE
	var proofSetCreates []models.PDPProofsetCreate
	err := db.WithContext(ctx).
		Where("ok = ? AND proofset_created = ?", true, false).
		Find(&proofSetCreates).Error
	if err != nil {
		return fmt.Errorf("failed to select proof set creates: %w", err)
	}

	if len(proofSetCreates) == 0 {
		// No pending proof set creates
		return nil
	}

	// Process each proof set create
	for _, psc := range proofSetCreates {
		err := processProofSetCreate(ctx, db, psc, ethClient, contractClient)
		if err != nil {
			log.Warnf("Failed to process proof set create for tx %s: %v", psc.CreateMessageHash, err)
			continue
		}
	}

	return nil
}

func processProofSetCreate(
	ctx context.Context,
	db *gorm.DB,
	psc models.PDPProofsetCreate,
	ethClient bind.ContractBackend,
	contactClient contract.PDP,
) error {
	// Retrieve the tx_receipt from message_waits_eth
	var msgWait models.MessageWaitsEth
	err := db.WithContext(ctx).
		Select("tx_receipt").
		First(&msgWait, "signed_tx_hash = ?", psc.CreateMessageHash).Error
	if err != nil {
		return fmt.Errorf("failed to get tx_receipt for tx %s: %w", psc.CreateMessageHash, err)
	}

	txReceiptJSON := msgWait.TxReceipt

	// Unmarshal the tx_receipt JSON into types.Receipt
	var txReceipt types.Receipt
	err = json.Unmarshal(txReceiptJSON, &txReceipt)
	if err != nil {
		return xerrors.Errorf("failed to unmarshal tx_receipt for tx %s: %w", psc.CreateMessageHash, err)
	}

	// Parse the logs to extract the proofSetId
	proofSetId, err := contactClient.GetProofSetIdFromReceipt(&txReceipt)
	if err != nil {
		return xerrors.Errorf("failed to extract proofSetId from receipt for tx %s: %w", psc.CreateMessageHash, err)
	}

	// Get the listener address for this proof set from the PDPVerifier contract
	pdpVerifier, err := contactClient.NewPDPVerifier(contract.Addresses().PDPVerifier, ethClient)
	if err != nil {
		return xerrors.Errorf("failed to instantiate PDPVerifier contract: %w", err)
	}

	listenerAddr, err := pdpVerifier.GetProofSetListener(nil, big.NewInt(int64(proofSetId)))
	if err != nil {
		return xerrors.Errorf("failed to get listener address for proof set %d: %w", proofSetId, err)
	}

	// Get the proving period from the listener
	// Assumption: listener is a PDP Service with proving window informational methods
	provingPeriod, challengeWindow, err := getProvingPeriodChallengeWindow(ctx, ethClient, listenerAddr, contactClient)
	if err != nil {
		return xerrors.Errorf("failed to get max proving period: %w", err)
	}

	// Insert a new entry into pdp_proof_sets
	err = insertProofSet(ctx, db, psc.CreateMessageHash, proofSetId, psc.Service, provingPeriod, challengeWindow)
	if err != nil {
		return xerrors.Errorf("failed to insert proof set %d for tx %+v: %w", proofSetId, psc, err)
	}

	// Update pdp_proofset_creates to set proofset_created = TRUE
	err = db.WithContext(ctx).
		Model(&models.PDPProofsetCreate{}).
		Where("create_message_hash = ?", psc.CreateMessageHash).
		Update("proofset_created", true).Error
	if err != nil {
		return fmt.Errorf("failed to update proofset_creates for tx %s: %w", psc.CreateMessageHash, err)
	}

	return nil
}

func insertProofSet(
	ctx context.Context,
	db *gorm.DB,
	createMsg string,
	proofSetId uint64,
	service string,
	provingPeriod uint64,
	challengeWindow uint64,
) error {
	// Implement the insertion into pdp_proof_sets table
	proofset := models.PDPProofSet{
		ID:                int64(proofSetId),
		CreateMessageHash: createMsg,
		Service:           service,
		ProvingPeriod:     models.Ptr(int64(provingPeriod)),
		ChallengeWindow:   models.Ptr(int64(challengeWindow)),
	}
	err := db.WithContext(ctx).Create(&proofset).Error
	if err != nil {
		return fmt.Errorf("failed to insert proof set %d for tx %+v: %w", proofSetId, proofset, err)
	}
	return nil
}

func getProvingPeriodChallengeWindow(ctx context.Context, ethClient bind.ContractBackend, listenerAddr common.Address, contractClient contract.PDP) (uint64, uint64, error) {
	// ProvingPeriod
	schedule, err := contractClient.NewIPDPProvingSchedule(listenerAddr, ethClient)
	if err != nil {
		return 0, 0, xerrors.Errorf("failed to create proving schedule binding, check that listener has proving schedule methods: %w", err)
	}

	period, err := schedule.GetMaxProvingPeriod(&bind.CallOpts{Context: ctx})
	if err != nil {
		return 0, 0, xerrors.Errorf("failed to get proving period: %w", err)
	}

	// ChallengeWindow
	challengeWindow, err := schedule.ChallengeWindow(&bind.CallOpts{Context: ctx})
	if err != nil {
		return 0, 0, xerrors.Errorf("failed to get challenge window: %w", err)
	}

	return period, challengeWindow.Uint64(), nil
}
