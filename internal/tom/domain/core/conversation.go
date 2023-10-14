package core

import (
	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type Conversation struct {
	ID               pgtype.Text
	Name             pgtype.Text
	Status           pgtype.Text
	ConversationType pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	LastMessageID    pgtype.Text
	Owner            pgtype.Text
}

func (c *Conversation) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"conversation_id", "name", "status", "conversation_type", "created_at", "updated_at", "last_message_id", "owner"}
	values = []interface{}{&c.ID, &c.Name, &c.Status, &c.ConversationType, &c.CreatedAt, &c.UpdatedAt, &c.LastMessageID, &c.Owner}
	return
}

func (*Conversation) TableName() string {
	return "conversations"
}

type Conversations []*Conversation

func (u *Conversations) Add() database.Entity {
	e := &Conversation{}
	*u = append(*u, e)

	return e
}

type ConversationFull struct {
	Conversation Conversation
	// StudentQuestionID pgtype.Text
	// ClassID           pgtype.Int4
	IsReply   pgtype.Bool
	StudentID pgtype.Text
}
