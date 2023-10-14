package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"
)

type StudentLatestSubmission struct {
	StudentSubmission
}

// FieldMap return "student_submissions" columns
func (s *StudentLatestSubmission) FieldMap() (fields []string, values []interface{}) {
	return s.StudentSubmission.FieldMap()
}

// TableName returns "student_submissions"
func (s *StudentLatestSubmission) TableName() string {
	return "student_latest_submissions"
}

// StudentSubmissions to use with db helper
type StudentLatestSubmissions []*StudentLatestSubmission

// Add appends new StudentSubmission to StudentSubmissions slide and returns that entity
func (ss *StudentLatestSubmissions) Add() database.Entity {
	e := &StudentLatestSubmission{}
	*ss = append(*ss, e)

	return e
}
