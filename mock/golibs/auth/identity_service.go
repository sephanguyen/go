// Code generated by mockery. DO NOT EDIT.

// This file can be generated by running: make gen-mock-repo

package mock_auth

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// IdentityService is an autogenerated mock type for the IdentityService type
type IdentityService struct {
	mock.Mock
}

// VerifyEmailPassword provides a mock function with given fields: ctx, email, password
func (_m *IdentityService) VerifyEmailPassword(ctx context.Context, email string, password string) (string, error) {
	ret := _m.Called(ctx, email, password)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = rf(ctx, email, password)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, email, password)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewIdentityService interface {
	mock.TestingT
	Cleanup(func())
}

// NewIdentityService creates a new instance of IdentityService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewIdentityService(t mockConstructorTestingTNewIdentityService) *IdentityService {
	mock := &IdentityService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
