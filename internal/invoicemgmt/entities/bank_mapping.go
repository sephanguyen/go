package entities

import "github.com/jackc/pgtype"

type BankMapping struct {
	BankID        pgtype.Text
	BankMappingID pgtype.Text
	PartnerBankID pgtype.Text
	Remarks       pgtype.Text
	IsArchived    pgtype.Bool
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
	ResourcePath  pgtype.Text
}

func (e *BankMapping) FieldMap() ([]string, []interface{}) {
	return []string{
			"bank_mapping_id",
			"bank_id",
			"partner_bank_id",
			"remarks",
			"is_archived",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.BankMappingID,
			&e.BankID,
			&e.PartnerBankID,
			&e.Remarks,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *BankMapping) TableName() string {
	return "bank_mapping"
}
