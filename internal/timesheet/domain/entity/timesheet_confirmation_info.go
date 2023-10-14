package entity

import (
	"time"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type TimesheetConfirmationInfo struct {
	ID         pgtype.Text
	LocationID pgtype.Text
	PeriodID   pgtype.Text
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

func (t *TimesheetConfirmationInfo) TableName() string {
	return "timesheet_confirmation_info"
}

func (t *TimesheetConfirmationInfo) PreInsert() error {
	now := time.Now()
	return multierr.Combine(
		t.CreatedAt.Set(now),
		t.UpdatedAt.Set(now),
		t.DeletedAt.Set(nil),
	)
}

func (t *TimesheetConfirmationInfo) FieldMap() ([]string, []interface{}) {
	return []string{
			"id",
			"location_id",
			"period_id",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&t.ID,
			&t.LocationID,
			&t.PeriodID,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.DeletedAt,
		}
}
