// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	utils "github.com/manabie-com/backend/internal/payment/utils"
)

// IElasticSearchServiceForCreateOrder is an autogenerated mock type for the IElasticSearchServiceForCreateOrder type
type IElasticSearchServiceForCreateOrder struct {
	mock.Mock
}

// InsertOrderData provides a mock function with given fields: ctx, data
func (_m *IElasticSearchServiceForCreateOrder) InsertOrderData(ctx context.Context, data utils.ElasticSearchData) error {
	ret := _m.Called(ctx, data)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, utils.ElasticSearchData) error); ok {
		r0 = rf(ctx, data)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewIElasticSearchServiceForCreateOrder interface {
	mock.TestingT
	Cleanup(func())
}

// NewIElasticSearchServiceForCreateOrder creates a new instance of IElasticSearchServiceForCreateOrder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIElasticSearchServiceForCreateOrder(t mockConstructorTestingTNewIElasticSearchServiceForCreateOrder) *IElasticSearchServiceForCreateOrder {
	mock := &IElasticSearchServiceForCreateOrder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
