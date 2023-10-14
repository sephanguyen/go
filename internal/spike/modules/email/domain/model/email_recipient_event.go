package model

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type EmailRecipientEvent struct {
	EmailRecipientEventID pgtype.Text
	EmailRecipientID      pgtype.Text
	Type                  pgtype.Text
	Event                 pgtype.Text
	Description           pgtype.JSONB
	CreatedAt             pgtype.Timestamptz
	UpdatedAt             pgtype.Timestamptz
	DeletedAt             pgtype.Timestamptz
}

func (*EmailRecipientEvent) TableName() string {
	return "email_recipient_events"
}

func (e *EmailRecipientEvent) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"id",
		"email_recipient_id",
		"type",
		"event",
		"description",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.EmailRecipientEventID,
		&e.EmailRecipientID,
		&e.Type,
		&e.Event,
		&e.Description,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

func (e *EmailRecipientEvent) GetDescription() (*EventDescription, error) {
	ed := &EventDescription{}
	err := e.Description.AssignTo(ed)
	return ed, err
}

type EmailRecipientEvents []*EmailRecipientEvent

func (u *EmailRecipientEvents) Add() database.Entity {
	e := &EmailRecipientEvent{}
	*u = append(*u, e)

	return e
}

type EventDescription struct {
	Event   string                   `json:"event"`
	Details []EventDescriptionDetail `json:"details"`
}

type EventDescriptionDetail struct {
	SGEventID            string `json:"sg_event_id"`
	Type                 string `json:"type"`
	BounceClassification string `json:"bounce_classification"`
	Status               string `json:"status"`
	Reason               string `json:"reason"`
	Response             string `json:"response"`
	Attempt              string `json:"attempt"`
	Timestamp            int64  `json:"timestamp"`
	SGMessageID          string `json:"sg_message_id"`
}
