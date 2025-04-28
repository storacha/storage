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
	"gorm.io/gorm"

	"github.com/storacha/storage/pkg/build"
	"github.com/storacha/storage/pkg/pdp/service"
	"github.com/storacha/storage/pkg/pdp/service/contract"
	"github.com/storacha/storage/pkg/pdp/service/contract/mocks"
	"github.com/storacha/storage/pkg/pdp/service/models"
	"github.com/storacha/storage/pkg/pdp/store"
	"github.com/storacha/storage/pkg/store/blobstore"
	"github.com/storacha/storage/pkg/store/keystore"
	"github.com/storacha/storage/pkg/wallet"
)

func init() {
	build.BuildType = build.BuildCalibnet
}

// NB: this address is never sent to during testing so it's value is insignificant.
// picked a valid address anyways, but can be whatever we want.
var RecordKeepAddress = common.HexToAddress("0x6170dE2b09b404776197485F3dc6c968Ef948505")

// this address also doesn't matter
var ListenerAddress = common.HexToAddress("0x6170dE2b09b404776197485F3dc6c968Ef948506")

const SQLiteDBConfig = "pdp.db"

type Harness struct {
	T          testing.TB
	Chain      *FakeChainClient
	EthClient  *MockEthClient
	Contract   *mocks.MockPDP
	Verifier   *mocks.MockPDPVerifier
	Schedule   *mocks.MockPDPProvingSchedule
	DB         *gorm.DB
	ClientAddr common.Address

	messageNonce uint64
}

func (h *Harness) SendCreateProofSet(proofSetID uint64) {
	defer func() {
		h.messageNonce++
	}()
	h.EthClient.MockSenderETHClient.EXPECT().EstimateGas(gomock.Any(), gomock.Any()).Return(uint64(1), nil)
	h.EthClient.MockSenderETHClient.EXPECT().HeaderByNumber(gomock.Any(), nil).Return(&types.Header{
		BaseFee: big.NewInt(1),
	}, nil)
	h.EthClient.MockSenderETHClient.EXPECT().SuggestGasTipCap(gomock.Any()).Return(big.NewInt(1), nil)
	h.EthClient.MockSenderETHClient.EXPECT().NetworkID(gomock.Any()).Return(big.NewInt(1), nil)
	h.EthClient.MockSenderETHClient.EXPECT().PendingNonceAt(gomock.Any(), gomock.Any()).Return(h.messageNonce, nil)
	h.EthClient.MockSenderETHClient.EXPECT().NetworkID(gomock.Any()).Return(big.NewInt(1), nil)
	h.EthClient.MockSenderETHClient.EXPECT().SendTransaction(gomock.Any(), gomock.Any()).Return(nil)
}

// Require_MessageSendsEth_SendSuccess asserts a message in MessageSendsEth was sent successfully
// given a reason and signedTx hash.
func (h *Harness) Require_MessageSendsEth_SendSuccess(reason string, signedTx common.Hash) {
	var record []models.MessageSendsEth
	result := h.DB.Where(&models.MessageSendsEth{SignedHash: Ptr(signedTx.Hex())}).Find(&record)

	require.NoError(h.T, result.Error)
	require.Len(h.T, record, 1)

	message := record[0]
	require.NotNil(h.T, message.SendSuccess)
	require.True(h.T, *message.SendSuccess)
	require.NotNil(h.T, message.SendError)
	require.Empty(h.T, *message.SendError)
	require.Equal(h.T, reason, message.SendReason)
	require.Equal(h.T, h.ClientAddr.Hex(), message.FromAddress)
	require.Equal(h.T, contract.Addresses().PDPVerifier.Hex(), message.ToAddress)
}

func (h *Harness) WaitFor_MessageWaitsEth_TxSuccess(signedTx common.Hash) {
	require.Eventually(h.T, func() bool {
		var record models.MessageWaitsEth
		result := h.DB.Where(&models.MessageWaitsEth{SignedTxHash: signedTx.Hex()}).First(&record)
		require.NoError(h.T, result.Error)

		return record.TxSuccess != nil && *record.TxSuccess
	},
		3*time.Second,
		50*time.Millisecond)
}

