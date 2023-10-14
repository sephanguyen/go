package lesson

import (
	"github.com/jackc/pgtype"
)

type PrivateConversationLesson struct {
	ConversationID  pgtype.Text
	LessonID        pgtype.Text
	FlattenUserIds  pgtype.Text
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	DeletedAt       pgtype.Timestamptz
	LatestStartTime pgtype.Timestamptz
}

func (c *PrivateConversationLesson) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"conversation_id", "lesson_id", "flatten_user_ids", "created_at", "updated_at", "deleted_at", "latest_start_time"}
	values = []interface{}{&c.ConversationID, &c.LessonID, &c.FlattenUserIds, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.LatestStartTime}
	return
}

func (*PrivateConversationLesson) TableName() string {
	return "private_conversation_lesson"
}
