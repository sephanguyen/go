package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type CourseStudentsAccessPath struct {
	BaseEntity

	CourseStudentID pgtype.Text
	LocationID      pgtype.Text
	CourseID        pgtype.Text
	StudentID       pgtype.Text
	AccessPath      pgtype.Text
}

func (rcv *CourseStudentsAccessPath) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"course_student_id",
		"location_id",
		"course_id",
		"student_id",
		"access_path",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&rcv.CourseStudentID,
		&rcv.LocationID,
		&rcv.CourseID,
		&rcv.StudentID,
		&rcv.AccessPath,
		&rcv.CreatedAt,
		&rcv.UpdatedAt,
		&rcv.DeletedAt,
	}
	return
}

func (rcv *CourseStudentsAccessPath) TableName() string {
	return "course_students_access_paths"
}

type CourseStudentsAccessPaths []*CourseStudentsAccessPath

func (c *CourseStudentsAccessPaths) Add() database.Entity {
	e := &CourseStudentsAccessPath{}
	*c = append(*c, e)

	return e
}
