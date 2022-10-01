// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/anzx/fabric-commandcentre-sdk/pkg/sdk (interfaces: Publisher)

// Package commandcentre is a generated GoMock package.
package commandcentre

import (
	context "context"
	reflect "reflect"

	sdk "github.com/anzx/fabric-commandcentre-sdk/pkg/sdk"
	gomock "github.com/golang/mock/gomock"
)

// MockPublisher is a mock of Publisher interface.
type MockPublisher struct {
	ctrl     *gomock.Controller
	recorder *MockPublisherMockRecorder
}

// MockPublisherMockRecorder is the mock recorder for MockPublisher.
type MockPublisherMockRecorder struct {
	mock *MockPublisher
}

// NewMockPublisher creates a new mock instance.
func NewMockPublisher(ctrl *gomock.Controller) *MockPublisher {
	mock := &MockPublisher{ctrl: ctrl}
	mock.recorder = &MockPublisherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPublisher) EXPECT() *MockPublisherMockRecorder {
	return m.recorder
}

// Publish mocks base method.
func (m *MockPublisher) Publish(arg0 context.Context, arg1 sdk.PublishRequester) (*sdk.PublishResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Publish", arg0, arg1)
	ret0, _ := ret[0].(*sdk.PublishResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Publish indicates an expected call of Publish.
func (mr *MockPublisherMockRecorder) Publish(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockPublisher)(nil).Publish), arg0, arg1)
}

// PublishSync mocks base method.
func (m *MockPublisher) PublishSync(arg0 context.Context, arg1 *sdk.PublishSyncRequest) (*sdk.PublishSyncResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PublishSync", arg0, arg1)
	ret0, _ := ret[0].(*sdk.PublishSyncResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PublishSync indicates an expected call of PublishSync.
func (mr *MockPublisherMockRecorder) PublishSync(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublishSync", reflect.TypeOf((*MockPublisher)(nil).PublishSync), arg0, arg1)
}
