package entities

import "github.com/jackc/pgtype"

// TopicsAssignments reflect topics_assignments table
type TopicsAssignments struct {
	TopicID      pgtype.Text
	AssignmentID pgtype.Text
	DisplayOrder pgtype.Int2
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
}

// FieldMap topics table data fields
func (t *TopicsAssignments) FieldMap() ([]string, []interface{}) {
	return []string{
			"topic_id",
			"assignment_id",
			"display_order",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&t.TopicID,
			&t.AssignmentID,
			&t.DisplayOrder,
			&t.CreatedAt,
			&t.UpdatedAt,
			&t.DeletedAt,
		}
}

// TableName returns "topics_assignments"
func (t *TopicsAssignments) TableName() string {
	return "topics_assignments"
}
