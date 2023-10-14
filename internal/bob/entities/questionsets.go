package entities

import (
	"github.com/jackc/pgtype"
)

//QuestionSets get renamed
type QuestionSets struct {
	QuestionID   pgtype.Text `sql:"question_id"`
	LoID         pgtype.Text `sql:"lo_id"`
	DisplayOrder pgtype.Int4
	UpdatedAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
}

// FieldMap return a map of field name and pointer to field
func (e *QuestionSets) FieldMap() ([]string, []interface{}) {
	return []string{
			"question_id", "lo_id", "display_order", "updated_at", "created_at",
		}, []interface{}{
			&e.QuestionID, &e.LoID, &e.DisplayOrder, &e.UpdatedAt, &e.CreatedAt,
		}
}

// TableName returns "question"
func (e *QuestionSets) TableName() string {
	return "quizsets"
}
