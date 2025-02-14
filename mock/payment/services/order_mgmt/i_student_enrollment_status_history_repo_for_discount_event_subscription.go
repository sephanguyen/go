// Code generated by mockery v2.32.0. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/payment/entities"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// IStudentEnrollmentStatusHistoryRepoForDiscountEventSubscription is an autogenerated mock type for the IStudentEnrollmentStatusHistoryRepoForDiscountEventSubscription type
type IStudentEnrollmentStatusHistoryRepoForDiscountEventSubscription struct {
	mock.Mock
}

// GetListEnrolledStatusByStudentIDAndTime provides a mock function with given fields: ctx, db, StudentID, time2
func (_m *IStudentEnrollmentStatusHistoryRepoForDiscountEventSubscription) GetListEnrolledStatusByStudentIDAndTime(ctx context.Context, db database.QueryExecer, StudentID string, time2 time.Time) ([]*entities.StudentEnrollmentStatusHistory, error) {
	ret := _m.Called(ctx, db, StudentID, time2)

	var r0 []*entities.StudentEnrollmentStatusHistory
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string, time.Time) ([]*entities.StudentEnrollmentStatusHistory, error)); ok {
		return rf(ctx, db, StudentID, time2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string, time.Time) []*entities.StudentEnrollmentStatusHistory); ok {
		r0 = rf(ctx, db, StudentID, time2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*entities.StudentEnrollmentStatusHistory)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, string, time.Time) error); ok {
		r1 = rf(ctx, db, StudentID, time2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewIStudentEnrollmentStatusHistoryRepoForDiscountEventSubscription creates a new instance of IStudentEnrollmentStatusHistoryRepoForDiscountEventSubscription. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIStudentEnrollmentStatusHistoryRepoForDiscountEventSubscription(t interface {
	mock.TestingT
	Cleanup(func())
}) *IStudentEnrollmentStatusHistoryRepoForDiscountEventSubscription {
	mock := &IStudentEnrollmentStatusHistoryRepoForDiscountEventSubscription{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
