package repo

import (
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
)

type Student struct {
	StudentID         pgtype.Text
	CurrentGrade      pgtype.Int2
	GradeID           pgtype.Text
	StudentExternalID pgtype.Text
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
}

func (s *Student) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"current_grade",
			"grade_id",
			"student_external_id",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&s.StudentID,
			&s.CurrentGrade,
			&s.GradeID,
			&s.StudentExternalID,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.DeletedAt,
		}
}

func (s *Student) TableName() string {
	return "students"
}

func (s *Student) ToStudentDomain() domain.Student {
	domain := domain.Student{
		StudentID:         s.StudentID.String,
		CurrentGrade:      int(s.CurrentGrade.Int),
		GradeID:           s.GradeID.String,
		StudentExternalID: s.StudentExternalID.String,
		CreatedAt:         s.CreatedAt.Time,
		UpdatedAt:         s.UpdatedAt.Time,
	}

	if s.DeletedAt.Status == pgtype.Present {
		domain.DeletedAt = &s.DeletedAt.Time
	}

	return domain
}
