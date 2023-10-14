package entity

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type TimesheetLessonHours struct {
	TimesheetID pgtype.Text
	LessonID    pgtype.Text
	FlagOn      pgtype.Bool
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
}

func (t *TimesheetLessonHours) FieldMap() ([]string, []interface{}) {
	return []string{
			"timesheet_id", "lesson_id", "flag_on", "created_at", "updated_at", "deleted_at",
		}, []interface{}{
			&t.TimesheetID, &t.LessonID, &t.FlagOn, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
		}
}

func (t *TimesheetLessonHours) PreInsert() error {
	now := time.Now()
	return multierr.Combine(
		t.CreatedAt.Set(now),
		t.UpdatedAt.Set(now),
		t.DeletedAt.Set(nil),
	)
}

func (t *TimesheetLessonHours) PreUpdate() error {
	return t.UpdatedAt.Set(time.Now())
}

func (t *TimesheetLessonHours) TableName() string {
	return "timesheet_lesson_hours"
}

func (t *TimesheetLessonHours) PrimaryKey() string {
	return "timesheet_id, lesson_id"
}

func (t *TimesheetLessonHours) UpdateOnConflictQuery() string {
	return `
	flag_on = EXCLUDED.flag_on,
	updated_at = EXCLUDED.updated_at,
	deleted_at = EXCLUDED.deleted_at`
}

type ListTimesheetLessonHours []*TimesheetLessonHours

func (u *ListTimesheetLessonHours) Add() database.Entity {
	e := &TimesheetLessonHours{}
	*u = append(*u, e)

	return e
}
