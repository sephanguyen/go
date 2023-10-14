package entities

import "github.com/jackc/pgtype"

type StudentEnrollmentStatusHistory struct {
	StudentID           pgtype.Text
	LocationID          pgtype.Text
	EnrollmentStatus    pgtype.Text
	StartDate           pgtype.Timestamptz
	EndDate             pgtype.Timestamptz
	Comment             pgtype.Text
	OrderID             pgtype.Text
	OrderSequenceNumber pgtype.Int4
	CreatedAt           pgtype.Timestamptz
	UpdatedAt           pgtype.Timestamptz
}

func (s *StudentEnrollmentStatusHistory) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"location_id",
			"enrollment_status",
			"start_date",
			"end_date",
			"comment",
			"order_id",
			"order_sequence_number",
			"created_at",
			"updated_at",
		}, []interface{}{
			&s.StudentID,
			&s.LocationID,
			&s.EnrollmentStatus,
			&s.StartDate,
			&s.EndDate,
			&s.Comment,
			&s.OrderID,
			&s.OrderSequenceNumber,
			&s.CreatedAt,
			&s.UpdatedAt,
		}
}

func (s *StudentEnrollmentStatusHistory) TableName() string {
	return "student_enrollment_status_history"
}
