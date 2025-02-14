// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
)

// ChatReaderService is an autogenerated mock type for the ChatReaderService type
type ChatReaderService struct {
	mock.Mock
}

// ListConversationByUsers provides a mock function with given fields: ctx, req
func (_m *ChatReaderService) ListConversationByUsers(ctx context.Context, req *tpb.ListConversationByUsersRequest) (*tpb.ListConversationByUsersResponse, error) {
	ret := _m.Called(ctx, req)

	var r0 *tpb.ListConversationByUsersResponse
	if rf, ok := ret.Get(0).(func(context.Context, *tpb.ListConversationByUsersRequest) *tpb.ListConversationByUsersResponse); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tpb.ListConversationByUsersResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *tpb.ListConversationByUsersRequest) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
