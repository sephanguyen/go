// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/payment/entities"

	mock "github.com/stretchr/testify/mock"
)

// IUpcomingBillItemServiceForVoidOrder is an autogenerated mock type for the IUpcomingBillItemServiceForVoidOrder type
type IUpcomingBillItemServiceForVoidOrder struct {
	mock.Mock
}

// VoidUpcomingBillItemsByOrder provides a mock function with given fields: ctx, db, order
func (_m *IUpcomingBillItemServiceForVoidOrder) VoidUpcomingBillItemsByOrder(ctx context.Context, db database.QueryExecer, order entities.Order) error {
	ret := _m.Called(ctx, db, order)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, entities.Order) error); ok {
		r0 = rf(ctx, db, order)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewIUpcomingBillItemServiceForVoidOrder interface {
	mock.TestingT
	Cleanup(func())
}

// NewIUpcomingBillItemServiceForVoidOrder creates a new instance of IUpcomingBillItemServiceForVoidOrder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIUpcomingBillItemServiceForVoidOrder(t mockConstructorTestingTNewIUpcomingBillItemServiceForVoidOrder) *IUpcomingBillItemServiceForVoidOrder {
	mock := &IUpcomingBillItemServiceForVoidOrder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
