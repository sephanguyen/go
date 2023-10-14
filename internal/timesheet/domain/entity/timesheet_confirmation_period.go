package entity

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type TimesheetConfirmationPeriod struct {
	ID        pgtype.Text
	StartDate pgtype.Timestamptz
	EndDate   pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (t *TimesheetConfirmationPeriod) TableName() string {
	return "timesheet_confirmation_period"
}

func (t *TimesheetConfirmationPeriod) FieldMap() ([]string, []interface{}) {
	return []string{
			"id",
			"start_date",
			"end_date",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&t.ID,
			&t.StartDate,
			&t.EndDate,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.DeletedAt,
		}
}

func (t *TimesheetConfirmationPeriod) PreInsert() error {
	now := time.Now()
	return multierr.Combine(
		t.CreatedAt.Set(now),
		t.UpdatedAt.Set(now),
		t.DeletedAt.Set(nil),
	)
}

type TimesheetConfirmationPeriods []*TimesheetConfirmationPeriod

func (t *TimesheetConfirmationPeriods) Add() database.Entity {
	e := &TimesheetConfirmationPeriod{}
	*t = append(*t, e)

	return e
}
