// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/payment/entities"

	mock "github.com/stretchr/testify/mock"

	utils "github.com/manabie-com/backend/internal/payment/utils"
)

// IBillItemServiceForRecurringBilling is an autogenerated mock type for the IBillItemServiceForRecurringBilling type
type IBillItemServiceForRecurringBilling struct {
	mock.Mock
}

type IBillItemServiceForRecurringBilling_Expecter struct {
	mock *mock.Mock
}

func (_m *IBillItemServiceForRecurringBilling) EXPECT() *IBillItemServiceForRecurringBilling_Expecter {
	return &IBillItemServiceForRecurringBilling_Expecter{mock: &_m.Mock}
}

// CreateCancelBillItemForRecurringBilling provides a mock function with given fields: ctx, db, orderItemData, proRatedBillItem, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, mapOldBillingItem
func (_m *IBillItemServiceForRecurringBilling) CreateCancelBillItemForRecurringBilling(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData, proRatedBillItem utils.BillingItemData, ratioOfProRatedBillingItem entities.BillingRatio, normalBillItem []utils.BillingItemData, mapPeriodInfo map[string]entities.BillingSchedulePeriod, mapOldBillingItem map[string]entities.BillItem) error {
	ret := _m.Called(ctx, db, orderItemData, proRatedBillItem, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, mapOldBillingItem)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, utils.OrderItemData, utils.BillingItemData, entities.BillingRatio, []utils.BillingItemData, map[string]entities.BillingSchedulePeriod, map[string]entities.BillItem) error); ok {
		r0 = rf(ctx, db, orderItemData, proRatedBillItem, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, mapOldBillingItem)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IBillItemServiceForRecurringBilling_CreateCancelBillItemForRecurringBilling_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateCancelBillItemForRecurringBilling'
type IBillItemServiceForRecurringBilling_CreateCancelBillItemForRecurringBilling_Call struct {
	*mock.Call
}

// CreateCancelBillItemForRecurringBilling is a helper method to define mock.On call
//   - ctx context.Context
//   - db database.QueryExecer
//   - orderItemData utils.OrderItemData
//   - proRatedBillItem utils.BillingItemData
//   - ratioOfProRatedBillingItem entities.BillingRatio
//   - normalBillItem []utils.BillingItemData
//   - mapPeriodInfo map[string]entities.BillingSchedulePeriod
//   - mapOldBillingItem map[string]entities.BillItem
func (_e *IBillItemServiceForRecurringBilling_Expecter) CreateCancelBillItemForRecurringBilling(ctx interface{}, db interface{}, orderItemData interface{}, proRatedBillItem interface{}, ratioOfProRatedBillingItem interface{}, normalBillItem interface{}, mapPeriodInfo interface{}, mapOldBillingItem interface{}) *IBillItemServiceForRecurringBilling_CreateCancelBillItemForRecurringBilling_Call {
	return &IBillItemServiceForRecurringBilling_CreateCancelBillItemForRecurringBilling_Call{Call: _e.mock.On("CreateCancelBillItemForRecurringBilling", ctx, db, orderItemData, proRatedBillItem, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, mapOldBillingItem)}
}

func (_c *IBillItemServiceForRecurringBilling_CreateCancelBillItemForRecurringBilling_Call) Run(run func(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData, proRatedBillItem utils.BillingItemData, ratioOfProRatedBillingItem entities.BillingRatio, normalBillItem []utils.BillingItemData, mapPeriodInfo map[string]entities.BillingSchedulePeriod, mapOldBillingItem map[string]entities.BillItem)) *IBillItemServiceForRecurringBilling_CreateCancelBillItemForRecurringBilling_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(database.QueryExecer), args[2].(utils.OrderItemData), args[3].(utils.BillingItemData), args[4].(entities.BillingRatio), args[5].([]utils.BillingItemData), args[6].(map[string]entities.BillingSchedulePeriod), args[7].(map[string]entities.BillItem))
	})
	return _c
}

