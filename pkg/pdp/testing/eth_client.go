package testing

import (
	"context"
	"math/big"

	"go.uber.org/mock/gomock"

	"github.com/storacha/piri/internal/mocks"
)

// MockEthClient combines all mock interfaces for EthClient
type MockEthClient struct {
	*mocks.MockSenderETHClient
	*mocks.MockMessageWatcherEthClient
	*MockContractBackendWrapper
}

// MockContractBackendWrapper wraps MockContractBackend but excludes the conflicting method
type MockContractBackendWrapper struct {
	*mocks.MockContractBackend
}

// SuggestGasTipCap delegates to MockSenderETHClient's implementation
// This overrides the method from MockContractBackend to avoid the conflict
func (m *MockEthClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	// Always use the SenderETHClient implementation
	return m.MockSenderETHClient.SuggestGasTipCap(ctx)
}

// NewMockEthClient creates a new mock instance that implements all required interfaces
func NewMockEthClient(ctrl *gomock.Controller) *MockEthClient {
	return &MockEthClient{
		MockSenderETHClient:         mocks.NewMockSenderETHClient(ctrl),
		MockMessageWatcherEthClient: mocks.NewMockMessageWatcherEthClient(ctrl),
		MockContractBackendWrapper: &MockContractBackendWrapper{
			MockContractBackend: mocks.NewMockContractBackend(ctrl),
		},
	}
}
