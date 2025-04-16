package tasks

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	logging "github.com/ipfs/go-log/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/filecoin-project/go-state-types/abi"
	chaintypes "github.com/filecoin-project/lotus/chain/types"

	"github.com/storacha/storage/pkg/pdp/ethereum"
	"github.com/storacha/storage/pkg/pdp/promise"
	"github.com/storacha/storage/pkg/pdp/scheduler"
	"github.com/storacha/storage/pkg/pdp/service/contract"
	"github.com/storacha/storage/pkg/pdp/service/models"
)

var log = logging.Logger("pdp/tasks")

// TODO determine if this is a requirement.
// based on curio it appears this is needed for task summary details via the RPC.
// var _ = scheduler.Reg(&InitProvingPeriodTask{})
var _ scheduler.TaskInterface = &InitProvingPeriodTask{}

type InitProvingPeriodTask struct {
	db        *gorm.DB
	ethClient bind.ContractBackend
	sender    ethereum.Sender

	chain ChainAPI

	addFunc promise.Promise[scheduler.AddTaskFunc]
}

type ChainAPI interface {
	ChainHead(context.Context) (*chaintypes.TipSet, error)
	StateGetRandomnessDigestFromBeacon(ctx context.Context, randEpoch abi.ChainEpoch, tsk chaintypes.TipSetKey) (abi.Randomness, error) //perm:read
}

func NewInitProvingPeriodTask(db *gorm.DB, ethClient bind.ContractBackend, chain ChainAPI, chainSched *scheduler.Chain, sender ethereum.Sender) (*InitProvingPeriodTask, error) {
	ipp := &InitProvingPeriodTask{
		db:        db,
		ethClient: ethClient,
		sender:    sender,
		chain:     chain,
	}

	if err := chainSched.AddHandler(func(ctx context.Context, revert, apply *chaintypes.TipSet) error {
		if apply == nil {
			return nil
		}

		// each time a new head is applied to the chain, query the db for proof sets needing initialization
		// via nextProvingPeriod initial call

		var proofSetIDs []int64
		if err := db.WithContext(ctx).
			Model(&models.PDPProofSet{}).
			Where("challenge_request_task_id IS NULL").
			Where("init_ready = ?", true).
			Where("prove_at_epoch IS NULL").
			Pluck("id", &proofSetIDs).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("failed to select proof sets needing nextProvingPeriod: %w", err)
			}
		}

		for _, psID := range proofSetIDs {
			ipp.addFunc.Val(ctx)(func(taskID scheduler.TaskID, tx *gorm.DB) (shouldCommit bool, seriousError error) {
				result := tx.Model(&models.PDPProofSet{}).
					Where("id = ? AND challenge_request_task_id IS NULL", psID).
					Update("challenge_request_task_id", taskID)
				if result.Error != nil {
					return false, fmt.Errorf("failed to update pdp_proof_sets: %w", result.Error)
				}
				if result.RowsAffected == 0 {
					// With only one worker executing tasks, if no rows are updated it likely means that
					// this record was already processed.
					return false, nil
				}
				return true, nil
			})
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to register pdp InitProvingPersiodTask: %w", err)
	}

	return ipp, nil
}

func (ipp *InitProvingPeriodTask) TypeDetails() scheduler.TaskTypeDetails {
	return scheduler.TaskTypeDetails{
		Name: "PDPInitPP",
	}
}

