// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/storacha/piri/pkg/pdp/service/contract (interfaces: PDPProvingSchedule)
//
// Generated by this command:
//
//	mockgen -destination=./internal/mocks/pdp_proving_schedule.go -package=mocks github.com/storacha/piri/pkg/pdp/service/contract PDPProvingSchedule
//

// Package mocks is a generated GoMock package.
package mocks

import (
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"go.uber.org/mock/gomock"
)

// MockPDPProvingSchedule is a mock of PDPProvingSchedule interface.
type MockPDPProvingSchedule struct {
	ctrl     *gomock.Controller
	recorder *MockPDPProvingScheduleMockRecorder
	isgomock struct{}
}

// MockPDPProvingScheduleMockRecorder is the mock recorder for MockPDPProvingSchedule.
type MockPDPProvingScheduleMockRecorder struct {
	mock *MockPDPProvingSchedule
}

// NewMockPDPProvingSchedule creates a new mock instance.
func NewMockPDPProvingSchedule(ctrl *gomock.Controller) *MockPDPProvingSchedule {
	mock := &MockPDPProvingSchedule{ctrl: ctrl}
	mock.recorder = &MockPDPProvingScheduleMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPDPProvingSchedule) EXPECT() *MockPDPProvingScheduleMockRecorder {
	return m.recorder
}

// ChallengeWindow mocks base method.
func (m *MockPDPProvingSchedule) ChallengeWindow(opts *bind.CallOpts) (*big.Int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChallengeWindow", opts)
	ret0, _ := ret[0].(*big.Int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ChallengeWindow indicates an expected call of ChallengeWindow.
func (mr *MockPDPProvingScheduleMockRecorder) ChallengeWindow(opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChallengeWindow", reflect.TypeOf((*MockPDPProvingSchedule)(nil).ChallengeWindow), opts)
}

// GetMaxProvingPeriod mocks base method.
func (m *MockPDPProvingSchedule) GetMaxProvingPeriod(opts *bind.CallOpts) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMaxProvingPeriod", opts)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMaxProvingPeriod indicates an expected call of GetMaxProvingPeriod.
func (mr *MockPDPProvingScheduleMockRecorder) GetMaxProvingPeriod(opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMaxProvingPeriod", reflect.TypeOf((*MockPDPProvingSchedule)(nil).GetMaxProvingPeriod), opts)
}

// InitChallengeWindowStart mocks base method.
func (m *MockPDPProvingSchedule) InitChallengeWindowStart(opts *bind.CallOpts) (*big.Int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitChallengeWindowStart", opts)
	ret0, _ := ret[0].(*big.Int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InitChallengeWindowStart indicates an expected call of InitChallengeWindowStart.
func (mr *MockPDPProvingScheduleMockRecorder) InitChallengeWindowStart(opts any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitChallengeWindowStart", reflect.TypeOf((*MockPDPProvingSchedule)(nil).InitChallengeWindowStart), opts)
}

// NextChallengeWindowStart mocks base method.
func (m *MockPDPProvingSchedule) NextChallengeWindowStart(opts *bind.CallOpts, setId *big.Int) (*big.Int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NextChallengeWindowStart", opts, setId)
	ret0, _ := ret[0].(*big.Int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NextChallengeWindowStart indicates an expected call of NextChallengeWindowStart.
func (mr *MockPDPProvingScheduleMockRecorder) NextChallengeWindowStart(opts, setId any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NextChallengeWindowStart", reflect.TypeOf((*MockPDPProvingSchedule)(nil).NextChallengeWindowStart), opts, setId)
}
