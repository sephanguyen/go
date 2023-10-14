package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type StudentPaymentDetailRepo struct {
}

func (r *StudentPaymentDetailRepo) FindByStudentID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.StudentPaymentDetail, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentPaymentDetailRepo.FindByStudentID")
	defer span.End()

	e := &entities.StudentPaymentDetail{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), e.TableName())

	err := database.Select(ctx, db, query, studentID).ScanOne(e)

	switch err {
	case nil:
		return e, nil
	case pgx.ErrNoRows:
		return nil, err
	default:
		return nil, fmt.Errorf("err FindByStudentID StudentPaymentDetailRepo: %w", err)
	}
}

func (r *StudentPaymentDetailRepo) FindStudentBillingByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.StudentBillingDetailsMap, error) {
	_, span := interceptors.StartSpan(ctx, "StudentPaymentDetailRepo.FindStudentBillingByStudentIDs")
	defer span.End()

	query := `
			SELECT 
				spd.student_payment_detail_id,
				spd.student_id,
				spd.payer_name,
				spd.payer_phone_number,
				spd.payment_method,
				ba.billing_address_id,
				ba.user_id,
				ba.postal_code,
				ba.prefecture_code,
				ba.city,
				ba.street1,
				ba.street2
			FROM student_payment_detail spd
			INNER JOIN billing_address ba
				ON spd.student_payment_detail_id = ba.student_payment_detail_id
			WHERE ba.user_id = ANY($1)
	`

	var ids pgtype.TextArray
	_ = ids.Set(studentIDs)

	rows, err := db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}

	results := []*entities.StudentBillingDetailsMap{}
	defer rows.Close()
	for rows.Next() {
		e := &entities.StudentBillingDetailsMap{
			StudentPaymentDetail: &entities.StudentPaymentDetail{},
			BillingAddress:       &entities.BillingAddress{},
		}

		err := rows.Scan(
			&e.StudentPaymentDetail.StudentPaymentDetailID,
			&e.StudentPaymentDetail.StudentID,
			&e.StudentPaymentDetail.PayerName,
			&e.StudentPaymentDetail.PayerPhoneNumber,
			&e.StudentPaymentDetail.PaymentMethod,
			&e.BillingAddress.BillingAddressID,
			&e.BillingAddress.UserID,
			&e.BillingAddress.PostalCode,
			&e.BillingAddress.PrefectureCode,
			&e.BillingAddress.City,
			&e.BillingAddress.Street1,
			&e.BillingAddress.Street2,
		)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		results = append(results, e)
	}

	return results, nil
}

func (r *StudentPaymentDetailRepo) FindStudentBankDetailsByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.StudentBankDetailsMap, error) {
	_, span := interceptors.StartSpan(ctx, "StudentPaymentDetailRepo.FindStudentBankDetailsByStudentIDs")
	defer span.End()

	query := `
			SELECT 
				spd.student_payment_detail_id,
				spd.student_id,
				spd.payer_name,
				spd.payer_phone_number,
				spd.payment_method,
				ba.bank_account_id,
				ba.student_id,
				ba.is_verified,
				ba.bank_account_number,
				ba.bank_account_holder,
				ba.bank_account_type,
				ba.bank_branch_id
			FROM student_payment_detail spd
			INNER JOIN bank_account ba
				ON ba.student_payment_detail_id = spd.student_payment_detail_id
			WHERE ba.student_id = ANY($1)
	`

	var ids pgtype.TextArray
	_ = ids.Set(studentIDs)

	rows, err := db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}

	results := []*entities.StudentBankDetailsMap{}
	defer rows.Close()
	for rows.Next() {
		e := &entities.StudentBankDetailsMap{
			BankAccount:          &entities.BankAccount{},
			StudentPaymentDetail: &entities.StudentPaymentDetail{},
		}

		err := rows.Scan(
			&e.StudentPaymentDetail.StudentPaymentDetailID,
			&e.StudentPaymentDetail.StudentID,
			&e.StudentPaymentDetail.PayerName,
			&e.StudentPaymentDetail.PayerPhoneNumber,
			&e.StudentPaymentDetail.PaymentMethod,
			&e.BankAccount.BankAccountID,
			&e.BankAccount.StudentID,
			&e.BankAccount.IsVerified,
			&e.BankAccount.BankAccountNumber,
			&e.BankAccount.BankAccountHolder,
			&e.BankAccount.BankAccountType,
			&e.BankAccount.BankBranchID,
		)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		results = append(results, e)
	}

	return results, nil
}

