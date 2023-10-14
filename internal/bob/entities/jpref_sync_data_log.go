package entities

import (
	"github.com/jackc/pgtype"
)

type JprepSyncDataLog struct {
	ID        pgtype.Text
	Signature pgtype.Text
	Payload   pgtype.JSONB
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
}

func (rcv *JprepSyncDataLog) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"jpref_sync_data_log_id", "signature", "payload", "updated_at", "created_at"}
	values = []interface{}{&rcv.ID, &rcv.Signature, &rcv.Payload, &rcv.UpdatedAt, &rcv.CreatedAt}
	return
}

func (*JprepSyncDataLog) TableName() string {
	return "jpref_sync_data_logs"
}
