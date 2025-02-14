// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import (
	context "context"

	database "github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/payment/entities"

	mock "github.com/stretchr/testify/mock"

	payload "github.com/manabie-com/backend/internal/golibs/kafka/payload"

	utils "github.com/manabie-com/backend/internal/payment/utils"
)

// ISubscriptionServiceForVoidOrder is an autogenerated mock type for the ISubscriptionServiceForVoidOrder type
type ISubscriptionServiceForVoidOrder struct {
	mock.Mock
}

// Publish provides a mock function with given fields: ctx, db, message
func (_m *ISubscriptionServiceForVoidOrder) Publish(ctx context.Context, db database.QueryExecer, message utils.MessageSyncData) error {
	ret := _m.Called(ctx, db, message)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, utils.MessageSyncData) error); ok {
		r0 = rf(ctx, db, message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ToNotificationMessage provides a mock function with given fields: ctx, tx, order, student, upsertNotificationData
func (_m *ISubscriptionServiceForVoidOrder) ToNotificationMessage(ctx context.Context, tx database.QueryExecer, order entities.Order, student entities.Student, upsertNotificationData utils.UpsertSystemNotificationData) (*payload.UpsertSystemNotification, error) {
	ret := _m.Called(ctx, tx, order, student, upsertNotificationData)

	var r0 *payload.UpsertSystemNotification
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, entities.Order, entities.Student, utils.UpsertSystemNotificationData) (*payload.UpsertSystemNotification, error)); ok {
		return rf(ctx, tx, order, student, upsertNotificationData)
	}
	if rf, ok := ret.Get(0).(func(context.Context, database.QueryExecer, entities.Order, entities.Student, utils.UpsertSystemNotificationData) *payload.UpsertSystemNotification); ok {
		r0 = rf(ctx, tx, order, student, upsertNotificationData)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*payload.UpsertSystemNotification)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, database.QueryExecer, entities.Order, entities.Student, utils.UpsertSystemNotificationData) error); ok {
		r1 = rf(ctx, tx, order, student, upsertNotificationData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewISubscriptionServiceForVoidOrder interface {
	mock.TestingT
	Cleanup(func())
}

// NewISubscriptionServiceForVoidOrder creates a new instance of ISubscriptionServiceForVoidOrder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewISubscriptionServiceForVoidOrder(t mockConstructorTestingTNewISubscriptionServiceForVoidOrder) *ISubscriptionServiceForVoidOrder {
	mock := &ISubscriptionServiceForVoidOrder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
