// Code generated by mockery. DO NOT EDIT.

// This file can be generated by running: make gen-mock-repo

package mock_firebase

import mock "github.com/stretchr/testify/mock"

// AuthUtils is an autogenerated mock type for the AuthUtils type
type AuthUtils struct {
	mock.Mock
}

// IsUserNotFound provides a mock function with given fields: err
func (_m *AuthUtils) IsUserNotFound(err error) bool {
	ret := _m.Called(err)

	var r0 bool
	if rf, ok := ret.Get(0).(func(error) bool); ok {
		r0 = rf(err)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

type mockConstructorTestingTNewAuthUtils interface {
	mock.TestingT
	Cleanup(func())
}

// NewAuthUtils creates a new instance of AuthUtils. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAuthUtils(t mockConstructorTestingTNewAuthUtils) *AuthUtils {
	mock := &AuthUtils{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
