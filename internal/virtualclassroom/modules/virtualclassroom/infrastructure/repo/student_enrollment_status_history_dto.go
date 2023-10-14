package repo

import (
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
)

type StudentEnrollmentStatusHistory struct {
	StudentID        pgtype.Text
	LocationID       pgtype.Text
	EnrollmentStatus pgtype.Text
	StartDate        pgtype.Timestamptz
	EndDate          pgtype.Timestamptz
	Comment          pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
}

func (s *StudentEnrollmentStatusHistory) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"location_id",
			"enrollment_status",
			"start_date",
			"end_date",
			"comment",
			"created_at",
			"updated_at",
		}, []interface{}{
			&s.StudentID,
			&s.LocationID,
			&s.EnrollmentStatus,
			&s.StartDate,
			&s.EndDate,
			&s.Comment,
			&s.CreatedAt,
			&s.UpdatedAt,
		}
}

func (s *StudentEnrollmentStatusHistory) TableName() string {
	return "student_enrollment_status_history"
}

func (s *StudentEnrollmentStatusHistory) ToStudentEnrollmentStatusHistoryDomain() *domain.StudentEnrollmentStatusHistory {
	return &domain.StudentEnrollmentStatusHistory{
		StudentID:        s.StudentID.String,
		LocationID:       s.LocationID.String,
		EnrollmentStatus: s.EnrollmentStatus.String,
		StartDate:        s.StartDate.Time,
		EndDate:          s.EndDate.Time,
		Comment:          s.Comment.String,
		CreatedAt:        s.CreatedAt.Time,
		UpdatedAt:        s.UpdatedAt.Time,
	}
}
