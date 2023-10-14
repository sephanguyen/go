package repo

import (
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
)

type StudentParent struct {
	StudentID    pgtype.Text
	ParentID     pgtype.Text
	Relationship pgtype.Text
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
}

func (sp *StudentParent) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"parent_id",
			"relationship",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&sp.StudentID,
			&sp.ParentID,
			&sp.Relationship,
			&sp.CreatedAt,
			&sp.UpdatedAt,
			&sp.DeletedAt,
		}
}

func (sp *StudentParent) TableName() string {
	return "student_parents"
}

func (sp *StudentParent) ToStudentParentDomain() domain.StudentParent {
	domain := domain.StudentParent{
		StudentID:    sp.StudentID.String,
		ParentID:     sp.ParentID.String,
		Relationship: sp.Relationship.String,
		CreatedAt:    sp.CreatedAt.Time,
		UpdatedAt:    sp.UpdatedAt.Time,
	}

	if sp.DeletedAt.Status == pgtype.Present {
		domain.DeletedAt = &sp.DeletedAt.Time
	}

	return domain
}
