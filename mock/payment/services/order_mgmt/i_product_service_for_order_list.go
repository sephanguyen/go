// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/payment/entities"

	mock "github.com/stretchr/testify/mock"
)

// IProductServiceForOrderList is an autogenerated mock type for the IProductServiceForOrderList type
type IProductServiceForOrderList struct {
	mock.Mock
}

// GetProductsByIDs provides a mock function with given fields: ctx, db, productIDs
func (_m *IProductServiceForOrderList) GetProductsByIDs(ctx context.Context, db database.Ext, productIDs []string) ([]entities.Product, error) {
	ret := _m.Called(ctx, db, productIDs)

	var r0 []entities.Product
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.Ext, []string) ([]entities.Product, error)); ok {
		return rf(ctx, db, productIDs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.Ext, []string) []entities.Product); ok {
		r0 = rf(ctx, db, productIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entities.Product)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.Ext, []string) error); ok {
		r1 = rf(ctx, db, productIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewIProductServiceForOrderList interface {
	mock.TestingT
	Cleanup(func())
}

// NewIProductServiceForOrderList creates a new instance of IProductServiceForOrderList. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIProductServiceForOrderList(t mockConstructorTestingTNewIProductServiceForOrderList) *IProductServiceForOrderList {
	mock := &IProductServiceForOrderList{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
