package entities

import "github.com/jackc/pgtype"

type ContentBankMedia struct {
	ID            pgtype.Text        `json:"id"`
	Name          pgtype.Text        `json:"name"`
	Resource      pgtype.Text        `json:"resource"`
	Type          pgtype.Text        `json:"type"`
	FileSizeBytes pgtype.Int8        `json:"file_size_bytes"`
	CreatedBy     pgtype.Text        `json:"created_by"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
}

func (t *ContentBankMedia) FieldMap() ([]string, []interface{}) {
	return []string{"id", "name", "resource", "type", "file_size_bytes", "created_by", "created_at", "updated_at"},
		[]interface{}{&t.ID, &t.Name, &t.Resource, &t.Type, &t.FileSizeBytes, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt}
}

func (t *ContentBankMedia) TableName() string {
	return "content_bank_medias"
}
