package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
)

type OtherWorkingHours struct {
	ID                pgtype.Text
	TimesheetID       pgtype.Text
	TimesheetConfigID pgtype.Text
	StartTime         pgtype.Timestamptz
	EndTime           pgtype.Timestamptz
	TotalHour         pgtype.Int2
	Remarks           pgtype.Text
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
}

func (o *OtherWorkingHours) FieldMap() ([]string, []interface{}) {
	return []string{
			"other_working_hours_id",
			"timesheet_id",
			"timesheet_config_id",
			"start_time",
			"end_time",
			"total_hour",
			"remarks",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&o.ID,
			&o.TimesheetID,
			&o.TimesheetConfigID,
			&o.StartTime,
			&o.EndTime,
			&o.TotalHour,
			&o.Remarks,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.DeletedAt,
		}
}

func (o *OtherWorkingHours) UpdateOnConflictQuery() string {
	return `
	timesheet_config_id = EXCLUDED.timesheet_config_id,
	start_time = EXCLUDED.start_time,
	end_time = EXCLUDED.end_time,
	total_hour = EXCLUDED.total_hour,
	remarks = EXCLUDED.remarks,
	updated_at = EXCLUDED.updated_at,
	deleted_at = EXCLUDED.deleted_at`
}

func (*OtherWorkingHours) TableName() string {
	return "other_working_hours"
}

func (*OtherWorkingHours) UpsertConflictField() string {
	return "other_working_hours_id"
}

func NewOtherWorkingHours() *OtherWorkingHours {
	return &OtherWorkingHours{
		ID:        database.Text(idutil.ULIDNow()),
		Remarks:   pgtype.Text{Status: pgtype.Null},
		DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
	}
}

type ListOtherWorkingHours []*OtherWorkingHours

// Add append new OtherWorkingHours
func (l *ListOtherWorkingHours) Add() database.Entity {
	e := &OtherWorkingHours{}
	*l = append(*l, e)

	return e
}
