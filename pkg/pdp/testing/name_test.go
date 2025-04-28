package testing

import (
	"context"
	"math/big"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/sqlite"

	"github.com/storacha/storage/pkg/build"
	"github.com/storacha/storage/pkg/pdp/service"
	"github.com/storacha/storage/pkg/pdp/service/contract"
	"github.com/storacha/storage/pkg/pdp/service/models"
	"github.com/storacha/storage/pkg/pdp/store"
	"github.com/storacha/storage/pkg/pdp/tasks"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/storacha/storage/pkg/store/keystore"
	"github.com/storacha/storage/pkg/wallet"
)

func init() {
	build.BuildType = build.BuildCalibnet
}

const PostgresDBConfig = "host=localhost user=postgres dbname=postgres port=5432 sslmode=disable"
const SQLiteDBConfig = "pdp.db"

// NB: this address is never sent to during testing so it's value is insignificant.
// picked a valid address anyways, but can be whatever we want.
var RecordKeepAddress = common.HexToAddress("0x6170dE2b09b404776197485F3dc6c968Ef948505")

// this address also doesn't matter
var ListenerAddress = common.HexToAddress("0x6170dE2b09b404776197485F3dc6c968Ef948506")

func TestPDPService(t *testing.T) {
	// TODO truncate database tables before each run.
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dbPath := filepath.Join(t.TempDir(), SQLiteDBConfig)

	ks := keystore.NewMemKeyStore()
	wlt, err := wallet.NewWallet(ks)
	require.NoError(t, err)

	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	clientAddress, err := wlt.Import(ctx, &keystore.KeyInfo{PrivateKey: crypto.FromECDSA(privKey)})
	require.NoError(t, err)
	dbDialector := sqlite.Open(dbPath)
	bs := blobstore.NewTODOMapBlobstore()
	ss, err := store.NewStashStore(t.TempDir())
	require.NoError(t, err)

	fakeChain := NewFakeChainClient(t)
	mockEth := NewMockEthClient(ctrl)
	mockContract, mockVerifier, mockScheduler := NewMockContractClient(ctrl)

	s, err := service.NewPDPService(
		dbDialector,
		clientAddress,
		wlt,
		bs,
		ss,
		fakeChain,
		mockEth,
		mockContract,
	)
	require.NoError(t, err)

	err = s.Start(ctx)
	require.NoError(t, err)

	defer func() {
		if err := s.Stop(ctx); err != nil {
			t.Logf("failed to stop service: %v", err)
		}
	}()

	db, err := service.SetupDatabase(dbDialector)
	require.NoError(t, err)

	//
	// simulate sending the create proof set transaction
	nonce := uint64(0)
	mockEth.MockSenderETHClient.EXPECT().EstimateGas(ctx, gomock.Any()).Return(uint64(1), nil)
	mockEth.MockSenderETHClient.EXPECT().HeaderByNumber(ctx, nil).Return(&types.Header{
		BaseFee: big.NewInt(1),
	}, nil)
	mockEth.MockSenderETHClient.EXPECT().SuggestGasTipCap(ctx).Return(big.NewInt(1), nil)
	mockEth.MockSenderETHClient.EXPECT().NetworkID(ctx).Return(big.NewInt(1), nil)
	mockEth.MockSenderETHClient.EXPECT().PendingNonceAt(gomock.Any(), gomock.Any()).Return(nonce, nil)
	mockEth.MockSenderETHClient.EXPECT().NetworkID(gomock.Any()).Return(big.NewInt(1), nil)
	mockEth.MockSenderETHClient.EXPECT().SendTransaction(gomock.Any(), gomock.Any()).Return(nil)
	//
	// end transaction send calls

	createProofSetTxHash, err := s.ProofSetCreate(ctx, RecordKeepAddress)
	require.NoError(t, err)

	signedTxHash := createProofSetTxHash.Hex()
	fromAddress := clientAddress.Hex()
	var message []models.MessageSendsEth
	result := db.Model(&models.MessageSendsEth{}).
		Where("from_address = ?", fromAddress).
		Where("send_reason = ?", "pdp-mkproofset").
		Where("signed_hash = ?", signedTxHash).
		Find(&message)

	require.NoError(t, result.Error)
	require.Len(t, message, 1)
	require.NotNil(t, message[0].SendSuccess)
	require.True(t, *message[0].SendSuccess)

	// advance the chain by min confidence heights to trigger watcher_eth task.
	fakeChain.AdvanceByHeight(tasks.MinConfidence)
	currentHeight := fakeChain.CurrentHeight()

	mockEth.MockMessageWatcherEthClient.EXPECT().
		TransactionReceipt(gomock.Any(), common.HexToHash(*message[0].SignedHash)).
		Return(&types.Receipt{
			BlockNumber: big.NewInt(int64(currentHeight - tasks.MinConfidence)),
			Status:      1, // indicates the message was a successful send
			Logs:        make([]*types.Log, 0),
		}, nil)

	// create the createProofSetTransaction
	abiData, err := contract.PDPVerifierMetaData()
	require.NoError(t, err)
	data, err := abiData.Pack("createProofSet", RecordKeepAddress, []byte{})
	require.NoError(t, err)
	tx := types.NewTransaction(
		0,
		contract.Addresses().PDPVerifier,
		contract.SybilFee(),
		0,
		nil,
		data,
	)
	mockEth.MockMessageWatcherEthClient.EXPECT().
		TransactionByHash(gomock.Any(), common.HexToHash(*message[0].SignedHash)).
		Return(tx, false, nil)

	// trigger the watch_create task
	fakeChain.AdvanceChain()

	// we need a wait here on the message moveing from pending to success before inrementing the
	// chain height to trigger watch_create task successfully.
	// the wait is a polling of the database
	// sleep works for now.
	time.Sleep(time.Second * 10)

	fakeChain.AdvanceChain()

	// TODO we should assert on the receipt here
	// return a proofset with ID 1
	mockContract.EXPECT().GetProofSetIdFromReceipt(gomock.Any()).Return(uint64(1), nil)

	mockVerifier.EXPECT().GetProofSetListener(gomock.Any(), big.NewInt(1)).Return(ListenerAddress, nil).AnyTimes()
	mockScheduler.EXPECT().GetMaxProvingPeriod(gomock.Any()).Return(uint64(1), nil)
	mockScheduler.EXPECT().ChallengeWindow(gomock.Any()).Return(big.NewInt(1), nil)

	time.Sleep(time.Second * 10)

	var proofSetCreate []models.PDPProofsetCreate
	result = db.Model(&models.PDPProofsetCreate{}).
		Where("create_message_hash = ?", *message[0].SignedHash).
		Find(&proofSetCreate)

	require.NoError(t, result.Error)
	require.Len(t, proofSetCreate, 1)
	require.True(t, *proofSetCreate[0].Ok)
	require.True(t, proofSetCreate[0].ProofsetCreated)

	var proofSet []models.PDPProofSet
	result = db.Model(&models.PDPProofSet{}).
		Where("create_message_hash = ?", *message[0].SignedHash).
		Find(&proofSet)

	require.NoError(t, result.Error)
	require.Len(t, proofSet, 1)
	require.Nil(t, proofSet[0].PrevChallengeRequestEpoch)
	require.Nil(t, proofSet[0].ChallengeRequestTaskID)
	require.Nil(t, proofSet[0].ChallengeRequestMsgHash)
	require.Nil(t, proofSet[0].ProveAtEpoch)
	require.False(t, proofSet[0].InitReady)
	require.EqualValues(t, 1, *proofSet[0].ChallengeWindow)
	require.EqualValues(t, 1, *proofSet[0].ProvingPeriod)

}
