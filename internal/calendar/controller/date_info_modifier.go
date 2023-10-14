package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/calendar/application"
	"github.com/manabie-com/backend/internal/calendar/application/command"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DateInfoModifierService struct {
	upsertDateInfoCmd application.UpsertDateInfoPort
}

func NewDateInfoModifierService(db database.QueryExecer, dateInfoRepo infrastructure.DateInfoPort, locationRepo infrastructure.LocationPort) *DateInfoModifierService {
	return &DateInfoModifierService{
		upsertDateInfoCmd: &command.UpsertDateInfoCommand{
			DB:           db,
			DateInfoRepo: dateInfoRepo,
			LocationRepo: locationRepo,
		},
	}
}

func (d *DateInfoModifierService) UpsertDateInfo(ctx context.Context, in *cpb.UpsertDateInfoRequest) (*cpb.UpsertDateInfoResponse, error) {
	request := &command.UpsertDateInfoRequest{
		Date:        in.DateInfo.Date.AsTime(),
		LocationID:  in.DateInfo.LocationId,
		DateTypeID:  in.DateInfo.DateTypeId,
		OpeningTime: in.DateInfo.OpeningTime,
		Status:      in.DateInfo.Status,
		Timezone:    in.DateInfo.Timezone,
	}

	if err := d.upsertDateInfoCmd.UpsertDateInfo(ctx, request); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &cpb.UpsertDateInfoResponse{
		Successful: true,
		Message:    "Date info setting has been successfully upserted",
	}, nil
}

func (d *DateInfoModifierService) DuplicateDateInfo(ctx context.Context, in *cpb.DuplicateDateInfoRequest) (*cpb.DuplicateDateInfoResponse, error) {
	request := &command.DuplicateDateInfoRequest{
		Date:        in.DateInfo.Date.AsTime(),
		LocationID:  in.DateInfo.LocationId,
		DateTypeID:  in.DateInfo.DateTypeId,
		OpeningTime: in.DateInfo.OpeningTime,
		Status:      in.DateInfo.Status,
		Timezone:    in.DateInfo.Timezone,
		StartDate:   in.RepeatInfo.StartDate.AsTime(),
		EndDate:     in.RepeatInfo.EndDate.AsTime(),
		Frequency:   in.RepeatInfo.Condition,
	}

	if err := d.upsertDateInfoCmd.DuplicateDateInfo(ctx, request); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &cpb.DuplicateDateInfoResponse{
		Successful: true,
		Message:    "Date info setting has been successfully duplicated",
	}, nil
}
