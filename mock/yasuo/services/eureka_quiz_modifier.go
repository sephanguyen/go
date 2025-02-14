// Code generated by mockery. DO NOT EDIT.

// This file can be generated by running: make gen-mock-repo

package mock_services

import (
	context "context"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"
)

// EurekaQuizModifier is an autogenerated mock type for the EurekaQuizModifier type
type EurekaQuizModifier struct {
	mock.Mock
}

// DeleteQuiz provides a mock function with given fields: ctx, in, opts
func (_m *EurekaQuizModifier) DeleteQuiz(ctx context.Context, in *epb.DeleteQuizRequest, opts ...grpc.CallOption) (*epb.DeleteQuizResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *epb.DeleteQuizResponse
	if rf, ok := ret.Get(0).(func(context.Context, *epb.DeleteQuizRequest, ...grpc.CallOption) *epb.DeleteQuizResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*epb.DeleteQuizResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *epb.DeleteQuizRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveQuizFromLO provides a mock function with given fields: ctx, in, opts
func (_m *EurekaQuizModifier) RemoveQuizFromLO(ctx context.Context, in *epb.RemoveQuizFromLORequest, opts ...grpc.CallOption) (*epb.RemoveQuizFromLOResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *epb.RemoveQuizFromLOResponse
	if rf, ok := ret.Get(0).(func(context.Context, *epb.RemoveQuizFromLORequest, ...grpc.CallOption) *epb.RemoveQuizFromLOResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*epb.RemoveQuizFromLOResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *epb.RemoveQuizFromLORequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateDisplayOrderOfQuizSet provides a mock function with given fields: ctx, in, opts
func (_m *EurekaQuizModifier) UpdateDisplayOrderOfQuizSet(ctx context.Context, in *epb.UpdateDisplayOrderOfQuizSetRequest, opts ...grpc.CallOption) (*epb.UpdateDisplayOrderOfQuizSetResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *epb.UpdateDisplayOrderOfQuizSetResponse
	if rf, ok := ret.Get(0).(func(context.Context, *epb.UpdateDisplayOrderOfQuizSetRequest, ...grpc.CallOption) *epb.UpdateDisplayOrderOfQuizSetResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*epb.UpdateDisplayOrderOfQuizSetResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *epb.UpdateDisplayOrderOfQuizSetRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
