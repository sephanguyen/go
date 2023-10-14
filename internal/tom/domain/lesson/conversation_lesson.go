package lesson

import (
	"github.com/jackc/pgtype"
)

type ConversationLesson struct {
	ConversationID  pgtype.Text
	LessonID        pgtype.Text
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	DeletedAt       pgtype.Timestamptz
	LatestStartTime pgtype.Timestamptz
}

func (c *ConversationLesson) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"conversation_id", "lesson_id", "created_at", "updated_at", "deleted_at", "latest_start_time"}
	values = []interface{}{&c.ConversationID, &c.LessonID, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.LatestStartTime}
	return
}

func (*ConversationLesson) TableName() string {
	return "conversation_lesson"
}
