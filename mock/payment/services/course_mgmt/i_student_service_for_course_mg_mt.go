// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	mock "github.com/stretchr/testify/mock"
)

// IStudentServiceForCourseMgMt is an autogenerated mock type for the IStudentServiceForCourseMgMt type
type IStudentServiceForCourseMgMt struct {
	mock.Mock
}

// GetMapLocationAccessStudentByStudentIDs provides a mock function with given fields: ctx, db, studentIDs
func (_m *IStudentServiceForCourseMgMt) GetMapLocationAccessStudentByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) (map[string]interface{}, error) {
	ret := _m.Called(ctx, db, studentIDs)

	var r0 map[string]interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, []string) (map[string]interface{}, error)); ok {
		return rf(ctx, db, studentIDs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, []string) map[string]interface{}); ok {
		r0 = rf(ctx, db, studentIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, []string) error); ok {
		r1 = rf(ctx, db, studentIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewIStudentServiceForCourseMgMt interface {
	mock.TestingT
	Cleanup(func())
}

// NewIStudentServiceForCourseMgMt creates a new instance of IStudentServiceForCourseMgMt. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIStudentServiceForCourseMgMt(t mockConstructorTestingTNewIStudentServiceForCourseMgMt) *IStudentServiceForCourseMgMt {
	mock := &IStudentServiceForCourseMgMt{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