func (r *StudentPaymentDetailRepo) FindByID(ctx context.Context, db database.QueryExecer, studentPaymentDetailID string) (*entities.StudentPaymentDetail, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentPaymentDetailRepo.FindByID")
	defer span.End()

	studentPaymentDetail := &entities.StudentPaymentDetail{}
	fields, _ := studentPaymentDetail.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_payment_detail_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), studentPaymentDetail.TableName())

	err := database.Select(ctx, db, query, studentPaymentDetailID).ScanOne(studentPaymentDetail)

	switch err {
	case nil:
		return studentPaymentDetail, nil
	case pgx.ErrNoRows:
		return nil, err
	default:
		return nil, fmt.Errorf("err FindByID StudentPaymentDetailRepo: %w", err)
	}
}

func (r *StudentPaymentDetailRepo) Upsert(ctx context.Context, db database.QueryExecer, studentPaymentDetails ...*entities.StudentPaymentDetail) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentPaymentDetailRepo.Upsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, studentPaymentDetail *entities.StudentPaymentDetail) {
		fields := database.GetFieldNames(studentPaymentDetail)
		fields = utils.RemoveStrFromSlice(fields, "resource_path")
		values := database.GetScanFields(studentPaymentDetail, fields)

		placeHolders := database.GeneratePlaceholders(len(fields))

		stmt :=
			`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT student_payment_detail__pk 
			DO UPDATE SET 
				payer_name = EXCLUDED.payer_name, 
				payer_phone_number = EXCLUDED.payer_phone_number, 
				payment_method = EXCLUDED.payment_method, 
				updated_at = now(), 
				deleted_at = NULL
			`

		stmt = fmt.Sprintf(stmt, studentPaymentDetail.TableName(), strings.Join(fields, ","), placeHolders)
		b.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}

	for _, studentPaymentDetail := range studentPaymentDetails {
		queueFn(batch, studentPaymentDetail)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for i := 0; i < len(studentPaymentDetails); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("no rows affected when upserting student_payment_detail")
		}
	}

	return nil
}

func (r *StudentPaymentDetailRepo) SoftDelete(ctx context.Context, db database.QueryExecer, studentPaymentDetailIDs ...string) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentPaymentDetailRepo.SoftDelete")
	defer span.End()

	stmt :=
		`
		UPDATE %s SET deleted_at = now() WHERE student_payment_detail_id = ANY($1) AND deleted_at IS NULL
		`

	stmt = fmt.Sprintf(stmt, (&entities.StudentPaymentDetail{}).TableName())
	_, err := db.Exec(ctx, stmt, database.TextArray(studentPaymentDetailIDs))
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (r *StudentPaymentDetailRepo) FindFromInvoiceIDTempTable(ctx context.Context, db database.QueryExecer) ([]*entities.StudentPaymentDetail, error) {
	_, span := interceptors.StartSpan(ctx, "StudentPaymentDetailRepo.FindFromInvoiceIDTempTable")
	defer span.End()

	e := &entities.StudentPaymentDetail{}
	fields, _ := e.FieldMap()

	newFields := make([]string, len(fields))
	for i, f := range fields {
		newFields[i] = fmt.Sprintf("spd.%v", f)
	}

	stmt := fmt.Sprintf(`
		SELECT %s 
		FROM student_payment_detail spd
		INNER JOIN invoice i
			ON i.student_id = spd.student_id
		INNER JOIN temporary_invoice_id_table ti
			ON i.invoice_id = ti.temp_invoice_id
	`, strings.Join(newFields, ","))

	rows, err := db.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	results := []*entities.StudentPaymentDetail{}
	defer rows.Close()
	for rows.Next() {
		spd := new(entities.StudentPaymentDetail)
		_, values := spd.FieldMap()

		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		results = append(results, spd)
	}

	return results, nil
}
