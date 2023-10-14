package entities

import "github.com/jackc/pgtype"

// TopicsLearningObjectives reflect topics_learning_objectives table
type TopicsLearningObjectives struct {
	TopicID      pgtype.Text
	LoID         pgtype.Text
	DisplayOrder pgtype.Int2
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
}

// FieldMap topics table data fields
func (t *TopicsLearningObjectives) FieldMap() ([]string, []interface{}) {
	return []string{
			"topic_id",
			"lo_id",
			"display_order",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&t.TopicID,
			&t.LoID,
			&t.DisplayOrder,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.DeletedAt,
		}
}

// TableName returns "topics_learning_objectives"
func (t *TopicsLearningObjectives) TableName() string {
	return "topics_learning_objectives"
}
