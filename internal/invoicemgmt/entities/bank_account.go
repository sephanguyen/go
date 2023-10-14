package entities

import "github.com/jackc/pgtype"

type BankAccount struct {
	BankAccountID          pgtype.Text
	StudentPaymentDetailID pgtype.Text
	StudentID              pgtype.Text
	IsVerified             pgtype.Bool
	BankID                 pgtype.Text
	BankBranchID           pgtype.Text
	BankAccountNumber      pgtype.Text
	BankAccountHolder      pgtype.Text
	BankAccountType        pgtype.Text
	CreatedAt              pgtype.Timestamptz
	UpdatedAt              pgtype.Timestamptz
	DeletedAt              pgtype.Timestamptz
	MigratedAt             pgtype.Timestamptz
	ResourcePath           pgtype.Text
}

func (e *BankAccount) FieldMap() ([]string, []interface{}) {
	return []string{
			"bank_account_id",
			"student_payment_detail_id",
			"student_id",
			"is_verified",
			"bank_id",
			"bank_branch_id",
			"bank_account_number",
			"bank_account_holder",
			"bank_account_type",
			"created_at",
			"updated_at",
			"deleted_at",
			"migrated_at",
			"resource_path",
		}, []interface{}{
			&e.BankAccountID,
			&e.StudentPaymentDetailID,
			&e.StudentID,
			&e.IsVerified,
			&e.BankID,
			&e.BankBranchID,
			&e.BankAccountNumber,
			&e.BankAccountHolder,
			&e.BankAccountType,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.MigratedAt,
			&e.ResourcePath,
		}
}

func (e *BankAccount) TableName() string {
	return "bank_account"
}
