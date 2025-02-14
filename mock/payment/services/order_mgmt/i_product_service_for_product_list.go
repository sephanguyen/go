// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/payment/entities"

	mock "github.com/stretchr/testify/mock"

	pmpb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

// IProductServiceForProductList is an autogenerated mock type for the IProductServiceForProductList type
type IProductServiceForProductList struct {
	mock.Mock
}

// GetGradeIDsByProductID provides a mock function with given fields: ctx, db, productID
func (_m *IProductServiceForProductList) GetGradeIDsByProductID(ctx context.Context, db database.QueryExecer, productID string) ([]string, error) {
	ret := _m.Called(ctx, db, productID)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string) ([]string, error)); ok {
		return rf(ctx, db, productID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string) []string); ok {
		r0 = rf(ctx, db, productID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, string) error); ok {
		r1 = rf(ctx, db, productID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGradeNamesByIDs provides a mock function with given fields: ctx, db, gradeIDs
func (_m *IProductServiceForProductList) GetGradeNamesByIDs(ctx context.Context, db database.Ext, gradeIDs []string) ([]string, error) {
	ret := _m.Called(ctx, db, gradeIDs)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.Ext, []string) ([]string, error)); ok {
		return rf(ctx, db, gradeIDs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.Ext, []string) []string); ok {
		r0 = rf(ctx, db, gradeIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.Ext, []string) error); ok {
		r1 = rf(ctx, db, gradeIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetListOfProductsByFilter provides a mock function with given fields: ctx, db, req, from, limit
func (_m *IProductServiceForProductList) GetListOfProductsByFilter(ctx context.Context, db database.QueryExecer, req *pmpb.RetrieveListOfProductsRequest, from int64, limit int64) ([]entities.Product, error) {
	ret := _m.Called(ctx, db, req, from, limit)

	var r0 []entities.Product
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, *pmpb.RetrieveListOfProductsRequest, int64, int64) ([]entities.Product, error)); ok {
		return rf(ctx, db, req, from, limit)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, *pmpb.RetrieveListOfProductsRequest, int64, int64) []entities.Product); ok {
		r0 = rf(ctx, db, req, from, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entities.Product)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, *pmpb.RetrieveListOfProductsRequest, int64, int64) error); ok {
		r1 = rf(ctx, db, req, from, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLocationIDsWithProductID provides a mock function with given fields: ctx, db, productID
func (_m *IProductServiceForProductList) GetLocationIDsWithProductID(ctx context.Context, db database.QueryExecer, productID string) ([]string, error) {
	ret := _m.Called(ctx, db, productID)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string) ([]string, error)); ok {
		return rf(ctx, db, productID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string) []string); ok {
		r0 = rf(ctx, db, productID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, string) error); ok {
		r1 = rf(ctx, db, productID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetProductStatsByFilter provides a mock function with given fields: ctx, db, req
func (_m *IProductServiceForProductList) GetProductStatsByFilter(ctx context.Context, db database.QueryExecer, req *pmpb.RetrieveListOfProductsRequest) (entities.ProductStats, error) {
	ret := _m.Called(ctx, db, req)

	var r0 entities.ProductStats
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, *pmpb.RetrieveListOfProductsRequest) (entities.ProductStats, error)); ok {
		return rf(ctx, db, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, *pmpb.RetrieveListOfProductsRequest) entities.ProductStats); ok {
		r0 = rf(ctx, db, req)
	} else {
		r0 = ret.Get(0).(entities.ProductStats)
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, *pmpb.RetrieveListOfProductsRequest) error); ok {
		r1 = rf(ctx, db, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetProductTypeByProductID provides a mock function with given fields: ctx, db, productID, currentProductType
func (_m *IProductServiceForProductList) GetProductTypeByProductID(ctx context.Context, db database.QueryExecer, productID string, currentProductType string) (pmpb.ProductSpecificType, error) {
	ret := _m.Called(ctx, db, productID, currentProductType)

	var r0 pmpb.ProductSpecificType
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string, string) (pmpb.ProductSpecificType, error)); ok {
		return rf(ctx, db, productID, currentProductType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string, string) pmpb.ProductSpecificType); ok {
		r0 = rf(ctx, db, productID, currentProductType)
	} else {
		r0 = ret.Get(0).(pmpb.ProductSpecificType)
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, string, string) error); ok {
		r1 = rf(ctx, db, productID, currentProductType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewIProductServiceForProductList interface {
	mock.TestingT
	Cleanup(func())
}

// NewIProductServiceForProductList creates a new instance of IProductServiceForProductList. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIProductServiceForProductList(t mockConstructorTestingTNewIProductServiceForProductList) *IProductServiceForProductList {
	mock := &IProductServiceForProductList{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
