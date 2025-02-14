// Code generated by mockery v2.23.1. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// SetFunc is an autogenerated mock type for the SetFunc type
type SetFunc struct {
	mock.Mock
}

// Execute provides a mock function with given fields: _a0
func (_m *SetFunc) Execute(_a0 interface{}) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewSetFunc interface {
	mock.TestingT
	Cleanup(func())
}

// NewSetFunc creates a new instance of SetFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSetFunc(t mockConstructorTestingTNewSetFunc) *SetFunc {
	mock := &SetFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
