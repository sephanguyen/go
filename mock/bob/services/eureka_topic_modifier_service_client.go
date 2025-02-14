// Code generated by mockery. DO NOT EDIT.

// This file can be generated by running: make gen-mock-repo

package mock_services

import (
	context "context"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"
)

// EurekaTopicModifierServiceClient is an autogenerated mock type for the EurekaTopicModifierServiceClient type
type EurekaTopicModifierServiceClient struct {
	mock.Mock
}

// AssignTopicItems provides a mock function with given fields: ctx, in, opts
func (_m *EurekaTopicModifierServiceClient) AssignTopicItems(ctx context.Context, in *epb.AssignTopicItemsRequest, opts ...grpc.CallOption) (*epb.AssignTopicItemsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *epb.AssignTopicItemsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *epb.AssignTopicItemsRequest, ...grpc.CallOption) *epb.AssignTopicItemsResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*epb.AssignTopicItemsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *epb.AssignTopicItemsRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
