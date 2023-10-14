package entities

import (
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
			"student_id", "location_id", "enrollment_status", "start_date", "end_date", "comment", "created_at", "updated_at",
		}, []interface{}{
			&s.StudentID, &s.LocationID, &s.EnrollmentStatus, &s.StartDate, &s.EndDate, &s.Comment, &s.CreatedAt, &s.UpdatedAt,
		}
}

func (s *StudentEnrollmentStatusHistory) TableName() string {
	return "student_enrollment_status_history"
}