func (ipp *InitProvingPeriodTask) Do(taskID scheduler.TaskID) (done bool, err error) {
	ctx := context.Background()

	// Select the proof set where challenge_request_task_id = taskID
	var proofSetID int64

	var proofSet models.PDPProofSet
	err = ipp.db.WithContext(ctx).
		Select("id").
		Where("challenge_request_task_id = ?", taskID).
		First(&proofSet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// No matching proof set; task is done (e.g., another task was spawned in place of this one)
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to select PDPProofSet: %w", err)
	}
	proofSetID = proofSet.ID

	// Get the listener address for this proof set from the PDPVerifier contract
	pdpVerifier, err := contract.NewPDPVerifier(contract.ContractAddresses().PDPVerifier, ipp.ethClient)
	if err != nil {
		return false, fmt.Errorf("failed to instantiate PDPVerifier contract: %w", err)
	}

	listenerAddr, err := pdpVerifier.GetProofSetListener(nil, big.NewInt(proofSetID))
	if err != nil {
		return false, fmt.Errorf("failed to get listener address for proof set %d: %w", proofSetID, err)
	}

	// Determine the next challenge window start by consulting the listener
	provingSchedule, err := contract.NewIPDPProvingSchedule(listenerAddr, ipp.ethClient)
	if err != nil {
		return false, fmt.Errorf("failed to create proving schedule binding, check that listener has proving schedule methods: %w", err)
	}

	// ChallengeWindow
	challengeWindow, err := provingSchedule.ChallengeWindow(&bind.CallOpts{Context: ctx})
	if err != nil {
		return false, fmt.Errorf("failed to get challenge window: %w", err)
	}

	init_prove_at, err := provingSchedule.InitChallengeWindowStart(&bind.CallOpts{Context: ctx})
	if err != nil {
		return false, fmt.Errorf("failed to get next challenge window start: %w", err)
	}
	prove_at_epoch := init_prove_at.Add(init_prove_at, challengeWindow.Div(challengeWindow, big.NewInt(2))) // Give a buffer of 1/2 challenge window epochs so that we are still within challenge window
	// Instantiate the PDPVerifier contract
	pdpContracts := contract.ContractAddresses()
	pdpVeriferAddress := pdpContracts.PDPVerifier

	// Prepare the transaction data
	abiData, err := contract.PDPVerifierMetaData.GetAbi()
	if err != nil {
		return false, fmt.Errorf("failed to get PDPVerifier ABI: %w", err)
	}

	data, err := abiData.Pack("nextProvingPeriod", big.NewInt(proofSetID), prove_at_epoch, []byte{})
	if err != nil {
		return false, fmt.Errorf("failed to pack data: %w", err)
	}

	// Prepare the transaction
	txEth := types.NewTransaction(
		0,                 // nonce (will be set by sender)
		pdpVeriferAddress, // to
		big.NewInt(0),     // value
		0,                 // gasLimit (to be estimated)
		nil,               // gasPrice (to be set by sender)
		data,              // data
	)

	fromAddress, _, err := pdpVerifier.GetProofSetOwner(nil, big.NewInt(proofSetID))
	if err != nil {
		return false, fmt.Errorf("failed to get default sender address: %w", err)
	}

	// Get the current tipset
	ts, err := ipp.chain.ChainHead(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to get chain head: %w", err)
	}

	// Send the transaction
	reason := "pdp-proving-init"
	txHash, err := ipp.sender.Send(ctx, fromAddress, txEth, reason)
	if err != nil {
		return false, fmt.Errorf("failed to send transaction: %w", err)
	}

	// Update the database in a transaction
	if err := ipp.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.PDPProofSet{}).
			Where("id = ?", proofSetID).
			Updates(map[string]interface{}{
				"challenge_request_msg_hash":   txHash.Hex(),
				"prev_challenge_request_epoch": ts.Height(),
				"prove_at_epoch":               init_prove_at.Uint64(),
			})
		if result.Error != nil {
			return fmt.Errorf("failed to update pdp_proof_sets: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("pdp_proof_sets update affected 0 rows")
		}

		msg := models.MessageWaitsEth{
			SignedTxHash: txHash.Hex(),
			TxStatus:     "pending",
		}
		// Use OnConflict DoNothing to avoid errors on duplicate keys.
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&msg).Error; err != nil {
			return fmt.Errorf("failed to insert into message_waits_eth: %w", err)
		}

		return nil
	}); err != nil {
		return false, fmt.Errorf("failed to perform database transaction: %w", err)
	}

	// Task completed successfully
	return true, nil
}

func (ipp *InitProvingPeriodTask) Adder(taskFunc scheduler.AddTaskFunc) {
	ipp.addFunc.Set(taskFunc)
}
