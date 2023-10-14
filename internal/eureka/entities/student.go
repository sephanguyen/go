package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Student struct {
	StudentID pgtype.Text
}

// FieldMap return "student_submissions" columns
func (s *Student) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"student_id",
	}
	values = []interface{}{
		&s.StudentID,
	}
	return
}

// TableName returns "student_submissions"
func (s *Student) TableName() string {
	return "students"
}

// Students to use with db helper
type Students []*Student

// Add appends new Student to Students slide and returns that entity
func (ss *Students) Add() database.Entity {
	e := &Student{}
	*ss = append(*ss, e)

	return e
}
