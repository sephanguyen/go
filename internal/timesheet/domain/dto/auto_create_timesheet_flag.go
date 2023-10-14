package dto

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
)

type AutoCreateTimesheetFlag struct {
	StaffID string
	FlagOn  bool
}

func (a *AutoCreateTimesheetFlag) ToEntity() *entity.AutoCreateTimesheetFlag {
	autoCreateE := &entity.AutoCreateTimesheetFlag{
		StaffID:   database.Text(a.StaffID),
		FlagOn:    database.Bool(a.FlagOn),
		CreatedAt: pgtype.Timestamptz{Status: pgtype.Null},
		UpdatedAt: pgtype.Timestamptz{Status: pgtype.Null},
		DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
	}

	return autoCreateE
}

func (a *AutoCreateTimesheetFlag) ValidateUpsertInfo() error {
	if a.StaffID == "" {
		return fmt.Errorf("staff id must not be empty")
	}

	return nil
}

func NewAutoCreateTimeSheetFlagFromRPCUpdateRequest(req *pb.UpdateAutoCreateTimesheetFlagRequest) *AutoCreateTimesheetFlag {
	autoFlag := &AutoCreateTimesheetFlag{
		StaffID: req.GetStaffId(),
		FlagOn:  req.GetFlagOn(),
	}

	return autoFlag
}

func NewAutoCreateTimeSheetFlagFromNATSUpdateRequest(req *pb.NatsUpdateAutoCreateTimesheetFlagRequest) *AutoCreateTimesheetFlag {
	autoFlag := &AutoCreateTimesheetFlag{
		StaffID: req.GetStaffId(),
		FlagOn:  req.GetFlagOn(),
	}

	return autoFlag
}
