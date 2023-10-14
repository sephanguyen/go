package domain

import (
	"fmt"
	"time"
)

type Subject struct {
	SubjectID string
	Name      string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (s *Subject) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"subject_id", "name", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&s.SubjectID, &s.Name, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt}
	return
}

func (*Subject) TableName() string {
	return "subject"
}

func (s *Subject) String() string {
	return fmt.Sprintf("[ID: %s;Name: %s;]\n", s.SubjectID, s.Name)
}
