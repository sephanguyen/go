// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"

	spb "github.com/manabie-com/backend/pkg/manabuf/shamir/v1"
)

// ShamirClient is an autogenerated mock type for the ShamirClient type
type ShamirClient struct {
	mock.Mock
}

// GenerateFakeToken provides a mock function with given fields: ctx, in, opts
func (_m *ShamirClient) GenerateFakeToken(ctx context.Context, in *spb.GenerateFakeTokenRequest, opts ...grpc.CallOption) (*spb.GenerateFakeTokenResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *spb.GenerateFakeTokenResponse
	if rf, ok := ret.Get(0).(func(context.Context, *spb.GenerateFakeTokenRequest, ...grpc.CallOption) *spb.GenerateFakeTokenResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*spb.GenerateFakeTokenResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *spb.GenerateFakeTokenRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewShamirClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewShamirClient creates a new instance of ShamirClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewShamirClient(t mockConstructorTestingTNewShamirClient) *ShamirClient {
	mock := &ShamirClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
