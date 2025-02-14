// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/payment/entities"

	mock "github.com/stretchr/testify/mock"

	utils "github.com/manabie-com/backend/internal/payment/utils"
)

// IBillingScheduleServiceForRecurringBilling is an autogenerated mock type for the IBillingScheduleServiceForRecurringBilling type
type IBillingScheduleServiceForRecurringBilling struct {
	mock.Mock
}

// CheckScheduleReturnProRatedItemAndMapPeriodInfo provides a mock function with given fields: ctx, db, orderItemData
func (_m *IBillingScheduleServiceForRecurringBilling) CheckScheduleReturnProRatedItemAndMapPeriodInfo(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData) (utils.BillingItemData, entities.BillingRatio, []utils.BillingItemData, map[string]entities.BillingSchedulePeriod, error) {
	ret := _m.Called(ctx, db, orderItemData)

	var r0 utils.BillingItemData
	var r1 entities.BillingRatio
	var r2 []utils.BillingItemData
	var r3 map[string]entities.BillingSchedulePeriod
	var r4 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, utils.OrderItemData) (utils.BillingItemData, entities.BillingRatio, []utils.BillingItemData, map[string]entities.BillingSchedulePeriod, error)); ok {
		return rf(ctx, db, orderItemData)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, utils.OrderItemData) utils.BillingItemData); ok {
		r0 = rf(ctx, db, orderItemData)
	} else {
		r0 = ret.Get(0).(utils.BillingItemData)
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, utils.OrderItemData) entities.BillingRatio); ok {
		r1 = rf(ctx, db, orderItemData)
	} else {
		r1 = ret.Get(1).(entities.BillingRatio)
	}

	if rf, ok := ret.Get(2).(func(context.Context, database.QueryExecer, utils.OrderItemData) []utils.BillingItemData); ok {
		r2 = rf(ctx, db, orderItemData)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).([]utils.BillingItemData)
		}
	}

	if rf, ok := ret.Get(3).(func(context.Context, database.QueryExecer, utils.OrderItemData) map[string]entities.BillingSchedulePeriod); ok {
		r3 = rf(ctx, db, orderItemData)
	} else {
		if ret.Get(3) != nil {
			r3 = ret.Get(3).(map[string]entities.BillingSchedulePeriod)
		}
	}

	if rf, ok := ret.Get(4).(func(context.Context, database.QueryExecer, utils.OrderItemData) error); ok {
		r4 = rf(ctx, db, orderItemData)
	} else {
		r4 = ret.Error(4)
	}

	return r0, r1, r2, r3, r4
}

type mockConstructorTestingTNewIBillingScheduleServiceForRecurringBilling interface {
	mock.TestingT
	Cleanup(func())
}

// NewIBillingScheduleServiceForRecurringBilling creates a new instance of IBillingScheduleServiceForRecurringBilling. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIBillingScheduleServiceForRecurringBilling(t mockConstructorTestingTNewIBillingScheduleServiceForRecurringBilling) *IBillingScheduleServiceForRecurringBilling {
	mock := &IBillingScheduleServiceForRecurringBilling{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
