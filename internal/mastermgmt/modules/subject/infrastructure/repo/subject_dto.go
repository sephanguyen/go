package repo

import (
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/domain"

	"github.com/jackc/pgtype"
)

type Subject struct {
	SubjectID pgtype.Text
	Name      pgtype.Text

	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (s *Subject) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"subject_id", "name", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&s.SubjectID, &s.Name, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt}
	return
}

func (*Subject) TableName() string {
	return "subject"
}

func (s *Subject) ToEntity() *domain.Subject {
	subject := &domain.Subject{
		SubjectID: s.SubjectID.String,
		Name:      s.Name.String,
		UpdatedAt: s.UpdatedAt.Time,
		CreatedAt: s.CreatedAt.Time,
	}
	if s.DeletedAt.Status == pgtype.Present {
		subject.DeletedAt = &s.DeletedAt.Time
	}
	return subject
}
