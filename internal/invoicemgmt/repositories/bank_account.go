package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type BankAccountRepo struct {
}

func (r *BankAccountRepo) FindByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.BankAccount, error) {
	ctx, span := interceptors.StartSpan(ctx, "BankAccountRepo.FindByStudentID")
	defer span.End()

	e := &entities.BankAccount{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), e.TableName())

	err := database.Select(ctx, db, query, studentID).ScanOne(e)

	switch err {
	case nil:
		return e, nil
	case pgx.ErrNoRows:
		return nil, err
	default:
		return nil, fmt.Errorf("err FindByStudentID BankAccountRepo: %w", err)
	}
}

func (r *BankAccountRepo) FindByID(ctx context.Context, db database.QueryExecer, bankAccountID string) (*entities.BankAccount, error) {
	ctx, span := interceptors.StartSpan(ctx, "BankAccountRepo.FindByID")
	defer span.End()

	bankAccount := &entities.BankAccount{}
	fields, _ := bankAccount.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE bank_account_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), bankAccount.TableName())

	err := database.Select(ctx, db, query, bankAccountID).ScanOne(bankAccount)

	switch err {
	case nil:
		return bankAccount, nil
	case pgx.ErrNoRows:
		return nil, err
	default:
		return nil, fmt.Errorf("err FindByID BankAccountRepo: %w", err)
	}
}

func (r *BankAccountRepo) Upsert(ctx context.Context, db database.QueryExecer, bankAccounts ...*entities.BankAccount) error {
	ctx, span := interceptors.StartSpan(ctx, "BankAccountRepo.Upsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, bankAccount *entities.BankAccount) {
		fields := database.GetFieldNames(bankAccount)
		fields = utils.RemoveStrFromSlice(fields, "resource_path")
		values := database.GetScanFields(bankAccount, fields)

		placeHolders := database.GeneratePlaceholders(len(fields))

		stmt :=
			`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT bank_account__pk 
			DO UPDATE SET 
				is_verified = EXCLUDED.is_verified, 
				bank_id = EXCLUDED.bank_id, 
				bank_branch_id = EXCLUDED.bank_branch_id, 
				bank_account_holder = EXCLUDED.bank_account_holder, 
				bank_account_number = EXCLUDED.bank_account_number, 
				bank_account_type = EXCLUDED.bank_account_type, 
				updated_at = now(), 
				deleted_at = NULL
			`

		stmt = fmt.Sprintf(stmt, bankAccount.TableName(), strings.Join(fields, ","), placeHolders)
		b.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}

	for _, bankAccount := range bankAccounts {
		queueFn(batch, bankAccount)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for i := 0; i < len(bankAccounts); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("no rows affected when upserting bank_account")
		}
	}

	return nil
}
