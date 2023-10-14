package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type ImportUserEvent struct {
	ID             pgtype.Int8
	UserID         pgtype.Text
	Status         pgtype.Text
	Payload        pgtype.JSONB
	ImporterID     pgtype.Text
	SequenceNumber pgtype.Int8
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	ResourcePath   pgtype.Text
}

func (e *ImportUserEvent) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"import_user_event_id", "user_id", "status", "payload", "importer_id", "sequence_number", "created_at", "updated_at", "resource_path"}
	values = []interface{}{&e.ID, &e.UserID, &e.Status, &e.Payload, &e.ImporterID, &e.SequenceNumber, &e.CreatedAt, &e.UpdatedAt, &e.ResourcePath}
	return
}

func (*ImportUserEvent) TableName() string {
	return "import_user_event"
}

type ImportUserEvents []*ImportUserEvent

func (importUserEvents *ImportUserEvents) Add() database.Entity {
	e := &ImportUserEvent{}
	*importUserEvents = append(*importUserEvents, e)

	return e
}

func (importUserEvents ImportUserEvents) IDs() []int64 {
	importUserEventIDs := []int64{}
	for _, importUserEvent := range importUserEvents {
		importUserEventIDs = append(importUserEventIDs, importUserEvent.ID.Int)
	}
	return importUserEventIDs
}
