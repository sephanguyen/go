package entity

import (
	"github.com/jackc/pgtype"
)

type StudentComment struct {
	CommentID      pgtype.Text
	StudentID      pgtype.Text
	CoachID        pgtype.Text
	CommentContent pgtype.Text
	UpdatedAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
}

// FieldMap return a map of field name and pointer to field
func (e *StudentComment) FieldMap() ([]string, []interface{}) {
	return []string{
			"comment_id", "student_id", "coach_id", "comment_content", "updated_at", "created_at",
		}, []interface{}{
			&e.CommentID, &e.StudentID, &e.CoachID, &e.CommentContent, &e.UpdatedAt, &e.CreatedAt,
		}
}

// TableName returning "student_comments"
func (e *StudentComment) TableName() string {
	return "student_comments"
}
