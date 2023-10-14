package model

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type EmailRecipient struct {
	EmailRecipientID pgtype.Text
	EmailID          pgtype.Text
	RecipientAddress pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

func (*EmailRecipient) TableName() string {
	return "email_recipients"
}

func (e *EmailRecipient) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"id",
		"email_id",
		"recipient_address",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.EmailRecipientID,
		&e.EmailID,
		&e.RecipientAddress,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

type EmailRecipients []*EmailRecipient

func (u *EmailRecipients) Add() database.Entity {
	e := &EmailRecipient{}
	*u = append(*u, e)

	return e
}
