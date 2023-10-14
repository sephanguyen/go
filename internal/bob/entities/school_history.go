package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type SchoolHistory struct {
	StudentID pgtype.Text
	SchoolID  pgtype.Text
	IsCurrent pgtype.Bool

	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func (s *SchoolHistory) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id", "school_id", "is_current", "created_at", "updated_at",
		}, []interface{}{
			&s.StudentID, &s.SchoolID, &s.IsCurrent, &s.CreatedAt, &s.UpdatedAt,
		}
}

// TableName returns "school_info"
func (s *SchoolHistory) TableName() string {
	return "school_history"
}

type SchoolHistories []*SchoolHistory

func (ss *SchoolHistories) Add() database.Entity {
	e := &SchoolHistory{}
	*ss = append(*ss, e)

	return e
}
