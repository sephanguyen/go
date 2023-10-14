package entities

import "github.com/jackc/pgtype"

type StudentTopicCompleteness struct {
	StudentID        pgtype.Text
	TopicID          pgtype.Text
	TotalFinishedLOs pgtype.Int4
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	IsCompleted      pgtype.Bool
}

func (t *StudentTopicCompleteness) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id", "topic_id", "total_finished_los", "created_at", "updated_at", "is_completed",
		}, []interface{}{
			&t.StudentID, &t.TopicID, &t.TotalFinishedLOs, &t.CreatedAt, &t.UpdatedAt, &t.IsCompleted,
		}
}

func (t *StudentTopicCompleteness) TableName() string {
	return "students_topics_completeness"
}
