package entity

import (
	"time"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type TimesheetActionLog struct {
	ID          pgtype.Text
	TimesheetID pgtype.Text
	UserID      pgtype.Text
	IsSystem    pgtype.Bool
	Action      pgtype.Text
	ExecutedAt  pgtype.Timestamptz
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
}

func (t *TimesheetActionLog) TableName() string {
	return "timesheet_action_log"
}

func (t *TimesheetActionLog) FieldMap() ([]string, []interface{}) {
	return []string{
			"action_log_id",
			"timesheet_id",
			"user_id",
			"action",
			"is_system",
			"executed_at",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&t.ID,
			&t.TimesheetID,
			&t.UserID,
			&t.Action,
			&t.IsSystem,
			&t.ExecutedAt,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.DeletedAt,
		}
}

func (t *TimesheetActionLog) PreInsert() error {
	now := time.Now()
	return multierr.Combine(
		t.CreatedAt.Set(now),
		t.UpdatedAt.Set(now),
		t.DeletedAt.Set(nil),
	)
}
