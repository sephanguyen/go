package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type AutoCreateTimesheetFlag struct {
	StaffID   pgtype.Text
	FlagOn    pgtype.Bool
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (a *AutoCreateTimesheetFlag) FieldMap() (fields []string, values []interface{}) {
	return []string{
			"staff_id",
			"flag_on",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&a.StaffID,
			&a.FlagOn,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.DeletedAt,
		}
}

func (*AutoCreateTimesheetFlag) TableName() string {
	return "auto_create_timesheet_flag"
}

type AutoCreateTimesheetFlags []*AutoCreateTimesheetFlag

func (a *AutoCreateTimesheetFlag) PrimaryKey() string {
	return "staff_id"
}

func (a *AutoCreateTimesheetFlag) UpdateOnConflictQuery() string {
	return `
	flag_on = EXCLUDED.flag_on,
	updated_at = EXCLUDED.updated_at,
	deleted_at = EXCLUDED.deleted_at`
}

func (la *AutoCreateTimesheetFlags) Add() database.Entity {
	e := &AutoCreateTimesheetFlag{}
	*la = append(*la, e)

	return e
}
