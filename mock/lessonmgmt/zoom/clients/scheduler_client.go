package mock_clients

import (
	"context"

	mpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"github.com/stretchr/testify/mock"
)

type MockSchedulerClient struct {
	mock.Mock
}

func (m *MockSchedulerClient) CreateScheduler(arg1 context.Context, arg2 *mpb.CreateSchedulerRequest) (*mpb.CreateSchedulerResponse, error) {
	args := m.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mpb.CreateSchedulerResponse), args.Error(1)
}

func (m *MockSchedulerClient) UpdateScheduler(arg1 context.Context, arg2 *mpb.UpdateSchedulerRequest) (*mpb.UpdateSchedulerResponse, error) {
	args := m.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mpb.UpdateSchedulerResponse), args.Error(1)
}

func (m *MockSchedulerClient) CreateManySchedulers(arg1 context.Context, arg2 *mpb.CreateManySchedulersRequest) (*mpb.CreateManySchedulersResponse, error) {
	args := m.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mpb.CreateManySchedulersResponse), args.Error(1)
}
