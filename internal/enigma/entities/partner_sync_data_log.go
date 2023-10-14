package entities

import (
	"github.com/jackc/pgtype"
)

type PartnerSyncDataLog struct {
	PartnerSyncDataLogID pgtype.Text
	Signature            pgtype.Text
	Payload              pgtype.JSONB
	UpdatedAt            pgtype.Timestamptz
	CreatedAt            pgtype.Timestamptz
}

func (p *PartnerSyncDataLog) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"partner_sync_data_log_id", "signature", "payload", "updated_at", "created_at"}
	values = []interface{}{&p.PartnerSyncDataLogID, &p.Signature, &p.Payload, &p.UpdatedAt, &p.CreatedAt}
	return
}

func (*PartnerSyncDataLog) TableName() string {
	return "partner_sync_data_log"
}
