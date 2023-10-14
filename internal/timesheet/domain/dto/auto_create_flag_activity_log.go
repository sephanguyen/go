package dto

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
)

type AutoCreateFlagActivityLog struct {
	ID         string
	StaffID    string
	ChangeTime time.Time
	FlagOn     bool
}

func (a *AutoCreateFlagActivityLog) ToEntity() *entity.AutoCreateFlagActivityLog {
	logE := &entity.AutoCreateFlagActivityLog{
		ID:         database.Text(a.ID),
		StaffID:    database.Text(a.StaffID),
		ChangeTime: database.Timestamptz(a.ChangeTime),
		FlagOn:     database.Bool(a.FlagOn),
		CreatedAt:  pgtype.Timestamptz{Status: pgtype.Null},
		UpdatedAt:  pgtype.Timestamptz{Status: pgtype.Null},
		DeletedAt:  pgtype.Timestamptz{Status: pgtype.Null},
	}

	return logE
}

func NewAutoCreateTimeSheetFlagLogFromRPCUpdateRequest(req *pb.UpdateAutoCreateTimesheetFlagRequest) *AutoCreateFlagActivityLog {
	autoFlag := &AutoCreateFlagActivityLog{
		StaffID: req.GetStaffId(),
		FlagOn:  req.GetFlagOn(),
	}

	return autoFlag
}

func NewAutoCreateFlagActivityLogFromEntity(autoCFALEntity *entity.AutoCreateFlagActivityLog) *AutoCreateFlagActivityLog {
	return &AutoCreateFlagActivityLog{
		ID:         autoCFALEntity.ID.String,
		StaffID:    autoCFALEntity.StaffID.String,
		ChangeTime: autoCFALEntity.ChangeTime.Time,
		FlagOn:     autoCFALEntity.FlagOn.Bool,
	}
}
