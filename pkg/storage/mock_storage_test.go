// Automatically generated by MockGen. DO NOT EDIT!
// Source: ./pkg/types/storage.go

package storage

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	io "io"
)

// Mock of Storage interface
type MockStorage struct {
	ctrl     *gomock.Controller
	recorder *_MockStorageRecorder
}

// Recorder for MockStorage (not exported)
type _MockStorageRecorder struct {
	mock *MockStorage
}

func NewMockStorage(ctrl *gomock.Controller) *MockStorage {
	mock := &MockStorage{ctrl: ctrl}
	mock.recorder = &_MockStorageRecorder{mock}
	return mock
}

func (_m *MockStorage) EXPECT() *_MockStorageRecorder {
	return _m.recorder
}

func (_m *MockStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	ret := _m.ctrl.Call(_m, "Get", ctx, key)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockStorageRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Get", arg0, arg1)
}

func (_m *MockStorage) Exists(ctx context.Context, key string) (bool, error) {
	ret := _m.ctrl.Call(_m, "Exists", ctx, key)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockStorageRecorder) Exists(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Exists", arg0, arg1)
}

func (_m *MockStorage) Store(ctx context.Context, key string, data io.ReadCloser) error {
	ret := _m.ctrl.Call(_m, "Store", ctx, key, data)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockStorageRecorder) Store(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Store", arg0, arg1, arg2)
}
