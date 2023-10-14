package core

import (
	"github.com/jackc/pgtype"
)

type FindMessagesArgs struct {
	ConversationID pgtype.Text
	EndAt          pgtype.Timestamptz
	Limit          uint32

	// SysStem message is ignored by default
	IncludeSystemMsg bool

	// IncludeMessageTypes and ExcludeMessageTypes can't be co-exist.
	IncludeMessageTypes  pgtype.TextArray
	ExcludeMessagesTypes pgtype.TextArray
}

type Message struct {
	ID             pgtype.Text
	ConversationID pgtype.Text
	UserID         pgtype.Text
	Message        pgtype.Text
	UrlMedia       pgtype.Text
	Type           pgtype.Text
	DeletedBy      pgtype.Text
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
	TargetUser     pgtype.Text
}

func (e *Message) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"message_id", "conversation_id", "user_id", "message", "url_media", "type", "deleted_by", "created_at", "updated_at", "deleted_at", "target_user"}
	values = []interface{}{&e.ID, &e.ConversationID, &e.UserID, &e.Message, &e.UrlMedia, &e.Type, &e.DeletedBy, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt, &e.TargetUser}
	return
}

func (*Message) TableName() string {
	return "messages"
}
