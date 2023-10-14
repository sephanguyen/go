package calendar

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) getDateInfoByDurations(ctx context.Context, sDate, eDate, locationID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	startDate, err1 := time.Parse(TimeLayout, sDate)
	endDate, err2 := time.Parse(TimeLayout, eDate)

	if err := multierr.Combine(err1, err2); err != nil {
		return ctx, fmt.Errorf("parse datetime error: %w", err)
	}

	req := &cpb.FetchDateInfoRequest{}

	if !startDate.IsZero() && !endDate.IsZero() && locationID != "" {
		req = &cpb.FetchDateInfoRequest{
			StartDate:  timestamppb.New(startDate),
			EndDate:    timestamppb.New(endDate),
			LocationId: locationID,
		}
	}
	req.Timezone = LoadLocalLocation().String()

	ctx = s.signedCtx(StepStateToContext(ctx, stepState))
	stepState.Response, stepState.ResponseErr = cpb.NewDateInfoReaderServiceClient(s.CalendarConn).FetchDateInfo(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsDateInfoByDurations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*cpb.FetchDateInfoRequest)
	res := stepState.Response.(*cpb.FetchDateInfoResponse)

	if len(res.DateInfos) == 0 {
		return StepStateToContext(ctx, stepState), errors.New("expected a list of date info")
	}

	startDate := req.StartDate.AsTime()
	endDate := req.EndDate.AsTime()
	failedDateInfos := make([]*cpb.DateInfoDetailed, 0, len(res.DateInfos))

	for _, dateInfoDetailed := range res.DateInfos {
		dateInfoDate := dateInfoDetailed.DateInfo.Date.AsTime()

		if dateInfoDate.Before(startDate) && dateInfoDate.After(endDate) || dateInfoDetailed.DateInfo.LocationId != req.LocationId {
			failedDateInfos = append(failedDateInfos, &cpb.DateInfoDetailed{
				DateInfo: &cpb.DateInfo{
					Date:       dateInfoDetailed.DateInfo.Date,
					LocationId: dateInfoDetailed.DateInfo.LocationId,
				},
			})
		}
	}

	if len(failedDateInfos) > 0 {
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("retrieved dateInfos are not within the criteria of %v to %v under location %s: %v", startDate, endDate, req.LocationId, failedDateInfos)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) existingDateInfos(ctx context.Context, date, locationID, dateType, openTime, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	timezone := LoadLocalLocation().String()

	if len(dateType) < 1 {
		return StepStateToContext(ctx, stepState), nil
	}

	convertedDate, err := time.Parse(TimeLayout, date)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("parse datetime error: %w", err)
	}

	// insert date type
	dateTypeQuery := "INSERT INTO day_type (day_type_id, display_name) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	if _, err := s.CalendarDBTrace.Exec(ctx, dateTypeQuery, dateType, strings.ToUpper(dateType)); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not init date type for this resource path: %w", err)
	}

	stepState.DateTypes = append(stepState.DateTypes, &dto.DateType{
		DateTypeID: dateType,
	})

	// insert date info
	dateInfoQuery := `INSERT INTO day_info (date, location_id, day_type_id, opening_time, status, time_zone)
					VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING`
	if _, err := s.CalendarDBTrace.Exec(ctx, dateInfoQuery, convertedDate, locationID, dateType, openTime, status, timezone); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not init date info: %w", err)
	}

	stepState.DateInfos = append(stepState.DateInfos, &dto.DateInfo{
		Date:        convertedDate,
		LocationID:  locationID,
		DateTypeID:  dateType,
		OpeningTime: openTime,
		Status:      status,
		TimeZone:    timezone,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) existingDateTypes(ctx context.Context, dateType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	dateTypeQuery := "INSERT INTO day_type (day_type_id) VALUES ($1) ON CONFLICT DO NOTHING"
	if _, err := s.CalendarDBTrace.Exec(ctx, dateTypeQuery, dateType); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not init date type for this resource path: %w", err)
	}

	stepState.DateTypes = append(stepState.DateTypes, &dto.DateType{
		DateTypeID: dateType,
	})
	return StepStateToContext(ctx, stepState), nil
}
