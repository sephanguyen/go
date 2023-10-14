package dto

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"

	"github.com/jackc/pgtype"
)

type PartnerAutoCreateTimesheetFlag struct {
	FlagOn bool
}

func (a *PartnerAutoCreateTimesheetFlag) ToEntity() *entity.PartnerAutoCreateTimesheetFlag {
	autoCreateE := &entity.PartnerAutoCreateTimesheetFlag{
		FlagOn:    database.Bool(a.FlagOn),
		CreatedAt: pgtype.Timestamptz{Status: pgtype.Null},
		UpdatedAt: pgtype.Timestamptz{Status: pgtype.Null},
		DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
	}

	return autoCreateE
}
