package entity

import (
	"github.com/jackc/pgtype"
)

type StudentEnrollmentStatusHistory struct {
	StudentID           pgtype.Text
	LocationID          pgtype.Text
	EnrollmentStatus    pgtype.Text
	StartDate           pgtype.Timestamptz
	EndDate             pgtype.Timestamptz
	Comment             pgtype.Text
	CreatedAt           pgtype.Timestamptz
	UpdatedAt           pgtype.Timestamptz
	DeletedAt           pgtype.Timestamptz
	OrderID             pgtype.Text
	OrderSequenceNumber pgtype.Int4
}

// FieldMap return a map of field name and pointer to field
func (e *StudentEnrollmentStatusHistory) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"location_id",
			"enrollment_status",
			"start_date",
			"end_date",
			"comment",
			"created_at",
			"updated_at",
			"deleted_at",
			"order_id",
			"order_sequence_number",
		}, []interface{}{
			&e.StudentID,
			&e.LocationID,
			&e.EnrollmentStatus,
			&e.StartDate,
			&e.EndDate,
			&e.Comment,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.OrderID,
			&e.OrderSequenceNumber,
		}
}

// TableName returning "student_comments"
func (e *StudentEnrollmentStatusHistory) TableName() string {
	return "student_enrollment_status_history"
}
