package entities

import "github.com/jackc/pgtype"

type Bank struct {
	BankID           pgtype.Text
	BankCode         pgtype.Text
	BankName         pgtype.Text
	BankNamePhonetic pgtype.Text
	IsArchived       pgtype.Bool
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
	ResourcePath     pgtype.Text
}

func (e *Bank) FieldMap() ([]string, []interface{}) {
	return []string{
			"bank_id",
			"bank_code",
			"bank_name",
			"bank_name_phonetic",
			"is_archived",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.BankID,
			&e.BankCode,
			&e.BankName,
			&e.BankNamePhonetic,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *Bank) TableName() string {
	return "bank"
}
