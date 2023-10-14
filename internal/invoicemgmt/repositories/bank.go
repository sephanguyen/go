package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"github.com/jackc/pgx/v4"
)

type BankRepo struct {
}

func (r *BankRepo) FindAll(ctx context.Context, db database.QueryExecer) ([]*entities.Bank, error) {
	ctx, span := interceptors.StartSpan(ctx, "BankRepo.FindAll")
	defer span.End()

	e := &entities.Bank{}
	fields, _ := e.FieldMap()

	resourcePath := golibs.ResourcePathFromCtx(ctx)

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE resource_path = $1", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(ctx, stmt, resourcePath)
	if err != nil {
		return nil, err
	}

	banks := []*entities.Bank{}
	defer rows.Close()
	for rows.Next() {
		bank := new(entities.Bank)
		database.AllNullEntity(bank)

		_, fieldValues := bank.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		banks = append(banks, bank)
	}

	return banks, nil
}

func (r *BankRepo) FindByID(ctx context.Context, db database.QueryExecer, bankID string) (*entities.Bank, error) {
	ctx, span := interceptors.StartSpan(ctx, "BankBranchRepo.FindByID")
	defer span.End()

	bank := &entities.Bank{}
	fields, _ := bank.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE bank_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), bank.TableName())

	err := database.Select(ctx, db, query, bankID).ScanOne(bank)

	switch err {
	case nil:
		return bank, nil
	case pgx.ErrNoRows:
		return nil, err
	default:
		return nil, fmt.Errorf("err FindByID BankRepo: %w", err)
	}
}

func (r *BankRepo) FindByBankCode(ctx context.Context, db database.QueryExecer, bankCode string) (*entities.Bank, error) {
	ctx, span := interceptors.StartSpan(ctx, "BankBranchRepo.FindByBankCode")
	defer span.End()

	bank := &entities.Bank{}
	fields, _ := bank.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE bank_code = $1 AND deleted_at IS NULL AND is_archived = false", strings.Join(fields, ","), bank.TableName())

	err := database.Select(ctx, db, query, bankCode).ScanOne(bank)

	switch err {
	case nil:
		return bank, nil
	case pgx.ErrNoRows:
		return nil, err
	default:
		return nil, fmt.Errorf("err FindByBankCode BankRepo: %w", err)
	}
}
