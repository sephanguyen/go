package clients

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SchedulerClient struct {
	schedulerClient cpb.SchedulerModifierServiceClient
}

type SchedulerClientInterface interface {
	CreateScheduler(ctx context.Context, req *cpb.CreateSchedulerRequest) (*cpb.CreateSchedulerResponse, error)
	UpdateScheduler(ctx context.Context, req *cpb.UpdateSchedulerRequest) (*cpb.UpdateSchedulerResponse, error)
	CreateManySchedulers(ctx context.Context, req *cpb.CreateManySchedulersRequest) (*cpb.CreateManySchedulersResponse, error)
}

func InitSchedulerClient(connect *grpc.ClientConn) *SchedulerClient {
	schedulerClient := cpb.NewSchedulerModifierServiceClient(connect)
	return &SchedulerClient{
		schedulerClient: schedulerClient,
	}
}

func CreateReqCreateScheduler(startTime time.Time, endTime time.Time, frequency constants.Frequency) *cpb.CreateSchedulerRequest {
	return &cpb.CreateSchedulerRequest{StartDate: timestamppb.New(startTime), EndDate: timestamppb.New(endTime), Frequency: constants.MapFrequencyToProtoBuf[frequency]}
}

func (c *SchedulerClient) CreateScheduler(ctx context.Context, req *cpb.CreateSchedulerRequest) (*cpb.CreateSchedulerResponse, error) {
	return c.schedulerClient.CreateScheduler(ctx, req)
}

func (c *SchedulerClient) UpdateScheduler(ctx context.Context, req *cpb.UpdateSchedulerRequest) (*cpb.UpdateSchedulerResponse, error) {
	return c.schedulerClient.UpdateScheduler(ctx, req)
}

func (c *SchedulerClient) CreateManySchedulers(ctx context.Context, req *cpb.CreateManySchedulersRequest) (*cpb.CreateManySchedulersResponse, error) {
	return c.schedulerClient.CreateManySchedulers(ctx, req)
}
