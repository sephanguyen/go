package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type NotificationClassMember struct {
	StudentID  pgtype.Text
	ClassID    pgtype.Text
	StartAt    pgtype.Timestamptz
	EndAt      pgtype.Timestamptz
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
	LocationID pgtype.Text
	CourseID   pgtype.Text
	DeletedAt  pgtype.Timestamptz
}

func (e *NotificationClassMember) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"student_id",
		"class_id",
		"start_at",
		"end_at",
		"created_at",
		"updated_at",
		"location_id",
		"course_id",
		"deleted_at",
	}
	values = []interface{}{
		&e.StudentID,
		&e.ClassID,
		&e.StartAt,
		&e.EndAt,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.LocationID,
		&e.CourseID,
		&e.DeletedAt,
	}
	return
}

func (*NotificationClassMember) TableName() string {
	return "notification_class_members"
}

type NotificationClassMembers []*NotificationClassMember

func (es *NotificationClassMembers) Add() database.Entity {
	e := &NotificationClassMember{}
	*es = append(*es, e)

	return e
}
