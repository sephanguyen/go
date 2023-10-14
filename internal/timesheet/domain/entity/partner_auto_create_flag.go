package entity

import (
	"github.com/jackc/pgtype"
)

type PartnerAutoCreateTimesheetFlag struct {
	FlagOn    pgtype.Bool
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (a *PartnerAutoCreateTimesheetFlag) FieldMap() (fields []string, values []interface{}) {
	return []string{
			"flag_on",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&a.FlagOn,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.DeletedAt,
		}
}

func (*PartnerAutoCreateTimesheetFlag) TableName() string {
	return "partner_auto_create_timesheet_flag"
}
