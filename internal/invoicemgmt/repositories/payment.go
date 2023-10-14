package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type PaymentRepo struct {
}

func (r *PaymentRepo) GetLatestPaymentDueDateByInvoiceID(ctx context.Context, db database.QueryExecer, invoiceID string) (*entities.Payment, error) {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.GetLatestPaymentDueDateByInvoiceID")
	defer span.End()

	e := &entities.Payment{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_id = $1 ORDER BY created_at DESC LIMIT 1", strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, query, invoiceID).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (r *PaymentRepo) FindByPaymentID(ctx context.Context, db database.QueryExecer, paymentID string) (*entities.Payment, error) {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.FindByPaymentID")
	defer span.End()

	e := &entities.Payment{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE payment_id = $1", strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, query, paymentID).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (r *PaymentRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.Payment) error {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.Create")
	defer span.End()

	now := time.Now()

	// Assign payment ID if the paymentID is not yet set
	errs := []error{e.UpdatedAt.Set(now)}
	if strings.TrimSpace(e.PaymentID.String) == "" {
		errs = append(errs, e.PaymentID.Set(idutil.ULIDNow()))
	}

	if err := multierr.Combine(errs...); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set PaymentID.Set: %w", err)
	}

	if e.CreatedAt.Time.IsZero() {
		err := e.CreatedAt.Set(now)
		if err != nil {
			return err
		}
	}

	// Remove the payment sequence number on insert statement if it is empty
	excludedFields := []string{"resource_path", "payment_sequence_number"}
	if e.PaymentSequenceNumber.Status == pgtype.Null || e.PaymentSequenceNumber.Int == 0 {
		excludedFields = []string{"resource_path"}
	}

	cmdTag, err := database.InsertExcept(ctx, e, excludedFields, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert PaymentRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert PaymentRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *PaymentRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.Payment) error {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "payment_id", []string{"payment_status", "updated_at"})

	if err != nil {
		return fmt.Errorf("err update PaymentRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update PaymentRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *PaymentRepo) UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.Payment, fieldsToUpdate []string) error {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.UpdateWithFields")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "payment_id", fieldsToUpdate)

	if err != nil {
		return fmt.Errorf("err updateWithFields PaymentRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err updateWithFields PaymentRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *PaymentRepo) FindPaymentInvoiceByIDs(ctx context.Context, db database.QueryExecer, paymentIDs []string) ([]*entities.PaymentInvoiceMap, error) {
	_, span := interceptors.StartSpan(ctx, "PaymentRepo.FindPaymentInvoiceByIDs")
	defer span.End()

	query := `
			SELECT 
				p.payment_id,
				p.invoice_id,
				p.payment_method,
				p.payment_due_date,
				p.payment_expiry_date,
				p.payment_date,
				p.payment_status,
				p.payment_sequence_number,
				p.student_id,
				p.is_exported,
				p.amount,
				p.bulk_payment_id,
				i.invoice_id,
				i.student_id,
				i.sub_total,
				i.total,
				i.invoice_sequence_number,
				i.is_exported
			FROM payment p
			INNER JOIN invoice i
				ON p.invoice_id = i.invoice_id
			WHERE p.payment_id = ANY($1)
	`

	var ids pgtype.TextArray
	_ = ids.Set(paymentIDs)

	rows, err := db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}

	results := []*entities.PaymentInvoiceMap{}
	defer rows.Close()
	for rows.Next() {
		e := &entities.PaymentInvoiceMap{
			Payment: &entities.Payment{},
			Invoice: &entities.Invoice{},
		}

		err := rows.Scan(
			&e.Payment.PaymentID,
			&e.Payment.InvoiceID,
			&e.Payment.PaymentMethod,
			&e.Payment.PaymentDueDate,
			&e.Payment.PaymentExpiryDate,
			&e.Payment.PaymentDate,
			&e.Payment.PaymentStatus,
			&e.Payment.PaymentSequenceNumber,
			&e.Payment.StudentID,
			&e.Payment.IsExported,
			&e.Payment.Amount,
			&e.Payment.BulkPaymentID,
			&e.Invoice.InvoiceID,
			&e.Invoice.StudentID,
			&e.Invoice.SubTotal,
			&e.Invoice.Total,
			&e.Invoice.InvoiceSequenceNumber,
			&e.Invoice.IsExported,
		)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		results = append(results, e)
	}

	return results, nil
}