func (_c *IBillItemServiceForRecurringBilling_CreateCancelBillItemForRecurringBilling_Call) Return(err error) *IBillItemServiceForRecurringBilling_CreateCancelBillItemForRecurringBilling_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *IBillItemServiceForRecurringBilling_CreateCancelBillItemForRecurringBilling_Call) RunAndReturn(run func(context.Context, database.QueryExecer, utils.OrderItemData, utils.BillingItemData, entities.BillingRatio, []utils.BillingItemData, map[string]entities.BillingSchedulePeriod, map[string]entities.BillItem) error) *IBillItemServiceForRecurringBilling_CreateCancelBillItemForRecurringBilling_Call {
	_c.Call.Return(run)
	return _c
}

// CreateNewBillItemForRecurringBilling provides a mock function with given fields: ctx, db, orderItemData, proRatedBillItem, proRatedPrice, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, discountName
func (_m *IBillItemServiceForRecurringBilling) CreateNewBillItemForRecurringBilling(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData, proRatedBillItem utils.BillingItemData, proRatedPrice entities.ProductPrice, ratioOfProRatedBillingItem entities.BillingRatio, normalBillItem []utils.BillingItemData, mapPeriodInfo map[string]entities.BillingSchedulePeriod, discountName string) error {
	ret := _m.Called(ctx, db, orderItemData, proRatedBillItem, proRatedPrice, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, discountName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, utils.OrderItemData, utils.BillingItemData, entities.ProductPrice, entities.BillingRatio, []utils.BillingItemData, map[string]entities.BillingSchedulePeriod, string) error); ok {
		r0 = rf(ctx, db, orderItemData, proRatedBillItem, proRatedPrice, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, discountName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IBillItemServiceForRecurringBilling_CreateNewBillItemForRecurringBilling_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateNewBillItemForRecurringBilling'
type IBillItemServiceForRecurringBilling_CreateNewBillItemForRecurringBilling_Call struct {
	*mock.Call
}

// CreateNewBillItemForRecurringBilling is a helper method to define mock.On call
//   - ctx context.Context
//   - db database.QueryExecer
//   - orderItemData utils.OrderItemData
//   - proRatedBillItem utils.BillingItemData
//   - proRatedPrice entities.ProductPrice
//   - ratioOfProRatedBillingItem entities.BillingRatio
//   - normalBillItem []utils.BillingItemData
//   - mapPeriodInfo map[string]entities.BillingSchedulePeriod
//   - discountName string
func (_e *IBillItemServiceForRecurringBilling_Expecter) CreateNewBillItemForRecurringBilling(ctx interface{}, db interface{}, orderItemData interface{}, proRatedBillItem interface{}, proRatedPrice interface{}, ratioOfProRatedBillingItem interface{}, normalBillItem interface{}, mapPeriodInfo interface{}, discountName interface{}) *IBillItemServiceForRecurringBilling_CreateNewBillItemForRecurringBilling_Call {
	return &IBillItemServiceForRecurringBilling_CreateNewBillItemForRecurringBilling_Call{Call: _e.mock.On("CreateNewBillItemForRecurringBilling", ctx, db, orderItemData, proRatedBillItem, proRatedPrice, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, discountName)}
}

func (_c *IBillItemServiceForRecurringBilling_CreateNewBillItemForRecurringBilling_Call) Run(run func(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData, proRatedBillItem utils.BillingItemData, proRatedPrice entities.ProductPrice, ratioOfProRatedBillingItem entities.BillingRatio, normalBillItem []utils.BillingItemData, mapPeriodInfo map[string]entities.BillingSchedulePeriod, discountName string)) *IBillItemServiceForRecurringBilling_CreateNewBillItemForRecurringBilling_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(database.QueryExecer), args[2].(utils.OrderItemData), args[3].(utils.BillingItemData), args[4].(entities.ProductPrice), args[5].(entities.BillingRatio), args[6].([]utils.BillingItemData), args[7].(map[string]entities.BillingSchedulePeriod), args[8].(string))
	})
	return _c
}

func (_c *IBillItemServiceForRecurringBilling_CreateNewBillItemForRecurringBilling_Call) Return(err error) *IBillItemServiceForRecurringBilling_CreateNewBillItemForRecurringBilling_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *IBillItemServiceForRecurringBilling_CreateNewBillItemForRecurringBilling_Call) RunAndReturn(run func(context.Context, database.QueryExecer, utils.OrderItemData, utils.BillingItemData, entities.ProductPrice, entities.BillingRatio, []utils.BillingItemData, map[string]entities.BillingSchedulePeriod, string) error) *IBillItemServiceForRecurringBilling_CreateNewBillItemForRecurringBilling_Call {
	_c.Call.Return(run)
	return _c
}

