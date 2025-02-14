// Code generated by mockery. DO NOT EDIT.

// This file can be generated by running: make gen-mock-repo

package mock_service

import (
	context "context"

	domain "github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	mock "github.com/stretchr/testify/mock"
)

// NotificationHandler is an autogenerated mock type for the NotificationHandler type
type NotificationHandler struct {
	mock.Mock
}

// PushNotification provides a mock function with given fields: ctx, message
func (_m *NotificationHandler) PushNotification(ctx context.Context, message *domain.OfflineMessage) error {
	ret := _m.Called(ctx, message)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *domain.OfflineMessage) error); ok {
		r0 = rf(ctx, message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewNotificationHandler interface {
	mock.TestingT
	Cleanup(func())
}

// NewNotificationHandler creates a new instance of NotificationHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewNotificationHandler(t mockConstructorTestingTNewNotificationHandler) *NotificationHandler {
	mock := &NotificationHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
