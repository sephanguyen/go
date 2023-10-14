package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_queries "github.com/manabie-com/backend/mock/calendar/application/queries"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestDateInfoReaderService_FetchDateInfo(t *testing.T) {
	t.Parallel()
	dateInfoQueryHandler := &mock_queries.MockDateInfoQueryHandler{}
	service := &DateInfoReaderService{
		dateInfoQueryHandler: dateInfoQueryHandler,
	}

	startDate := timestamppb.New(time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC))
	endDate := timestamppb.New(time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC))
	locationID := idutil.ULIDNow()
	timezone := "sample-timezone"

	t.Run("success", func(t *testing.T) {
		req := &cpb.FetchDateInfoRequest{
			StartDate:  startDate,
			EndDate:    endDate,
			LocationId: locationID,
			Timezone:   timezone,
		}
		dateInfoQueryHandler.On("FetchDateInfoByDateRangeAndLocationID", mock.Anything, &payloads.FetchDateInfoByDateRangeRequest{
			StartDate:  req.StartDate.AsTime(),
			EndDate:    req.EndDate.AsTime(),
			LocationID: req.LocationId,
			Timezone:   req.Timezone,
		}).Once().Return(&payloads.FetchDateInfoByDateRangeResponse{
			DateInfos: []*dto.DateInfo{
				{
					Date:                time.Now(),
					LocationID:          req.LocationId,
					DateTypeID:          "regular",
					OpeningTime:         "09:00",
					Status:              "draft",
					TimeZone:            "sample-timezone",
					DateTypeDisplayName: "Regular",
				},
			},
		}, nil)
		res, err := service.FetchDateInfo(context.Background(), req)
		require.Nil(t, err)
		require.NotNil(t, res)
		require.Equal(t, true, res.Successful)
	})
	t.Run("empty result", func(t *testing.T) {
		req := &cpb.FetchDateInfoRequest{
			StartDate:  startDate,
			EndDate:    endDate,
			LocationId: locationID,
			Timezone:   timezone,
		}
		dateInfoQueryHandler.On("FetchDateInfoByDateRangeAndLocationID", mock.Anything, &payloads.FetchDateInfoByDateRangeRequest{
			StartDate:  req.StartDate.AsTime(),
			EndDate:    req.EndDate.AsTime(),
			LocationID: req.LocationId,
			Timezone:   req.Timezone,
		}).Once().Return(&payloads.FetchDateInfoByDateRangeResponse{
			DateInfos: []*dto.DateInfo{},
		}, nil)
		res, err := service.FetchDateInfo(context.Background(), req)
		require.Nil(t, err)
		require.NotNil(t, res)
	})
	t.Run("failed", func(t *testing.T) {
		req := &cpb.FetchDateInfoRequest{
			StartDate:  startDate,
			EndDate:    endDate,
			LocationId: locationID,
			Timezone:   timezone,
		}
		dateInfoQueryHandler.On("FetchDateInfoByDateRangeAndLocationID", mock.Anything, &payloads.FetchDateInfoByDateRangeRequest{
			StartDate:  req.StartDate.AsTime(),
			EndDate:    req.EndDate.AsTime(),
			LocationID: req.LocationId,
			Timezone:   req.Timezone,
		}).Once().Return(nil, errors.New("error"))
		res, err := service.FetchDateInfo(context.Background(), req)
		require.NotNil(t, err)
		require.Nil(t, res)
	})
}
