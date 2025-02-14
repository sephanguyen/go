// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/payment/entities"

	mock "github.com/stretchr/testify/mock"
)

// IStudentProductServiceForVoidOrder is an autogenerated mock type for the IStudentProductServiceForVoidOrder type
type IStudentProductServiceForVoidOrder struct {
	mock.Mock
}

// GetStudentProductsByStudentProductIDs provides a mock function with given fields: ctx, db, studentProductIDs
func (_m *IStudentProductServiceForVoidOrder) GetStudentProductsByStudentProductIDs(ctx context.Context, db database.Ext, studentProductIDs []string) ([]entities.StudentProduct, error) {
	ret := _m.Called(ctx, db, studentProductIDs)

	var r0 []entities.StudentProduct
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.Ext, []string) ([]entities.StudentProduct, error)); ok {
		return rf(ctx, db, studentProductIDs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.Ext, []string) []entities.StudentProduct); ok {
		r0 = rf(ctx, db, studentProductIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entities.StudentProduct)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.Ext, []string) error); ok {
		r1 = rf(ctx, db, studentProductIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// VoidStudentProduct provides a mock function with given fields: ctx, db, studentProductID, orderType
func (_m *IStudentProductServiceForVoidOrder) VoidStudentProduct(ctx context.Context, db database.QueryExecer, studentProductID string, orderType string) (entities.StudentProduct, entities.Product, bool, error) {
	ret := _m.Called(ctx, db, studentProductID, orderType)

	var r0 entities.StudentProduct
	var r1 entities.Product
	var r2 bool
	var r3 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string, string) (entities.StudentProduct, entities.Product, bool, error)); ok {
		return rf(ctx, db, studentProductID, orderType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string, string) entities.StudentProduct); ok {
		r0 = rf(ctx, db, studentProductID, orderType)
	} else {
		r0 = ret.Get(0).(entities.StudentProduct)
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, string, string) entities.Product); ok {
		r1 = rf(ctx, db, studentProductID, orderType)
	} else {
		r1 = ret.Get(1).(entities.Product)
	}

	if rf, ok := ret.Get(2).(func(context.Context, database.QueryExecer, string, string) bool); ok {
		r2 = rf(ctx, db, studentProductID, orderType)
	} else {
		r2 = ret.Get(2).(bool)
	}

	if rf, ok := ret.Get(3).(func(context.Context, database.QueryExecer, string, string) error); ok {
		r3 = rf(ctx, db, studentProductID, orderType)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

type mockConstructorTestingTNewIStudentProductServiceForVoidOrder interface {
	mock.TestingT
	Cleanup(func())
}

// NewIStudentProductServiceForVoidOrder creates a new instance of IStudentProductServiceForVoidOrder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIStudentProductServiceForVoidOrder(t mockConstructorTestingTNewIStudentProductServiceForVoidOrder) *IStudentProductServiceForVoidOrder {
	mock := &IStudentProductServiceForVoidOrder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
