// Code generated by mockgen. DO NOT EDIT.
package mock_whiteboard

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (r *MockService) CreateConversionTasks(arg1 context.Context, arg2 []string) ([]string, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (r *MockService) CreateRoom(arg1 context.Context, arg2 *whiteboard.CreateRoomRequest) (*whiteboard.CreateRoomResponse, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*whiteboard.CreateRoomResponse), args.Error(1)
}

func (r *MockService) FetchRoomToken(arg1 context.Context, arg2 string) (string, error) {
	args := r.Called(arg1, arg2)
	return args.Get(0).(string), args.Error(1)
}

func (r *MockService) FetchTasksProgress(arg1 context.Context, arg2 []string) ([]*whiteboard.FetchTaskProgressResponse, error) {
	args := r.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*whiteboard.FetchTaskProgressResponse), args.Error(1)
}