func (r *PaymentRepo) FindByPaymentSequenceNumber(ctx context.Context, db database.QueryExecer, paymentSequenceNumber int) (*entities.Payment, error) {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.FindByPaymentSequenceNumber")
	defer span.End()

	e := &entities.Payment{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE payment_sequence_number = $1", strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, query, paymentSequenceNumber).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (r *PaymentRepo) UpdateIsExportedByPaymentRequestFileID(ctx context.Context, db database.QueryExecer, fileID string, isExported bool) error {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.UpdateIsExportedByPaymentRequestFileID")
	defer span.End()

	query := `
		UPDATE payment
		SET is_exported = $2
		WHERE payment_id IN (
			SELECT payment_id
			FROM bulk_payment_request_file_payment
			WHERE bulk_payment_request_file_id = $1
		)
	`

	cmdTag, err := db.Exec(ctx, query, fileID, isExported)
	if err != nil {
		return fmt.Errorf("err UpdateIsExportedByPaymentRequestFileID PaymentRepo: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("err UpdateIsExportedByPaymentRequestFileID PaymentRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *PaymentRepo) UpdateIsExportedByPaymentIDs(ctx context.Context, db database.QueryExecer, paymentIDs []string, isExported bool) error {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.UpdateIsExportedByPaymentIDs")
	defer span.End()

	query := `UPDATE payment SET is_exported = $1 WHERE payment_id = ANY($2)`

	cmdTag, err := db.Exec(ctx, query, isExported, paymentIDs)
	if err != nil {
		return fmt.Errorf("err UpdateIsExportedByPaymentIDs PaymentRepo: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("err UpdateIsExportedByPaymentIDs PaymentRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *PaymentRepo) FindAllByBulkPaymentID(ctx context.Context, db database.QueryExecer, bulkPaymentID string) ([]*entities.Payment, error) {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.FindByBulkPaymentID")
	defer span.End()

	e := &entities.Payment{}
	fields, _ := e.FieldMap()

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE bulk_payment_id = $1", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(ctx, stmt, &bulkPaymentID)
	if err != nil {
		return nil, fmt.Errorf("err InvoiceRepo.RetrievedMigratedInvoices: %w", err)
	}

	defer rows.Close()
	var payments []*entities.Payment
	for rows.Next() {
		payment := new(entities.Payment)
		_, values := payment.FieldMap()

		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func (r *PaymentRepo) UpdateStatusAndAmountByPaymentIDs(ctx context.Context, db database.QueryExecer, paymentIDs []string, status string, amount float64) error {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.UpdateStatusAndAmountByPaymentIDs")
	defer span.End()

	query := `UPDATE payment SET payment_status = $1, amount = $2 WHERE payment_id = ANY($3)`

	cmdTag, err := db.Exec(ctx, query, status, amount, paymentIDs)
	if err != nil {
		return fmt.Errorf("err UpdateStatusAndAmountByPaymentIDs PaymentRepo: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("err UpdateStatusAndAmountByPaymentIDs PaymentRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *PaymentRepo) CountOtherPaymentsByBulkPaymentIDNotInStatus(ctx context.Context, db database.QueryExecer, bulkPaymentID, paymentID, paymentStatus string) (int, error) {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.FindOtherPaymentsByBulkPaymentIDNotInStatus")
	defer span.End()

	e := &entities.Payment{}

	var count int
	stmt := fmt.Sprintf("SELECT count(payment_id) AS total_count FROM %s WHERE bulk_payment_id = $1 AND payment_id != $2 AND payment_status != $3", e.TableName())

	if err := db.QueryRow(ctx, stmt, &bulkPaymentID, &paymentID, &paymentStatus).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *PaymentRepo) FindByPaymentIDs(ctx context.Context, db database.QueryExecer, paymentIds []string) ([]*entities.Payment, error) {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.FindByPaymentIDs")
	defer span.End()

	e := &entities.Payment{}
	fields, _ := e.FieldMap()

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE payment_id = ANY($1)", strings.Join(fields, ","), e.TableName())

	var arr pgtype.TextArray
	_ = arr.Set(paymentIds)

	rows, err := db.Query(ctx, stmt, arr)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var payments []*entities.Payment
	for rows.Next() {
		payment := new(entities.Payment)
		_, values := payment.FieldMap()

		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, nil
}

func (r *PaymentRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, payments []*entities.Payment) error {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, payment *entities.Payment) {
		fields := database.GetFieldNames(payment)
		fields = utils.RemoveStrFromSlice(fields, "resource_path")

		// Remove the payment sequence number on insert statement if it is empty
		if payment.PaymentSequenceNumber.Status == pgtype.Null || payment.PaymentSequenceNumber.Int == 0 {
			fields = utils.RemoveStrFromSlice(fields, "payment_sequence_number")
		}

		values := database.GetScanFields(payment, fields)

		placeHolders := database.GeneratePlaceholders(len(fields))

		stmt :=
			`
			INSERT INTO %s (%s) VALUES (%s);
			`

		stmt = fmt.Sprintf(stmt, payment.TableName(), strings.Join(fields, ","), placeHolders)
		b.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}

	now := time.Now().UTC()
	for _, payment := range payments {
		errs := []error{payment.UpdatedAt.Set(now), payment.CreatedAt.Set(now)}
		if strings.TrimSpace(payment.PaymentID.String) == "" {
			errs = append(errs, payment.PaymentID.Set(idutil.ULIDNow()))
		}

		err := multierr.Combine(errs...)
		if err != nil {
			return err
		}

		queueFn(batch, payment)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for i := 0; i < len(payments); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("no rows affected when creating payment")
		}
	}

	return nil
}

func (r *PaymentRepo) PaymentSeqNumberLockAdvisory(ctx context.Context, db database.QueryExecer) (bool, error) {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.PaymentSeqNumberLockAdvisory")
	defer span.End()

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	lockID := fmt.Sprintf("%s-%s", "payment-sequence-number", resourcePath)

	query := "SELECT pg_try_advisory_lock(hashtext($1))"
	var lockAcquired bool

	err := db.QueryRow(ctx, query, lockID).Scan(&lockAcquired)
	if err != nil {
		return false, fmt.Errorf("err PaymentSeqNumberLockAdvisory PaymentRepo: %w - resourcePath: %s", err, resourcePath)
	}

	return lockAcquired, nil
}

func (r *PaymentRepo) PaymentSeqNumberUnLockAdvisory(ctx context.Context, db database.QueryExecer) error {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.PaymentSeqNumberUnLockAdvisory")
	defer span.End()

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	lockID := fmt.Sprintf("%s-%s", "payment-sequence-number", resourcePath)

	query := "SELECT pg_advisory_unlock(hashtext($1))"

	_, err := db.Exec(ctx, query, lockID)
	if err != nil {
		return fmt.Errorf("err PaymentSeqNumberLockAdvisory PaymentRepo: %w - resourcePath: %s", err, resourcePath)
	}

	return nil
}

func (r *PaymentRepo) GetLatestPaymentSequenceNumber(ctx context.Context, db database.QueryExecer) (int32, error) {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.GetLatestPaymentSequenceNumber")
	defer span.End()

	resourcePath := golibs.ResourcePathFromCtx(ctx)

	var paymentSequenceNumber int32
	stmt := "SELECT payment_sequence_number FROM payment WHERE resource_path = $1 AND payment_sequence_number IS NOT null ORDER BY payment_sequence_number DESC LIMIT 1"
	if err := db.QueryRow(ctx, stmt, resourcePath).Scan(&paymentSequenceNumber); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return paymentSequenceNumber, nil
}

func (r *PaymentRepo) UpdateMultipleWithFields(ctx context.Context, db database.QueryExecer, payments []*entities.Payment, fields []string) error {
	ctx, span := interceptors.StartSpan(ctx, "PaymentRepo.UpdateMultipleWithFields")
	defer span.End()

	primaryField := "payment_id"
	newValueAlias := "nv"
	payment := &entities.Payment{}

	// Include the primary field in argument as the first field per row
	fieldWithPrimary := []string{primaryField}
	fieldWithPrimary = append(fieldWithPrimary, fields...)

	// This is the statement for SET clause e.g. (col1 = nv.col1, col2 = nv.col2)
	setStmt := generateBatchUpdateSetStmt(fields, newValueAlias)

	// This is the statement for VALUES clause e.g. ($1, $2, $3::timestamptz), ($4, $5, $6::timestamptz)
	valuesStmt := generateBatchUpdateValuesStmt(fieldWithPrimary, len(payments), payment) // primary field should be included

	// This is the statement for ALIAS clause e.g. nv(primary_field, col1, col2)
	aliasStmt := generateBatchUpdateAliasStmt(fields, primaryField, newValueAlias)
	stmt := fmt.Sprintf(`
			UPDATE %s p 
			SET %s 
			FROM (
				VALUES 
					%s
			) AS %s 
			WHERE p.%s = %s.%s
		`, payment.TableName(), setStmt, valuesStmt, aliasStmt, primaryField, newValueAlias, primaryField,
	)

	args := []interface{}{}
	for _, payment := range payments {
		args = append(args, database.GetScanFields(payment, fieldWithPrimary)...)
	}

	_, err := db.Exec(ctx, stmt, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *PaymentRepo) InsertPaymentNumbersTempTable(ctx context.Context, db database.QueryExecer, paymentSeqNumbers []int) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.InsertPaymentNumbersTempTable")
	defer span.End()

	// Create temporary table to store the invoice IDs
	createTempTableQuery := `CREATE TEMPORARY TABLE temp_tbl__payment_number(temp_payment_sequence_number INT PRIMARY KEY, resource_path TEXT) ON COMMIT DROP`
	_, err := db.Exec(ctx, createTempTableQuery)
	if err != nil {
		return fmt.Errorf("err creating temporary table: %w", err)
	}

	resourcePath := golibs.ResourcePathFromCtx(ctx)

	// Batch insert the invoice IDs to temporary table
	queueFn := func(b *pgx.Batch, paymentNo int) {
		stmt :=
			`
			INSERT INTO temp_tbl__payment_number(temp_payment_sequence_number, resource_path) VALUES ($1, $2)
			ON CONFLICT
			DO NOTHING;
			`
		b.Queue(stmt, paymentNo, resourcePath)
	}

	batch := &pgx.Batch{}
	for _, n := range paymentSeqNumbers {
		queueFn(batch, n)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for i := 0; i < len(paymentSeqNumbers); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("no rows affected when creating payment temp values")
		}
	}

	return nil
}

func (r *PaymentRepo) FindPaymentInvoiceUserFromTempTable(ctx context.Context, db database.QueryExecer) ([]*entities.PaymentInvoiceUserMap, error) {
	_, span := interceptors.StartSpan(ctx, "PaymentRepo.FindPaymentInvoiceUserFromTempTable")
	defer span.End()

	query := `
			SELECT %s
			FROM payment p
			INNER JOIN invoice i
				ON p.invoice_id = i.invoice_id
			INNER JOIN user_basic_info u
				ON u.user_id = p.student_id
			WHERE EXISTS (
				SELECT 1
				FROM temp_tbl__payment_number tp
				WHERE tp.temp_payment_sequence_number = p.payment_sequence_number AND resource_path = $1
			)
	`

	fields := []string{}
	payment := &entities.Payment{}
	invoice := &entities.Invoice{}
	userBasicInfo := &entities.UserBasicInfo{}

	paymentFields, _ := payment.FieldMap()
	invoiceFields, _ := invoice.FieldMap()
	userBasicInfoFields, _ := userBasicInfo.FieldMap()

	for _, f := range paymentFields {
		fields = append(fields, fmt.Sprintf("p.%s", f))
	}

	for _, f := range invoiceFields {
		fields = append(fields, fmt.Sprintf("i.%s", f))
	}

	for _, f := range userBasicInfoFields {
		fields = append(fields, fmt.Sprintf("u.%s", f))
	}

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	rows, err := db.Query(ctx, fmt.Sprintf(query, strings.Join(fields, ",")), resourcePath)
	if err != nil {
		return nil, err
	}

	results := []*entities.PaymentInvoiceUserMap{}
	defer rows.Close()
	for rows.Next() {
		e := &entities.PaymentInvoiceUserMap{
			Payment:       &entities.Payment{},
			Invoice:       &entities.Invoice{},
			UserBasicInfo: &entities.UserBasicInfo{},
		}

		_, paymentValues := e.Payment.FieldMap()
		_, invoiceValues := e.Invoice.FieldMap()
		_, userValues := e.UserBasicInfo.FieldMap()

		args := []interface{}{}
		args = append(args, paymentValues...)
		args = append(args, invoiceValues...)
		args = append(args, userValues...)

		err := rows.Scan(args...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		results = append(results, e)
	}

	return results, nil
}

func generateBatchUpdateSetStmt(fields []string, newValueAlias string) string {
	var builder strings.Builder
	sep := ", "

	totalField := len(fields)
	for i, field := range fields {
		if i == totalField-1 {
			sep = ""
		}

		builder.WriteString(fmt.Sprintf("%s = %s.%s%s", field, newValueAlias, field, sep))
	}

	return builder.String()
}

func generateBatchUpdateValuesStmt(fields []string, rowLength int, e database.Entity) string {
	entityField, values := e.FieldMap()

	// map the field to entity value
	fieldValueMap := make(map[string]interface{})
	for i, f := range entityField {
		fieldValueMap[f] = values[i]
	}

	// generate the place holders per row
	placeHolders := []string{}
	for i := 0; i < rowLength; i++ {
		newPlaceHolders := []string{}
		for j, field := range fields {
			var holder string

			// To support other type to be cast in pg, add case here
			switch fieldValueMap[field].(type) {
			case *pgtype.Timestamptz: // cast the type Timestamptz in timestamptz in pg
				holder = "$%d::timestamptz"
			case *pgtype.Numeric: // cast the type Numeric in numeric in pg
				holder = "$%d::numeric"
			default:
				holder = "$%d"
			}

			newPlaceHolders = append(newPlaceHolders, fmt.Sprintf(holder, (j+1)+(i*len(fields))))
		}
		placeHolders = append(placeHolders, fmt.Sprintf("(%s)", strings.Join(newPlaceHolders, ", ")))
	}

	return strings.Join(placeHolders, ",")
}

func generateBatchUpdateAliasStmt(fields []string, primaryField string, newValueAlias string) string {
	var builder strings.Builder

	builder.WriteString(newValueAlias + "(")

	newFields := []string{primaryField}
	newFields = append(newFields, fields...)

	builder.WriteString(strings.Join(newFields, ", "))

	builder.WriteString(")")

	return builder.String()
}
