package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/calendar/application/command"
	mock_command "github.com/manabie-com/backend/mock/calendar/application/command"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestScheduler_UpsertDateInfo(t *testing.T) {
	t.Parallel()
	upsertDateInfoCmd := &mock_command.MockUpsertDateInfoCommand{}
	service := &DateInfoModifierService{
		upsertDateInfoCmd: upsertDateInfoCmd,
	}

	t.Run("success", func(t *testing.T) {
		req := &cpb.UpsertDateInfoRequest{
			DateInfo: &cpb.DateInfo{
				Date:        timestamppb.New(time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC)),
				LocationId:  "location-id-1",
				DateTypeId:  "regular",
				OpeningTime: "09:00",
				Status:      "draft",
				Timezone:    "sample-timezone",
			},
		}
		upsertDateInfoCmd.On("UpsertDateInfo", mock.Anything, &command.UpsertDateInfoRequest{
			Date:        req.DateInfo.Date.AsTime(),
			LocationID:  req.DateInfo.LocationId,
			DateTypeID:  req.DateInfo.DateTypeId,
			OpeningTime: req.DateInfo.OpeningTime,
			Status:      req.DateInfo.Status,
			Timezone:    req.DateInfo.Timezone,
		}).Once().Return(nil)
		res, err := service.UpsertDateInfo(context.Background(), req)
		require.Nil(t, err)
		require.NotNil(t, res)
		require.Equal(t, true, res.Successful)
	})
	t.Run("failed", func(t *testing.T) {
		req := &cpb.UpsertDateInfoRequest{
			DateInfo: &cpb.DateInfo{
				Date:        timestamppb.New(time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC)),
				LocationId:  "location-id-1",
				DateTypeId:  "regular",
				OpeningTime: "09:00",
				Status:      "draft",
				Timezone:    "sample-timezone",
			},
		}
		upsertDateInfoCmd.On("UpsertDateInfo", mock.Anything, &command.UpsertDateInfoRequest{
			Date:        req.DateInfo.Date.AsTime(),
			LocationID:  req.DateInfo.LocationId,
			DateTypeID:  req.DateInfo.DateTypeId,
			OpeningTime: req.DateInfo.OpeningTime,
			Status:      req.DateInfo.Status,
			Timezone:    req.DateInfo.Timezone,
		}).Once().Return(errors.New("error"))
		res, err := service.UpsertDateInfo(context.Background(), req)
		require.NotNil(t, err)
		require.Nil(t, res)
	})
}

func TestScheduler_DuplicateDateInfo(t *testing.T) {
	t.Parallel()
	upsertDateInfoCmd := &mock_command.MockUpsertDateInfoCommand{}
	service := &DateInfoModifierService{
		upsertDateInfoCmd: upsertDateInfoCmd,
	}

	t.Run("success", func(t *testing.T) {
		req := &cpb.DuplicateDateInfoRequest{
			DateInfo: &cpb.DateInfo{
				Date:        timestamppb.New(time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC)),
				LocationId:  "location-id-1",
				DateTypeId:  "regular",
				OpeningTime: "09:00",
				Status:      "draft",
				Timezone:    "sample-timezone",
			},
			RepeatInfo: &cpb.RepeatInfo{
				StartDate: timestamppb.New(time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC)),
				EndDate:   timestamppb.New(time.Date(2022, 10, 02, 0, 0, 0, 0, time.UTC)),
				Condition: "daily",
			},
		}
		upsertDateInfoCmd.On("DuplicateDateInfo", mock.Anything, &command.DuplicateDateInfoRequest{
			Date:        req.DateInfo.Date.AsTime(),
			LocationID:  req.DateInfo.LocationId,
			DateTypeID:  req.DateInfo.DateTypeId,
			OpeningTime: req.DateInfo.OpeningTime,
			Status:      req.DateInfo.Status,
			Timezone:    req.DateInfo.Timezone,
			StartDate:   req.RepeatInfo.StartDate.AsTime(),
			EndDate:     req.RepeatInfo.EndDate.AsTime(),
			Frequency:   req.RepeatInfo.Condition,
		}).Once().Return(nil)
		res, err := service.DuplicateDateInfo(context.Background(), req)
		require.Nil(t, err)
		require.NotNil(t, res)
		require.Equal(t, true, res.Successful)
	})
	t.Run("failed", func(t *testing.T) {
		req := &cpb.DuplicateDateInfoRequest{
			DateInfo: &cpb.DateInfo{
				Date:        timestamppb.New(time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC)),
				LocationId:  "location-id-1",
				DateTypeId:  "regular",
				OpeningTime: "09:00",
				Status:      "draft",
				Timezone:    "sample-timezone",
			},
			RepeatInfo: &cpb.RepeatInfo{
				StartDate: timestamppb.New(time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC)),
				EndDate:   timestamppb.New(time.Date(2022, 10, 02, 0, 0, 0, 0, time.UTC)),
				Condition: "daily",
			},
		}
		upsertDateInfoCmd.On("DuplicateDateInfo", mock.Anything, &command.DuplicateDateInfoRequest{
			Date:        req.DateInfo.Date.AsTime(),
			LocationID:  req.DateInfo.LocationId,
			DateTypeID:  req.DateInfo.DateTypeId,
			OpeningTime: req.DateInfo.OpeningTime,
			Status:      req.DateInfo.Status,
			Timezone:    req.DateInfo.Timezone,
			StartDate:   req.RepeatInfo.StartDate.AsTime(),
			EndDate:     req.RepeatInfo.EndDate.AsTime(),
			Frequency:   req.RepeatInfo.Condition,
		}).Once().Return(errors.New("error"))
		res, err := service.DuplicateDateInfo(context.Background(), req)
		require.NotNil(t, err)
		require.Nil(t, res)
	})
}
