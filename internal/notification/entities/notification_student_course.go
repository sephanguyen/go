package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type NotificationStudentCourse struct {
	StudentCourseID pgtype.Text
	CourseID        pgtype.Text
	StudentID       pgtype.Text
	LocationID      pgtype.Text
	StartAt         pgtype.Timestamptz
	EndAt           pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	DeletedAt       pgtype.Timestamptz
}

type NotificationStudentCourses []*NotificationStudentCourse

func (e *NotificationStudentCourse) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"student_course_id",
		"course_id",
		"student_id",
		"location_id",
		"start_at",
		"end_at",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.StudentCourseID,
		&e.CourseID,
		&e.StudentID,
		&e.LocationID,
		&e.StartAt,
		&e.EndAt,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

func (e *NotificationStudentCourse) TableName() string { return "notification_student_courses" }

func (ss *NotificationStudentCourses) Add() database.Entity {
	e := &NotificationStudentCourse{}
	*ss = append(*ss, e)

	return e
}
