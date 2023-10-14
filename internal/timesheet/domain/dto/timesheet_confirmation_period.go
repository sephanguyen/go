package dto

import (
	"errors"
	"time"

	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type TimesheetConfirmationPeriods []*TimesheetConfirmationPeriod

type TimesheetConfirmationPeriod struct {
	ID        string
	StartDate time.Time
	EndDate   time.Time
}

func NewTimesheetConfirmationPeriodToRPCResponse(timesheetConfirmationPeriod *TimesheetConfirmationPeriod) *tpb.GetTimesheetConfirmationPeriodByDateResponse {
	timesheetConfirmationPeriodConvert := tpb.TimesheetConfirmationPeriod{
		Id:        timesheetConfirmationPeriod.ID,
		StartDate: timestamppb.New(timesheetConfirmationPeriod.StartDate),
		EndDate:   timestamppb.New(timesheetConfirmationPeriod.EndDate),
	}

	return &tpb.GetTimesheetConfirmationPeriodByDateResponse{
		TimesheetConfirmationPeriod: &timesheetConfirmationPeriodConvert,
	}
}

func NewTimesheetConfirmationPeriodFromEntity(timesheetConfirmationPeriod *entity.TimesheetConfirmationPeriod) *TimesheetConfirmationPeriod {
	return &TimesheetConfirmationPeriod{
		ID:        timesheetConfirmationPeriod.ID.String,
		StartDate: timesheetConfirmationPeriod.StartDate.Time,
		EndDate:   timesheetConfirmationPeriod.EndDate.Time,
	}
}

func ValidateGetPeriodInfo(req *tpb.GetTimesheetConfirmationPeriodByDateRequest) error {
	if req.GetDate() == nil {
		return errors.New("date must not be empty")
	}

	if req.GetDate().AsTime().Before(constant.KTimesheetMinDate) {
		return errors.New("date must be larger than 2022-01-01")
	}

	return nil
}
