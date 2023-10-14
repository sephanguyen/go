package entities

import "github.com/jackc/pgtype"

type StudentTopicOverdue struct {
	StudentID pgtype.Text
	TopicID   pgtype.Text
	DueDate   pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func (t *StudentTopicOverdue) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id", "topic_id", "due_date", "created_at", "updated_at",
		}, []interface{}{
			&t.StudentID, &t.TopicID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
		}
}

func (t *StudentTopicOverdue) TableName() string {
	return "students_topics_overdue"
}
