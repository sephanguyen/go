// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/payment/entities"

	mock "github.com/stretchr/testify/mock"
)

// IProductPriceRepoForDiscountEventSubscription is an autogenerated mock type for the IProductPriceRepoForDiscountEventSubscription type
type IProductPriceRepoForDiscountEventSubscription struct {
	mock.Mock
}

// GetByProductIDAndBillingSchedulePeriodIDAndPriceType provides a mock function with given fields: ctx, db, productID, billingSchedulePeriodID, priceType
func (_m *IProductPriceRepoForDiscountEventSubscription) GetByProductIDAndBillingSchedulePeriodIDAndPriceType(ctx context.Context, db database.QueryExecer, productID string, billingSchedulePeriodID string, priceType string) (entities.ProductPrice, error) {
	ret := _m.Called(ctx, db, productID, billingSchedulePeriodID, priceType)

	var r0 entities.ProductPrice
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string, string, string) (entities.ProductPrice, error)); ok {
		return rf(ctx, db, productID, billingSchedulePeriodID, priceType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string, string, string) entities.ProductPrice); ok {
		r0 = rf(ctx, db, productID, billingSchedulePeriodID, priceType)
	} else {
		r0 = ret.Get(0).(entities.ProductPrice)
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, string, string, string) error); ok {
		r1 = rf(ctx, db, productID, billingSchedulePeriodID, priceType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByProductIDAndPriceType provides a mock function with given fields: ctx, db, productID, priceType
func (_m *IProductPriceRepoForDiscountEventSubscription) GetByProductIDAndPriceType(ctx context.Context, db database.QueryExecer, productID string, priceType string) ([]entities.ProductPrice, error) {
	ret := _m.Called(ctx, db, productID, priceType)

	var r0 []entities.ProductPrice
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string, string) ([]entities.ProductPrice, error)); ok {
		return rf(ctx, db, productID, priceType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string, string) []entities.ProductPrice); ok {
		r0 = rf(ctx, db, productID, priceType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entities.ProductPrice)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, string, string) error); ok {
		r1 = rf(ctx, db, productID, priceType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByProductIDAndQuantityAndPriceType provides a mock function with given fields: ctx, db, productID, weight, priceType
func (_m *IProductPriceRepoForDiscountEventSubscription) GetByProductIDAndQuantityAndPriceType(ctx context.Context, db database.QueryExecer, productID string, weight int32, priceType string) (entities.ProductPrice, error) {
	ret := _m.Called(ctx, db, productID, weight, priceType)

	var r0 entities.ProductPrice
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string, int32, string) (entities.ProductPrice, error)); ok {
		return rf(ctx, db, productID, weight, priceType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, string, int32, string) entities.ProductPrice); ok {
		r0 = rf(ctx, db, productID, weight, priceType)
	} else {
		r0 = ret.Get(0).(entities.ProductPrice)
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, string, int32, string) error); ok {
		r1 = rf(ctx, db, productID, weight, priceType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewIProductPriceRepoForDiscountEventSubscription interface {
	mock.TestingT
	Cleanup(func())
}

// NewIProductPriceRepoForDiscountEventSubscription creates a new instance of IProductPriceRepoForDiscountEventSubscription. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIProductPriceRepoForDiscountEventSubscription(t mockConstructorTestingTNewIProductPriceRepoForDiscountEventSubscription) *IProductPriceRepoForDiscountEventSubscription {
	mock := &IProductPriceRepoForDiscountEventSubscription{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
