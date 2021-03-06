// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/demeero/pocket-link/keygen/key (interfaces: UsedKeysRepository)

// Package key is a generated GoMock package.
package key

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockUsedKeysRepository is a mock of UsedKeysRepository interface.
type MockUsedKeysRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUsedKeysRepositoryMockRecorder
}

// MockUsedKeysRepositoryMockRecorder is the mock recorder for MockUsedKeysRepository.
type MockUsedKeysRepositoryMockRecorder struct {
	mock *MockUsedKeysRepository
}

// NewMockUsedKeysRepository creates a new mock instance.
func NewMockUsedKeysRepository(ctrl *gomock.Controller) *MockUsedKeysRepository {
	mock := &MockUsedKeysRepository{ctrl: ctrl}
	mock.recorder = &MockUsedKeysRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUsedKeysRepository) EXPECT() *MockUsedKeysRepositoryMockRecorder {
	return m.recorder
}

// Exists mocks base method.
func (m *MockUsedKeysRepository) Exists(arg0 context.Context, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exists", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exists indicates an expected call of Exists.
func (mr *MockUsedKeysRepositoryMockRecorder) Exists(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockUsedKeysRepository)(nil).Exists), arg0, arg1)
}

// Store mocks base method.
func (m *MockUsedKeysRepository) Store(arg0 context.Context, arg1 string, arg2 time.Duration) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", arg0, arg1, arg2)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Store indicates an expected call of Store.
func (mr *MockUsedKeysRepositoryMockRecorder) Store(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockUsedKeysRepository)(nil).Store), arg0, arg1, arg2)
}
