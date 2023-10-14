package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/calendar/infrastructure/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) userCreatesADateInfoForDateAndLocation(ctx context.Context, date, location string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	dateTypeID := stepState.DateTypeID
	openingTime := "9:00"
	if dateTypeID == "closed" {
		openingTime = ""
	}

	return s.prepareUpsertDateInfoRequest(StepStateToContext(ctx, stepState), date, location, dateTypeID, openingTime, "draft", "")
}

func (s *suite) prepareUpsertDateInfoRequest(ctx context.Context, date, location, dateTypeID, openingTime, status, timezone string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	dateInfoDate, err := time.Parse(TimeLayout, date)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error in parse datetime: %w", err)
	}

	if len(timezone) == 0 {
		timezone = LoadLocalLocation().String()
	}

	req := &cpb.UpsertDateInfoRequest{
		DateInfo: &cpb.DateInfo{
			Date:        timestamppb.New(dateInfoDate),
			LocationId:  location,
			Status:      status,
			DateTypeId:  dateTypeID,
			OpeningTime: openingTime,
			Timezone:    timezone,
		},
	}

	return s.upsertDateInfo(StepStateToContext(ctx, stepState), req)
}

func (s *suite) upsertDateInfo(ctx context.Context, req *cpb.UpsertDateInfoRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Request = req
	ctx = s.signedCtx(StepStateToContext(ctx, stepState))
	stepState.Response, stepState.ResponseErr = cpb.NewDateInfoModifierServiceClient(s.CalendarConn).
		UpsertDateInfo(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) dateInfoIsCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*cpb.UpsertDateInfoRequest)

	repo := repositories.DateInfoRepo{}
	dateInfo, err := repo.GetDateInfoByDateAndLocationID(ctx, s.CalendarDBTrace, req.DateInfo.Date.AsTime(), req.DateInfo.LocationId)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if !dateInfo.Date.Equal(req.DateInfo.Date.AsTime()) {
		return StepStateToContext(ctx, stepState), NotMatchError("date", req.DateInfo.Date.AsTime(), dateInfo.Date)
	}

	if dateInfo.LocationID != req.DateInfo.LocationId {
		return StepStateToContext(ctx, stepState), NotMatchError("location", req.DateInfo.LocationId, dateInfo.LocationID)
	}

	if dateInfo.DateTypeID != req.DateInfo.DateTypeId {
		return StepStateToContext(ctx, stepState), NotMatchError("day_type_id", req.DateInfo.DateTypeId, dateInfo.DateTypeID)
	}

	if dateInfo.Status != req.DateInfo.Status {
		return StepStateToContext(ctx, stepState), NotMatchError("status", req.DateInfo.Status, dateInfo.Status)
	}

	if dateInfo.OpeningTime != req.DateInfo.OpeningTime {
		return StepStateToContext(ctx, stepState), NotMatchError("opening_time", req.DateInfo.OpeningTime, dateInfo.OpeningTime)
	}

	if dateInfo.TimeZone != req.DateInfo.Timezone {
		return StepStateToContext(ctx, stepState), NotMatchError("time_zone", req.DateInfo.Timezone, dateInfo.TimeZone)
	}

	return StepStateToContext(ctx, stepState), nil
}
