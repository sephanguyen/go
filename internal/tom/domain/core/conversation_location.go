package core

import "github.com/jackc/pgtype"

type ConversationLocation struct {
	ConversationID pgtype.Text
	LocationID     pgtype.Text
	AccessPath     pgtype.Text
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

func (cap *ConversationLocation) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"conversation_id", "location_id", "access_path", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&cap.ConversationID, &cap.LocationID, &cap.AccessPath, &cap.CreatedAt, &cap.UpdatedAt, &cap.DeletedAt}
	return
}

func (*ConversationLocation) TableName() string {
	return "conversation_locations"
}
