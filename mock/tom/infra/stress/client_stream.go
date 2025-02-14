// Code generated by mockery. DO NOT EDIT.

// This file can be generated by running: make gen-mock-repo

package mock_stress

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	metadata "google.golang.org/grpc/metadata"

	pb "github.com/manabie-com/backend/pkg/genproto/tom"
)

// ClientStream is an autogenerated mock type for the ClientStream type
type ClientStream struct {
	mock.Mock
}

// CloseSend provides a mock function with given fields:
func (_m *ClientStream) CloseSend() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Context provides a mock function with given fields:
func (_m *ClientStream) Context() context.Context {
	ret := _m.Called()

	var r0 context.Context
	if rf, ok := ret.Get(0).(func() context.Context); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	return r0
}

// Header provides a mock function with given fields:
func (_m *ClientStream) Header() (metadata.MD, error) {
	ret := _m.Called()

	var r0 metadata.MD
	if rf, ok := ret.Get(0).(func() metadata.MD); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(metadata.MD)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Recv provides a mock function with given fields:
func (_m *ClientStream) Recv() (*pb.SubscribeV2Response, error) {
	ret := _m.Called()

	var r0 *pb.SubscribeV2Response
	if rf, ok := ret.Get(0).(func() *pb.SubscribeV2Response); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*pb.SubscribeV2Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RecvMsg provides a mock function with given fields: m
func (_m *ClientStream) RecvMsg(m interface{}) error {
	ret := _m.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SendMsg provides a mock function with given fields: m
func (_m *ClientStream) SendMsg(m interface{}) error {
	ret := _m.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Trailer provides a mock function with given fields:
func (_m *ClientStream) Trailer() metadata.MD {
	ret := _m.Called()

	var r0 metadata.MD
	if rf, ok := ret.Get(0).(func() metadata.MD); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(metadata.MD)
		}
	}

	return r0
}

type mockConstructorTestingTNewClientStream interface {
	mock.TestingT
	Cleanup(func())
}

// NewClientStream creates a new instance of ClientStream. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewClientStream(t mockConstructorTestingTNewClientStream) *ClientStream {
	mock := &ClientStream{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