// CreateUpdateBillItemForRecurringBilling provides a mock function with given fields: ctx, db, orderItemData, proRatedBillItem, proRatedPrice, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, mapOldBillingItem, discountName
func (_m *IBillItemServiceForRecurringBilling) CreateUpdateBillItemForRecurringBilling(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData, proRatedBillItem utils.BillingItemData, proRatedPrice entities.ProductPrice, ratioOfProRatedBillingItem entities.BillingRatio, normalBillItem []utils.BillingItemData, mapPeriodInfo map[string]entities.BillingSchedulePeriod, mapOldBillingItem map[string]entities.BillItem, discountName string) error {
	ret := _m.Called(ctx, db, orderItemData, proRatedBillItem, proRatedPrice, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, mapOldBillingItem, discountName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, utils.OrderItemData, utils.BillingItemData, entities.ProductPrice, entities.BillingRatio, []utils.BillingItemData, map[string]entities.BillingSchedulePeriod, map[string]entities.BillItem, string) error); ok {
		r0 = rf(ctx, db, orderItemData, proRatedBillItem, proRatedPrice, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, mapOldBillingItem, discountName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IBillItemServiceForRecurringBilling_CreateUpdateBillItemForRecurringBilling_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateUpdateBillItemForRecurringBilling'
type IBillItemServiceForRecurringBilling_CreateUpdateBillItemForRecurringBilling_Call struct {
	*mock.Call
}

// CreateUpdateBillItemForRecurringBilling is a helper method to define mock.On call
//   - ctx context.Context
//   - db database.QueryExecer
//   - orderItemData utils.OrderItemData
//   - proRatedBillItem utils.BillingItemData
//   - proRatedPrice entities.ProductPrice
//   - ratioOfProRatedBillingItem entities.BillingRatio
//   - normalBillItem []utils.BillingItemData
//   - mapPeriodInfo map[string]entities.BillingSchedulePeriod
//   - mapOldBillingItem map[string]entities.BillItem
//   - discountName string
func (_e *IBillItemServiceForRecurringBilling_Expecter) CreateUpdateBillItemForRecurringBilling(ctx interface{}, db interface{}, orderItemData interface{}, proRatedBillItem interface{}, proRatedPrice interface{}, ratioOfProRatedBillingItem interface{}, normalBillItem interface{}, mapPeriodInfo interface{}, mapOldBillingItem interface{}, discountName interface{}) *IBillItemServiceForRecurringBilling_CreateUpdateBillItemForRecurringBilling_Call {
	return &IBillItemServiceForRecurringBilling_CreateUpdateBillItemForRecurringBilling_Call{Call: _e.mock.On("CreateUpdateBillItemForRecurringBilling", ctx, db, orderItemData, proRatedBillItem, proRatedPrice, ratioOfProRatedBillingItem, normalBillItem, mapPeriodInfo, mapOldBillingItem, discountName)}
}

func (_c *IBillItemServiceForRecurringBilling_CreateUpdateBillItemForRecurringBilling_Call) Run(run func(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData, proRatedBillItem utils.BillingItemData, proRatedPrice entities.ProductPrice, ratioOfProRatedBillingItem entities.BillingRatio, normalBillItem []utils.BillingItemData, mapPeriodInfo map[string]entities.BillingSchedulePeriod, mapOldBillingItem map[string]entities.BillItem, discountName string)) *IBillItemServiceForRecurringBilling_CreateUpdateBillItemForRecurringBilling_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(database.QueryExecer), args[2].(utils.OrderItemData), args[3].(utils.BillingItemData), args[4].(entities.ProductPrice), args[5].(entities.BillingRatio), args[6].([]utils.BillingItemData), args[7].(map[string]entities.BillingSchedulePeriod), args[8].(map[string]entities.BillItem), args[9].(string))
	})
	return _c
}

func (_c *IBillItemServiceForRecurringBilling_CreateUpdateBillItemForRecurringBilling_Call) Return(err error) *IBillItemServiceForRecurringBilling_CreateUpdateBillItemForRecurringBilling_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *IBillItemServiceForRecurringBilling_CreateUpdateBillItemForRecurringBilling_Call) RunAndReturn(run func(context.Context, database.QueryExecer, utils.OrderItemData, utils.BillingItemData, entities.ProductPrice, entities.BillingRatio, []utils.BillingItemData, map[string]entities.BillingSchedulePeriod, map[string]entities.BillItem, string) error) *IBillItemServiceForRecurringBilling_CreateUpdateBillItemForRecurringBilling_Call {
	_c.Call.Return(run)
	return _c
}

