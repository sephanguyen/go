package controller

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/calendar/application/command"
	mock_command "github.com/manabie-com/backend/mock/calendar/application/command"
	"github.com/manabie-com/backend/mock/testutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestScheduler_CreateScheduler(t *testing.T) {
	t.Parallel()
	createSchedulerCmd := &mock_command.MockCreateSchedulerCommand{}
	mockDB := testutil.NewMockDB()
	service := &SchedulerModifierService{
		db:                 mockDB.DB,
		createSchedulerCmd: createSchedulerCmd,
	}
	t.Run("success", func(t *testing.T) {
		req := &cpb.CreateSchedulerRequest{
			StartDate: timestamppb.New(time.Date(2022, 9, 21, 0, 0, 0, 0, time.UTC)),
			EndDate:   timestamppb.New(time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC)),
			Frequency: cpb.Frequency_WEEKLY,
		}
		createSchedulerCmd.On("CreateScheduler", mock.Anything, mockDB.DB, &command.CreateSchedulerRequest{
			StartDate: req.StartDate.AsTime(),
			EndDate:   req.EndDate.AsTime(),
			Frequency: req.Frequency.String(),
		}).Once().Return(&command.CreateSchedulerResponse{
			SchedulerID: "scheduler-id",
		}, nil)
		res, err := service.CreateScheduler(context.Background(), req)
		require.Nil(t, err)
		require.NotNil(t, res)
		require.Equal(t, "scheduler-id", res.SchedulerId)
	})
	t.Run("failed", func(t *testing.T) {
		req := &cpb.CreateSchedulerRequest{
			StartDate: timestamppb.New(time.Date(2022, 11, 21, 0, 0, 0, 0, time.UTC)),
			EndDate:   timestamppb.New(time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC)),
			Frequency: cpb.Frequency_WEEKLY,
		}
		createSchedulerCmd.On("CreateScheduler", mock.Anything, mockDB.DB, &command.CreateSchedulerRequest{
			StartDate: req.StartDate.AsTime(),
			EndDate:   req.EndDate.AsTime(),
			Frequency: req.Frequency.String(),
		}).Once().Return(nil, errors.New("start date cannot greater than end date"))
		res, err := service.CreateScheduler(context.Background(), req)
		require.NotNil(t, err)
		require.Nil(t, res)
	})
}

func TestScheduler_UpdateScheduler(t *testing.T) {
	t.Parallel()
	updateSchedulerCmd := &mock_command.MockUpdateSchedulerCommand{}
	mockDB := testutil.NewMockDB()
	service := &SchedulerModifierService{
		db:                 mockDB.DB,
		updateSchedulerCmd: updateSchedulerCmd,
	}
	t.Run("success", func(t *testing.T) {
		req := &cpb.UpdateSchedulerRequest{
			SchedulerId: "scheduler-id",
			EndDate:     timestamppb.New(time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC)),
		}
		updateSchedulerCmd.On("UpdateScheduler", mock.Anything, mockDB.DB, &command.UpdateSchedulerRequest{
			SchedulerID: "scheduler-id",
			EndDate:     req.EndDate.AsTime(),
		}).Once().Return(nil)
		res, err := service.UpdateScheduler(context.Background(), req)
		require.Nil(t, err)
		require.NotNil(t, res)
	})
	t.Run("failed", func(t *testing.T) {
		req := &cpb.UpdateSchedulerRequest{
			SchedulerId: "scheduler-id",
			EndDate:     timestamppb.New(time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC)),
		}
		updateSchedulerCmd.On("UpdateScheduler", mock.Anything, mockDB.DB, &command.UpdateSchedulerRequest{
			SchedulerID: "scheduler-id",
			EndDate:     req.EndDate.AsTime(),
		}).Once().Return(errors.New("something went wrong"))
		res, err := service.UpdateScheduler(context.Background(), req)
		require.NotNil(t, err)
		require.Nil(t, res)
	})
}

func TestScheduler_CreateManySchedulers(t *testing.T) {
	t.Parallel()
	createSchedulerCmd := &mock_command.MockCreateSchedulerCommand{}
	mockDB := testutil.NewMockDB()
	service := &SchedulerModifierService{
		db:                 mockDB.DB,
		createSchedulerCmd: createSchedulerCmd,
	}

	t.Run("success", func(t *testing.T) {
		req := &cpb.CreateManySchedulersRequest{
			Schedulers: []*cpb.CreateSchedulerWithIdentityRequest{
				{
					Identity: "lesson_id_01",
					Request: &cpb.CreateSchedulerRequest{
						StartDate: timestamppb.New(time.Date(2022, 9, 21, 0, 0, 0, 0, time.UTC)),
						EndDate:   timestamppb.New(time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC)),
						Frequency: cpb.Frequency_WEEKLY,
					},
				},
			},
		}
		mockResp := &cpb.CreateManySchedulersResponse{
			MapSchedulers: map[string]string{
				"lesson_id_01": "scheduler_id_01",
				"lesson_id_02": "scheduler_id_02",
			},
		}
		createSchedulerCmd.On("CreateManySchedulers", mock.Anything, mockDB.DB, req).Once().Return(mockResp, nil)
		resp, err := service.CreateManySchedulers(context.Background(), req)
		require.Nil(t, err)
		require.NotNil(t, resp.MapSchedulers)
		require.Equal(t, mockResp.MapSchedulers, resp.MapSchedulers)
	})
	t.Run("failed", func(t *testing.T) {
		req := &cpb.CreateManySchedulersRequest{
			Schedulers: []*cpb.CreateSchedulerWithIdentityRequest{
				{
					Identity: "lesson_id_01",
					Request: &cpb.CreateSchedulerRequest{
						StartDate: timestamppb.New(time.Date(2022, 9, 21, 0, 0, 0, 0, time.UTC)),
						EndDate:   timestamppb.New(time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC)),
						Frequency: cpb.Frequency_WEEKLY,
					},
				},
			},
		}
		createSchedulerCmd.On("CreateManySchedulers", mock.Anything, mockDB.DB, req).Once().Return(nil, fmt.Errorf("error"))
		resp, err := service.CreateManySchedulers(context.Background(), req)
		require.NotNil(t, err)
		require.Nil(t, resp)
	})
}
