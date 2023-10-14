package model

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Email struct {
	EmailID           pgtype.Text
	SendGridMessageID pgtype.Text
	Subject           pgtype.Text
	Content           pgtype.JSONB
	EmailFrom         pgtype.Text
	Status            pgtype.Text
	EmailRecipients   pgtype.TextArray
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
}

func (*Email) TableName() string {
	return "emails"
}

func (e *Email) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"email_id",
		"sg_message_id",
		"subject",
		"content",
		"email_from",
		"status",
		"email_recipients",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.EmailID,
		&e.SendGridMessageID,
		&e.Subject,
		&e.Content,
		&e.EmailFrom,
		&e.Status,
		&e.EmailRecipients,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

type Emails []*Email

func (u *Emails) Add() database.Entity {
	e := &Email{}
	*u = append(*u, e)

	return e
}

type EmailContent struct {
	PlainTextContent string `json:"plain_text_content"`
	HTMLContent      string `json:"html_content"`
}
