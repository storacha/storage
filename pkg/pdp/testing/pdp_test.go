package testing

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/storacha/storage/pkg/pdp/service/models"
	"github.com/storacha/storage/pkg/pdp/tasks"
)

func TestPDPService(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	harness, svc := SetupTestDeps(t, ctx, ctrl)

	err := svc.Start(ctx)
	require.NoError(t, err)

	defer func() {
		if err := svc.Stop(ctx); err != nil {
			t.Logf("failed to stop service: %v", err)
		}
	}()

	proofSetID := uint64(1)
	harness.SendCreateProofSet(proofSetID)

	proofSetCreatTx, err := svc.ProofSetCreate(ctx, RecordKeepAddress)
	require.NoError(t, err)

	harness.Require_MessageSendsEth_SendSuccess("pdp-mkproofset", proofSetCreatTx)

	// trigger processing of watch_eth task
	currentHeight := harness.Chain.AdvanceByHeight(tasks.MinConfidence)

	harness.EthClient.MockMessageWatcherEthClient.EXPECT().
		TransactionReceipt(gomock.Any(), common.HexToHash(proofSetCreatTx.Hex())).
		Return(NewSuccessfulReceipt(int64(currentHeight-tasks.MinConfidence)), nil)

	harness.EthClient.MockMessageWatcherEthClient.EXPECT().
		TransactionByHash(gomock.Any(), common.HexToHash(proofSetCreatTx.Hex())).
		Return(NewCreateProofSetTransaction(t), false, nil)

	harness.WaitFor_MessageWaitsEth_TxSuccess(proofSetCreatTx)

	// trigger processing of watch_create task
	harness.Chain.AdvanceChain()

	harness.Contract.EXPECT().GetProofSetIdFromReceipt(gomock.Any()).Return(proofSetID, nil)
	harness.Verifier.EXPECT().GetProofSetListener(gomock.Any(), big.NewInt(1)).Return(ListenerAddress, nil).AnyTimes()
	harness.Schedule.EXPECT().GetMaxProvingPeriod(gomock.Any()).Return(uint64(1), nil)
	harness.Schedule.EXPECT().ChallengeWindow(gomock.Any()).Return(big.NewInt(1), nil)

	harness.WaitFor_PDPProofsetCreate_OK(proofSetCreatTx)
	harness.WaitFor_PDPProofsetCreate_ProofsetCreated(proofSetCreatTx)

	var proofSet []models.PDPProofSet
	res := harness.DB.Where(&models.PDPProofSet{CreateMessageHash: proofSetCreatTx.Hex()}).
		Find(&proofSet)

	require.NoError(t, res.Error)
	require.Len(t, proofSet, 1)
	require.Nil(t, proofSet[0].PrevChallengeRequestEpoch)
	require.Nil(t, proofSet[0].ChallengeRequestTaskID)
	require.Nil(t, proofSet[0].ChallengeRequestMsgHash)
	require.Nil(t, proofSet[0].ProveAtEpoch)
	require.False(t, proofSet[0].InitReady)
	require.EqualValues(t, 1, *proofSet[0].ChallengeWindow)
	require.EqualValues(t, 1, *proofSet[0].ProvingPeriod)
}
