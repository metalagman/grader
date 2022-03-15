// Code generated by MockGen. DO NOT EDIT.
// Source: ./interface.go

// Package tokenmock is a generated GoMock package.
package tokenmock

import (
	token "grader/pkg/token"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockIdentity is a mock of Identity interface.
type MockIdentity struct {
	ctrl     *gomock.Controller
	recorder *MockIdentityMockRecorder
}

// MockIdentityMockRecorder is the mock recorder for MockIdentity.
type MockIdentityMockRecorder struct {
	mock *MockIdentity
}

// NewMockIdentity creates a new mock instance.
func NewMockIdentity(ctrl *gomock.Controller) *MockIdentity {
	mock := &MockIdentity{ctrl: ctrl}
	mock.recorder = &MockIdentityMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIdentity) EXPECT() *MockIdentityMockRecorder {
	return m.recorder
}

// Identity mocks base method.
func (m *MockIdentity) Identity() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Identity")
	ret0, _ := ret[0].(string)
	return ret0
}

// Identity indicates an expected call of Identity.
func (mr *MockIdentityMockRecorder) Identity() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Identity", reflect.TypeOf((*MockIdentity)(nil).Identity))
}

// MockManager is a mock of Manager interface.
type MockManager struct {
	ctrl     *gomock.Controller
	recorder *MockManagerMockRecorder
}

// MockManagerMockRecorder is the mock recorder for MockManager.
type MockManagerMockRecorder struct {
	mock *MockManager
}

// NewMockManager creates a new mock instance.
func NewMockManager(ctrl *gomock.Controller) *MockManager {
	mock := &MockManager{ctrl: ctrl}
	mock.recorder = &MockManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockManager) EXPECT() *MockManagerMockRecorder {
	return m.recorder
}

// Decode mocks base method.
func (m *MockManager) Decode(tk string) (token.Identity, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Decode", tk)
	ret0, _ := ret[0].(token.Identity)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Decode indicates an expected call of Decode.
func (mr *MockManagerMockRecorder) Decode(tk interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Decode", reflect.TypeOf((*MockManager)(nil).Decode), tk)
}

// Issue mocks base method.
func (m *MockManager) Issue(id token.Identity, exp time.Duration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Issue", id, exp)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Issue indicates an expected call of Issue.
func (mr *MockManagerMockRecorder) Issue(id, exp interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Issue", reflect.TypeOf((*MockManager)(nil).Issue), id, exp)
}

// Validate mocks base method.
func (m *MockManager) Validate(token string, target token.Identity) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Validate", token, target)
	ret0, _ := ret[0].(error)
	return ret0
}

// Validate indicates an expected call of Validate.
func (mr *MockManagerMockRecorder) Validate(token, target interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Validate", reflect.TypeOf((*MockManager)(nil).Validate), token, target)
}
