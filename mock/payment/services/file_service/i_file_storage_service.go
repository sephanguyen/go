// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"

	io "io"

	mock "github.com/stretchr/testify/mock"
)

// IFileStorageService is an autogenerated mock type for the IFileStorageService type
type IFileStorageService struct {
	mock.Mock
}

// GetDownloadFileByName provides a mock function with given fields: ctx, db, fileName
func (_m *IFileStorageService) GetDownloadFileByName(ctx context.Context, db database.QueryExecer, fileName string) (string, error) {
	ret := _m.Called(ctx, db, fileName)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string) (string, error)); ok {
		return rf(ctx, db, fileName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string) string); ok {
		r0 = rf(ctx, db, fileName)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, string) error); ok {
		r1 = rf(ctx, db, fileName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UploadFile provides a mock function with given fields: ctx, reader, db, fileName, fileType, fileSize
func (_m *IFileStorageService) UploadFile(ctx context.Context, reader io.Reader, db database.QueryExecer, fileName string, fileType string, fileSize int64) (string, error) {
	ret := _m.Called(ctx, reader, db, fileName, fileType, fileSize)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, io.Reader, database.QueryExecer, string, string, int64) (string, error)); ok {
		return rf(ctx, reader, db, fileName, fileType, fileSize)
	}
	if rf, ok := ret.Get(0).(func(context.Context, io.Reader, database.QueryExecer, string, string, int64) string); ok {
		r0 = rf(ctx, reader, db, fileName, fileType, fileSize)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, io.Reader, database.QueryExecer, string, string, int64) error); ok {
		r1 = rf(ctx, reader, db, fileName, fileType, fileSize)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewIFileStorageService interface {
	mock.TestingT
	Cleanup(func())
}

// NewIFileStorageService creates a new instance of IFileStorageService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIFileStorageService(t mockConstructorTestingTNewIFileStorageService) *IFileStorageService {
	mock := &IFileStorageService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
