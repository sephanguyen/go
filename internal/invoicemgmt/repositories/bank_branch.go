package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	export_entities "github.com/manabie-com/backend/internal/invoicemgmt/export_entities"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type BankBranchRepo struct {
}

func (r *BankBranchRepo) FindRelatedBankOfBankBranches(ctx context.Context, db database.QueryExecer, branchIDs []string) ([]*entities.BankRelationMap, error) {
	_, span := interceptors.StartSpan(ctx, "BankBranchRepo.FindRelatedBankOfBankBranches")
	defer span.End()

	query := `
			SELECT 
				bb.bank_branch_id,
				bb.bank_branch_code,
				bb.bank_branch_name,
				bb.bank_branch_phonetic_name,
				b.bank_id,
				b.bank_code,
				b.bank_name,
				b.bank_name_phonetic,
				pb.partner_bank_id,
				pb.bank_number,
				pb.bank_name,
				pb.bank_branch_number,
				pb.bank_branch_name,
				pb.deposit_items,
				pb.account_number,
				pb.consignor_code,
				pb.consignor_name,
				pb.record_limit
			FROM bank_branch bb
			INNER JOIN bank b
				ON bb.bank_id = b.bank_id
			INNER JOIN bank_mapping bm
				ON bm.bank_id = b.bank_id
			INNER JOIN partner_bank pb
				ON pb.partner_bank_id = bm.partner_bank_id
			WHERE bb.bank_branch_id = ANY($1)
	`

	var ids pgtype.TextArray
	_ = ids.Set(branchIDs)

	rows, err := db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}

	results := []*entities.BankRelationMap{}
	defer rows.Close()
	for rows.Next() {
		e := &entities.BankRelationMap{
			BankBranch:  &entities.BankBranch{},
			Bank:        &entities.Bank{},
			PartnerBank: &entities.PartnerBank{},
		}

		err := rows.Scan(
			&e.BankBranch.BankBranchID,
			&e.BankBranch.BankBranchCode,
			&e.BankBranch.BankBranchName,
			&e.BankBranch.BankBranchPhoneticName,
			&e.Bank.BankID,
			&e.Bank.BankCode,
			&e.Bank.BankName,
			&e.Bank.BankNamePhonetic,
			&e.PartnerBank.PartnerBankID,
			&e.PartnerBank.BankNumber,
			&e.PartnerBank.BankName,
			&e.PartnerBank.BankBranchNumber,
			&e.PartnerBank.BankBranchName,
			&e.PartnerBank.DepositItems,
			&e.PartnerBank.AccountNumber,
			&e.PartnerBank.ConsignorCode,
			&e.PartnerBank.ConsignorName,
			&e.PartnerBank.RecordLimit,
		)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		results = append(results, e)
	}

	return results, nil
}

func (r *BankBranchRepo) FindExportableBankBranches(ctx context.Context, db database.QueryExecer) ([]*export_entities.BankBranchExport, error) {
	ctx, span := interceptors.StartSpan(ctx, "BankBranchRepo.FindExportableBankBranches")
	defer span.End()

	resourcePath := golibs.ResourcePathFromCtx(ctx)

	stmt := `
			SELECT 
				bb.bank_branch_id,
				bb.bank_branch_code,
				bb.bank_branch_name,
				bb.bank_branch_phonetic_name,
				b.bank_code,
				bb.is_archived
			FROM bank_branch bb
			INNER JOIN bank b
				ON bb.bank_id = b.bank_id
			WHERE bb.resource_path = $1
	`

	rows, err := db.Query(ctx, stmt, resourcePath)
	if err != nil {
		return nil, err
	}

	bankBranchExports := []*export_entities.BankBranchExport{}
	defer rows.Close()

	for rows.Next() {
		e := &export_entities.BankBranchExport{}

		err := rows.Scan(
			&e.BankBranchID,
			&e.BankBranchCode,
			&e.BankBranchName,
			&e.BankBranchPhoneticName,
			&e.BankCode,
			&e.IsArchived,
		)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		bankBranchExports = append(bankBranchExports, e)
	}

	return bankBranchExports, nil
}

func (r *BankBranchRepo) FindByID(ctx context.Context, db database.QueryExecer, bankBranchID string) (*entities.BankBranch, error) {
	ctx, span := interceptors.StartSpan(ctx, "BankBranchRepo.FindByID")
	defer span.End()

	bankBranch := &entities.BankBranch{}
	fields, _ := bankBranch.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE bank_branch_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), bankBranch.TableName())

	err := database.Select(ctx, db, query, bankBranchID).ScanOne(bankBranch)

	switch err {
	case nil:
		return bankBranch, nil
	case pgx.ErrNoRows:
		return nil, err
	default:
		return nil, fmt.Errorf("err FindByID BankBranchRepo: %w", err)
	}
}

func (r *BankBranchRepo) FindByBankBranchCodeAndBank(ctx context.Context, db database.QueryExecer, bankBranchCode, bankID string) (*entities.BankBranch, error) {
	ctx, span := interceptors.StartSpan(ctx, "BankBranchRepo.FindByBankBranchCodeAndBank")
	defer span.End()

	bankBranch := &entities.BankBranch{}
	fields, _ := bankBranch.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE bank_branch_code = $1 AND bank_id = $2 AND deleted_at IS NULL AND is_archived = false", strings.Join(fields, ","), bankBranch.TableName())

	err := database.Select(ctx, db, query, bankBranchCode, bankID).ScanOne(bankBranch)

	switch err {
	case nil:
		return bankBranch, nil
	case pgx.ErrNoRows:
		return nil, err
	default:
		return nil, fmt.Errorf("err FindByBankBranchCodeAndBank BankBranchRepo: %w", err)
	}
}
