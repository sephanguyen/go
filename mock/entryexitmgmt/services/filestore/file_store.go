// Code generated by mockery. DO NOT EDIT.

// This file can be generated by running: make gen-mock-repo

package mock_filestore

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// FileStore is an autogenerated mock type for the FileStore type
type FileStore struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *FileStore) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetDownloadURL provides a mock function with given fields: objectName
func (_m *FileStore) GetDownloadURL(objectName string) string {
	ret := _m.Called(objectName)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(objectName)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// UploadFromFile provides a mock function with given fields: ctx, objectName, pathName, contentType
func (_m *FileStore) UploadFromFile(ctx context.Context, objectName string, pathName string, contentType string) error {
	ret := _m.Called(ctx, objectName, pathName, contentType)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(ctx, objectName, pathName, contentType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewFileStore interface {
	mock.TestingT
	Cleanup(func())
}

// NewFileStore creates a new instance of FileStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFileStore(t mockConstructorTestingTNewFileStore) *FileStore {
	mock := &FileStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
