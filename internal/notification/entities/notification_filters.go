package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type NotificationCourseFilter struct {
	NotificationID pgtype.Text
	CourseID       pgtype.Text
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

func (e *NotificationCourseFilter) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"notification_id",
		"course_id",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.NotificationID,
		&e.CourseID,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

func (*NotificationCourseFilter) TableName() string {
	return "notification_course_filter"
}

type NotificationCourseFilters []*NotificationCourseFilter

func (ss *NotificationCourseFilters) Add() database.Entity {
	e := &NotificationCourseFilter{}
	*ss = append(*ss, e)
	return e
}

type NotificationLocationFilter struct {
	NotificationID pgtype.Text
	LocationID     pgtype.Text
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

func (e *NotificationLocationFilter) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"notification_id",
		"location_id",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.NotificationID,
		&e.LocationID,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

func (*NotificationLocationFilter) TableName() string {
	return "notification_location_filter"
}

type NotificationLocationFilters []*NotificationLocationFilter

func (ss *NotificationLocationFilters) Add() database.Entity {
	e := &NotificationLocationFilter{}
	*ss = append(*ss, e)
	return e
}

type NotificationClassFilter struct {
	NotificationID pgtype.Text
	ClassID        pgtype.Text
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

func (e *NotificationClassFilter) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"notification_id",
		"class_id",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.NotificationID,
		&e.ClassID,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

func (*NotificationClassFilter) TableName() string {
	return "notification_class_filter"
}

type NotificationClassFilters []*NotificationClassFilter

func (ss *NotificationClassFilters) Add() database.Entity {
	e := &NotificationClassFilter{}
	*ss = append(*ss, e)
	return e
}