func (h *Harness) WaitFor_PDPProofsetCreate_OK(signedTx common.Hash) {
	require.Eventually(h.T, func() bool {
		var record models.PDPProofsetCreate
		result := h.DB.Where(&models.PDPProofsetCreate{CreateMessageHash: signedTx.Hex()}).First(&record)
		require.NoError(h.T, result.Error)

		return record.Ok != nil && *record.Ok
	},
		3*time.Second,
		50*time.Millisecond,
	)
}

func (h *Harness) WaitFor_PDPProofsetCreate_ProofsetCreated(signedTx common.Hash) {
	require.Eventually(h.T, func() bool {
		var record models.PDPProofsetCreate
		result := h.DB.Where(&models.PDPProofsetCreate{CreateMessageHash: signedTx.Hex()}).First(&record)
		require.NoError(h.T, result.Error)

		return record.Ok != nil && *record.Ok && record.ProofsetCreated
	},
		3*time.Second,
		50*time.Millisecond,
	)

}

func SetupTestDeps(t testing.TB, ctx context.Context, ctrl *gomock.Controller) (*Harness, *service.PDPService) {
	// a fake chain with methods to advance the state of the chain
	fakeChain := NewFakeChainClient(t)
	// a mocked ethereum client for interactions with the PDP contract
	mockEth := NewMockEthClient(ctrl)
	// various mocks for the PDP constract
	mockContract, mockVerifier, mockScheduler := NewMockContractClient(ctrl)

	// directory all state for test will be initialized in
	testDir := t.TempDir()

	// path to SQLite db used for test case
	dbPath := filepath.Join(testDir, SQLiteDBConfig)
	t.Logf("Database path: %s", dbPath)
	dbDialector := sqlite.Open(dbPath)
	// sqlite database handle, may be queries in tests to assert state of DB
	db, err := service.SetupDatabase(dbDialector)
	require.NoError(t, err)

	// memory backed blob and stash store for PDP pieces.
	ss, err := store.NewStashStore(testDir)
	require.NoError(t, err)
	bs := blobstore.NewTODOMapBlobstore()

	// memory backed keystore for persisting wallets
	ks := keystore.NewMemKeyStore()
	wlt, err := wallet.NewWallet(ks)
	require.NoError(t, err)

	// create a client address for interacting with PDP contract
	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	clientAddress, err := wlt.Import(ctx, &keystore.KeyInfo{PrivateKey: crypto.FromECDSA(privKey)})
	require.NoError(t, err)
	t.Logf("Client address: %s", clientAddress.Hex())

	// The PDP service, backed by mocks and a fake chain
	svc, err := service.NewPDPService(
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

	return &Harness{
		T:          t,
		Chain:      fakeChain,
		EthClient:  mockEth,
		Contract:   mockContract,
		Verifier:   mockVerifier,
		Schedule:   mockScheduler,
		DB:         db,
		ClientAddr: clientAddress,
	}, svc

}

func Ptr[T any](v T) *T {
	return &v
}

func NewCreateProofSetTransaction(t testing.TB) *types.Transaction {
	abiData, err := contract.PDPVerifierMetaData()
	require.NoError(t, err)
	data, err := abiData.Pack("createProofSet", RecordKeepAddress, []byte{})
	require.NoError(t, err)
	return types.NewTransaction(
		0,
		contract.Addresses().PDPVerifier,
		contract.SybilFee(),
		0,
		nil,
		data,
	)
}

func NewSuccessfulReceipt(height int64) *types.Receipt {
	return &types.Receipt{
		BlockNumber: big.NewInt(height),
		Status:      types.ReceiptStatusSuccessful,
		Logs:        make([]*types.Log, 0),
	}
}
