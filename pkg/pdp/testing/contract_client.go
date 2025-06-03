package testing

import (
	"go.uber.org/mock/gomock"

	mocks2 "github.com/storacha/piri/pkg/pdp/service/contract/mocks"
)

func NewMockContractClient(ctrl *gomock.Controller) (*mocks2.MockPDP, *mocks2.MockPDPVerifier, *mocks2.MockPDPProvingSchedule) {
	mockPDP := mocks2.NewMockPDP(ctrl)
	mockVerifier := mocks2.NewMockPDPVerifier(ctrl)
	mockSchedule := mocks2.NewMockPDPProvingSchedule(ctrl)

	// Setup the PDP mock to return our mocked verifier and schedule
	mockPDP.EXPECT().NewPDPVerifier(gomock.Any(), gomock.Any()).Return(mockVerifier, nil).AnyTimes()
	mockPDP.EXPECT().NewIPDPProvingSchedule(gomock.Any(), gomock.Any()).Return(mockSchedule, nil).AnyTimes()

	return mockPDP, mockVerifier, mockSchedule
}
