package entities

import "github.com/jackc/pgtype"

const (
	StudentAssignmentStatusActive           = "STUDENT_ASSIGNMENT_STATUS_ACTIVE"
	StudentAssignmentStatusCompleted        = "STUDENT_ASSIGNMENT_STATUS_COMPLETED"
	StudentAssignmentStatusRemovedFromClass = "STUDENT_ASSIGNMENT_STATUS_REMOVED_FROM_CLASS"
)

type StudentAssignment struct {
	AssignmentID     pgtype.Text
	StudentID        pgtype.Text
	AssignmentStatus pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	CompletedAt      pgtype.Timestamptz
}

func (e *StudentAssignment) FieldMap() ([]string, []interface{}) {
	return []string{
			"assignment_id", "student_id", "assignment_status", "updated_at", "created_at", "completed_at",
		}, []interface{}{
			&e.AssignmentID, &e.StudentID, &e.AssignmentStatus, &e.UpdatedAt, &e.CreatedAt, &e.CompletedAt,
		}
}

func (e *StudentAssignment) TableName() string {
	return "student_assignments"
}