// GetMapOldBillingItemForRecurringBilling provides a mock function with given fields: ctx, db, orderItemData, mapPeriodInfo
func (_m *IBillItemServiceForRecurringBilling) GetMapOldBillingItemForRecurringBilling(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData, mapPeriodInfo map[string]entities.BillingSchedulePeriod) (map[string]entities.BillItem, error) {
	ret := _m.Called(ctx, db, orderItemData, mapPeriodInfo)

	var r0 map[string]entities.BillItem
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, utils.OrderItemData, map[string]entities.BillingSchedulePeriod) (map[string]entities.BillItem, error)); ok {
		return rf(ctx, db, orderItemData, mapPeriodInfo)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, utils.OrderItemData, map[string]entities.BillingSchedulePeriod) map[string]entities.BillItem); ok {
		r0 = rf(ctx, db, orderItemData, mapPeriodInfo)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]entities.BillItem)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, utils.OrderItemData, map[string]entities.BillingSchedulePeriod) error); ok {
		r1 = rf(ctx, db, orderItemData, mapPeriodInfo)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IBillItemServiceForRecurringBilling_GetMapOldBillingItemForRecurringBilling_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetMapOldBillingItemForRecurringBilling'
type IBillItemServiceForRecurringBilling_GetMapOldBillingItemForRecurringBilling_Call struct {
	*mock.Call
}

// GetMapOldBillingItemForRecurringBilling is a helper method to define mock.On call
//   - ctx context.Context
//   - db database.QueryExecer
//   - orderItemData utils.OrderItemData
//   - mapPeriodInfo map[string]entities.BillingSchedulePeriod
func (_e *IBillItemServiceForRecurringBilling_Expecter) GetMapOldBillingItemForRecurringBilling(ctx interface{}, db interface{}, orderItemData interface{}, mapPeriodInfo interface{}) *IBillItemServiceForRecurringBilling_GetMapOldBillingItemForRecurringBilling_Call {
	return &IBillItemServiceForRecurringBilling_GetMapOldBillingItemForRecurringBilling_Call{Call: _e.mock.On("GetMapOldBillingItemForRecurringBilling", ctx, db, orderItemData, mapPeriodInfo)}
}

func (_c *IBillItemServiceForRecurringBilling_GetMapOldBillingItemForRecurringBilling_Call) Run(run func(ctx context.Context, db database.QueryExecer, orderItemData utils.OrderItemData, mapPeriodInfo map[string]entities.BillingSchedulePeriod)) *IBillItemServiceForRecurringBilling_GetMapOldBillingItemForRecurringBilling_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(database.QueryExecer), args[2].(utils.OrderItemData), args[3].(map[string]entities.BillingSchedulePeriod))
	})
	return _c
}

func (_c *IBillItemServiceForRecurringBilling_GetMapOldBillingItemForRecurringBilling_Call) Return(mapOldBillingItem map[string]entities.BillItem, err error) *IBillItemServiceForRecurringBilling_GetMapOldBillingItemForRecurringBilling_Call {
	_c.Call.Return(mapOldBillingItem, err)
	return _c
}

func (_c *IBillItemServiceForRecurringBilling_GetMapOldBillingItemForRecurringBilling_Call) RunAndReturn(run func(context.Context, database.QueryExecer, utils.OrderItemData, map[string]entities.BillingSchedulePeriod) (map[string]entities.BillItem, error)) *IBillItemServiceForRecurringBilling_GetMapOldBillingItemForRecurringBilling_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewIBillItemServiceForRecurringBilling interface {
	mock.TestingT
	Cleanup(func())
}

// NewIBillItemServiceForRecurringBilling creates a new instance of IBillItemServiceForRecurringBilling. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIBillItemServiceForRecurringBilling(t mockConstructorTestingTNewIBillItemServiceForRecurringBilling) *IBillItemServiceForRecurringBilling {
	mock := &IBillItemServiceForRecurringBilling{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
