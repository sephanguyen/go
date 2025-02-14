// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/payment/entities"

	mock "github.com/stretchr/testify/mock"
)

// IStudentServiceForVoidOrder is an autogenerated mock type for the IStudentServiceForVoidOrder type
type IStudentServiceForVoidOrder struct {
	mock.Mock
}

// GetStudentAndNameByID provides a mock function with given fields: ctx, db, studentID
func (_m *IStudentServiceForVoidOrder) GetStudentAndNameByID(ctx context.Context, db database.QueryExecer, studentID string) (entities.Student, string, error) {
	ret := _m.Called(ctx, db, studentID)

	var r0 entities.Student
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string) (entities.Student, string, error)); ok {
		return rf(ctx, db, studentID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string) entities.Student); ok {
		r0 = rf(ctx, db, studentID)
	} else {
		r0 = ret.Get(0).(entities.Student)
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, string) string); ok {
		r1 = rf(ctx, db, studentID)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(context.Context, database.QueryExecer, string) error); ok {
		r2 = rf(ctx, db, studentID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

type mockConstructorTestingTNewIStudentServiceForVoidOrder interface {
	mock.TestingT
	Cleanup(func())
}

// NewIStudentServiceForVoidOrder creates a new instance of IStudentServiceForVoidOrder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIStudentServiceForVoidOrder(t mockConstructorTestingTNewIStudentServiceForVoidOrder) *IStudentServiceForVoidOrder {
	mock := &IStudentServiceForVoidOrder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
