package entities

import "github.com/jackc/pgtype"

type PartnerBank struct {
	PartnerBankID    pgtype.Text
	ConsignorCode    pgtype.Text
	ConsignorName    pgtype.Text
	BankNumber       pgtype.Text
	BankName         pgtype.Text
	BankBranchNumber pgtype.Text
	BankBranchName   pgtype.Text
	DepositItems     pgtype.Text
	AccountNumber    pgtype.Text
	IsArchived       pgtype.Bool
	Remarks          pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
	ResourcePath     pgtype.Text
	IsDefault        pgtype.Bool
	RecordLimit      pgtype.Int4
}

func (e *PartnerBank) FieldMap() ([]string, []interface{}) {
	return []string{
			"partner_bank_id",
			"consignor_code",
			"consignor_name",
			"bank_number",
			"bank_name",
			"bank_branch_number",
			"bank_branch_name",
			"deposit_items",
			"account_number",
			"is_archived",
			"remarks",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
			"is_default",
			"record_limit",
		}, []interface{}{
			&e.PartnerBankID,
			&e.ConsignorCode,
			&e.ConsignorName,
			&e.BankNumber,
			&e.BankName,
			&e.BankBranchNumber,
			&e.BankBranchName,
			&e.DepositItems,
			&e.AccountNumber,
			&e.IsArchived,
			&e.Remarks,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
			&e.IsDefault,
			&e.RecordLimit,
		}
}

func (e *PartnerBank) TableName() string {
	return "partner_bank"
}

func (e *PartnerBank) UpdateOnConflictQuery() string {
	return `
	consignor_code = EXCLUDED.consignor_code,
	consignor_name = EXCLUDED.consignor_name,
	bank_number = EXCLUDED.bank_number,
	bank_name = EXCLUDED.bank_name,
	bank_branch_number = EXCLUDED.bank_branch_number,
	bank_branch_name = EXCLUDED.bank_branch_name,
	deposit_items = EXCLUDED.deposit_items,
	account_number = EXCLUDED.account_number,
	is_archived = EXCLUDED.is_archived,
	remarks = EXCLUDED.remarks,
	created_at = EXCLUDED.created_at,
	updated_at = EXCLUDED.updated_at`
}

func (e *PartnerBank) PrimaryKey() string {
	return "partner_bank_id"
}

func (e *PartnerBank) FieldMapForUpsert() ([]string, []interface{}) {
	return []string{
			"partner_bank_id",
			"consignor_code",
			"consignor_name",
			"bank_number",
			"bank_name",
			"bank_branch_number",
			"bank_branch_name",
			"deposit_items",
			"account_number",
			"is_archived",
			"remarks",
			"created_at",
			"updated_at",
			"deleted_at",
			"is_default",
			"record_limit",
		}, []interface{}{
			&e.PartnerBankID,
			&e.ConsignorCode,
			&e.ConsignorName,
			&e.BankNumber,
			&e.BankName,
			&e.BankBranchNumber,
			&e.BankBranchName,
			&e.DepositItems,
			&e.AccountNumber,
			&e.IsArchived,
			&e.Remarks,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.IsDefault,
			&e.RecordLimit,
		}
}
