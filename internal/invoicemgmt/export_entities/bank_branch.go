package export

type BankBranchExport struct {
	BankBranchID           string
	BankBranchCode         string
	BankBranchName         string
	BankBranchPhoneticName string
	BankCode               string
	IsArchived             bool
}

func (e *BankBranchExport) FieldMap() ([]string, []interface{}) {
	return []string{
			"bank_branch_id",
			"bank_branch_code",
			"bank_branch_name",
			"bank_branch_phonetic_name",
			"bank_code",
			"is_archived",
		}, []interface{}{
			&e.BankBranchID,
			&e.BankBranchCode,
			&e.BankBranchName,
			&e.BankBranchPhoneticName,
			&e.BankCode,
			&e.IsArchived,
		}
}

func (e *BankBranchExport) TableName() string {
	return "bank_branch"
}
