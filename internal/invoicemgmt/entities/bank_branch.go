package entities

import "github.com/jackc/pgtype"

type BankBranch struct {
	BankBranchID           pgtype.Text
	BankBranchCode         pgtype.Text
	BankBranchName         pgtype.Text
	BankBranchPhoneticName pgtype.Text
	BankID                 pgtype.Text
	IsArchived             pgtype.Bool
	CreatedAt              pgtype.Timestamptz
	UpdatedAt              pgtype.Timestamptz
	DeletedAt              pgtype.Timestamptz
	ResourcePath           pgtype.Text
}

func (e *BankBranch) FieldMap() ([]string, []interface{}) {
	return []string{
			"bank_branch_id",
			"bank_branch_code",
			"bank_branch_name",
			"bank_branch_phonetic_name",
			"bank_id",
			"is_archived",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.BankBranchID,
			&e.BankBranchCode,
			&e.BankBranchName,
			&e.BankBranchPhoneticName,
			&e.BankID,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *BankBranch) TableName() string {
	return "bank_branch"
}

type BankRelationMap struct {
	BankBranch  *BankBranch
	Bank        *Bank
	PartnerBank *PartnerBank
}
