// Code generated by mockery. DO NOT EDIT.

// This file can be generated by running: make gen-mock-repo

package mock_multitenant

import mock "github.com/stretchr/testify/mock"

// TenantInfo is an autogenerated mock type for the TenantInfo type
type TenantInfo struct {
	mock.Mock
}

// GetDisplayName provides a mock function with given fields:
func (_m *TenantInfo) GetDisplayName() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetEmailLinkSignInEnabled provides a mock function with given fields:
func (_m *TenantInfo) GetEmailLinkSignInEnabled() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// GetPasswordSignUpAllowed provides a mock function with given fields:
func (_m *TenantInfo) GetPasswordSignUpAllowed() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

type mockConstructorTestingTNewTenantInfo interface {
	mock.TestingT
	Cleanup(func())
}

// NewTenantInfo creates a new instance of TenantInfo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTenantInfo(t mockConstructorTestingTNewTenantInfo) *TenantInfo {
	mock := &TenantInfo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
