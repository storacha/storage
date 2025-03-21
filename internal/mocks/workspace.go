// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/pdp/aggregator/steps.go
//
// Generated by this command:
//
//	mockgen -source=./pkg/pdp/aggregator/steps.go -destination=./internal/mocks/mocks.go -package=mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	fns "github.com/storacha/storage/pkg/pdp/aggregator/fns"
	gomock "go.uber.org/mock/gomock"
)

// MockInProgressWorkspace is a mock of InProgressWorkspace interface.
type MockInProgressWorkspace struct {
	ctrl     *gomock.Controller
	recorder *MockInProgressWorkspaceMockRecorder
	isgomock struct{}
}

// MockInProgressWorkspaceMockRecorder is the mock recorder for MockInProgressWorkspace.
type MockInProgressWorkspaceMockRecorder struct {
	mock *MockInProgressWorkspace
}

// NewMockInProgressWorkspace creates a new mock instance.
func NewMockInProgressWorkspace(ctrl *gomock.Controller) *MockInProgressWorkspace {
	mock := &MockInProgressWorkspace{ctrl: ctrl}
	mock.recorder = &MockInProgressWorkspaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInProgressWorkspace) EXPECT() *MockInProgressWorkspaceMockRecorder {
	return m.recorder
}

// GetBuffer mocks base method.
func (m *MockInProgressWorkspace) GetBuffer(arg0 context.Context) (fns.Buffer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBuffer", arg0)
	ret0, _ := ret[0].(fns.Buffer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBuffer indicates an expected call of GetBuffer.
func (mr *MockInProgressWorkspaceMockRecorder) GetBuffer(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBuffer", reflect.TypeOf((*MockInProgressWorkspace)(nil).GetBuffer), arg0)
}

// PutBuffer mocks base method.
func (m *MockInProgressWorkspace) PutBuffer(arg0 context.Context, arg1 fns.Buffer) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutBuffer", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutBuffer indicates an expected call of PutBuffer.
func (mr *MockInProgressWorkspaceMockRecorder) PutBuffer(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutBuffer", reflect.TypeOf((*MockInProgressWorkspace)(nil).PutBuffer), arg0, arg1)
}
