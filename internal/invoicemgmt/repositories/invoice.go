package repositories

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type InvoiceRepo struct {
}

func (r *InvoiceRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.Invoice) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "invoice_id", []string{"status", "updated_at"})

	if err != nil {
		return fmt.Errorf("err update InvoiceRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update InvoiceRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
func (r *InvoiceRepo) RetrieveRecordsByStudentID(ctx context.Context, db database.QueryExecer, studentID string, limit, offset pgtype.Int8) ([]*entities.Invoice, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.RetrieveRecordsByStudentID")
	defer span.End()

	e := &entities.Invoice{}
	query := fmt.Sprintf("SELECT invoice_id, status, total FROM %s WHERE student_id = $1 AND status != 'DRAFT' ORDER BY created_at DESC LIMIT $2 OFFSET $3", e.TableName())
	rows, err := db.Query(ctx, query, &studentID, &limit, &offset)
	if err != nil {
		return nil, fmt.Errorf("err retrieve records InvoiceRepo: %w", err)
	}
	defer rows.Close()
	var result []*entities.Invoice
	for rows.Next() {
		var invoiceRecord entities.Invoice
		if err := rows.Scan(&invoiceRecord.InvoiceID, &invoiceRecord.Status, &invoiceRecord.Total); err != nil {
			return nil, err
		}
		result = append(result, &invoiceRecord)
	}
	return result, nil
}

func (r *InvoiceRepo) RetrieveInvoiceByInvoiceID(ctx context.Context, db database.QueryExecer, invoiceID string) (*entities.Invoice, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.RetrieveInvoiceByInvoiceID")
	defer span.End()

	e := &entities.Invoice{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_id = $1", strings.Join(fields, ","), e.TableName())

	err := database.Select(ctx, db, query, &invoiceID).ScanOne(e)

	if err != nil {
		return nil, err
	}

	return e, nil
}

// Create invoice
func (r *InvoiceRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.Invoice) (invoiceID pgtype.Text, err error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.Create")
	defer span.End()
	now := time.Now()

	if e.CreatedAt.Time.IsZero() {
		err := e.CreatedAt.Set(now)
		if err != nil {
			return invoiceID, err
		}
	}

	if err = multierr.Combine(
		e.InvoiceID.Set(idutil.ULIDNow()),
		e.UpdatedAt.Set(now),
	); err != nil {
		err = fmt.Errorf("multierr.Combine UpdatedAt.Set InvoiceID.Set: %w", err)
		return
	}

	err = database.InsertReturningAndExcept(ctx, e, db, []string{"invoice_sequence_number", "resource_path"}, "invoice_id", &invoiceID)
	if err != nil {
		err = fmt.Errorf("err insert Invoice: %w", err)
		return
	}

	return
}

func (r *InvoiceRepo) UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.Invoice, fieldsToUpdate []string) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.UpdateWithFields")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "invoice_id", fieldsToUpdate)

	if err != nil {
		return fmt.Errorf("err updateWithFields InvoiceRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err updateWithFields InvoiceRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *InvoiceRepo) UpdateIsExportedByPaymentRequestFileID(ctx context.Context, db database.QueryExecer, fileID string, isExported bool) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.UpdateIsExportedByPaymentRequestFileID")
	defer span.End()

	query := `
		UPDATE invoice
		SET is_exported = $2
		WHERE invoice_id IN (
			SELECT p.invoice_id
			FROM bulk_payment_request_file_payment fp
			INNER JOIN payment p
				ON p.payment_id = fp.payment_id
			WHERE bulk_payment_request_file_id = $1
		)
	`

	cmdTag, err := db.Exec(ctx, query, fileID, isExported)
	if err != nil {
		return fmt.Errorf("err UpdateIsExportedByPaymentRequestFileID InvoiceRepo: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("err UpdateIsExportedByPaymentRequestFileID InvoiceRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *InvoiceRepo) UpdateIsExportedByInvoiceIDs(ctx context.Context, db database.QueryExecer, invoiceIDs []string, isExported bool) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.UpdateIsExportedByPaymentIDs")
	defer span.End()

	query := `UPDATE invoice SET is_exported = $1 WHERE invoice_id = ANY($2)`

	cmdTag, err := db.Exec(ctx, query, isExported, invoiceIDs)
	if err != nil {
		return fmt.Errorf("err UpdateIsExportedByPaymentIDs InvoiceRepo: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("err UpdateIsExportedByPaymentIDs InvoiceRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

// used in data migration
func (r *InvoiceRepo) RetrieveInvoiceByInvoiceReferenceID(ctx context.Context, db database.QueryExecer, invoiceReferenceID string) (*entities.Invoice, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.RetrieveInvoiceByInvoiceReferenceID")
	defer span.End()

	e := &entities.Invoice{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_reference_id = $1 AND resource_path = $2", strings.Join(fields, ","), e.TableName())

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	err := database.Select(ctx, db, query, &invoiceReferenceID, &resourcePath).ScanOne(e)
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (r *InvoiceRepo) RetrievedMigratedInvoices(ctx context.Context, db database.QueryExecer) ([]*entities.Invoice, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.RetrievedMigratedInvoices")
	defer span.End()

	e := &entities.Invoice{}
	fields, _ := e.FieldMap()

	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_reference_id IS NOT NULL AND migrated_at IS NOT NULL AND resource_path = $1", strings.Join(fields, ","), e.TableName())

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	rows, err := db.Query(ctx, stmt, &resourcePath)
	if err != nil {
		return nil, fmt.Errorf("err InvoiceRepo.RetrievedMigratedInvoices: %w", err)
	}

	defer rows.Close()
	var invoices []*entities.Invoice
	for rows.Next() {
		invoice := new(entities.Invoice)
		_, values := invoice.FieldMap()

		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *InvoiceRepo) InsertInvoiceIDsTempTable(ctx context.Context, db database.QueryExecer, invoiceIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.InsertInvoiceIDsTempTable")
	defer span.End()

	// Create temporary table to store the invoice IDs
	createTempTableQuery := `CREATE TEMPORARY TABLE temporary_invoice_id_table(temp_invoice_id TEXT PRIMARY KEY) ON COMMIT DROP`
	_, err := db.Exec(ctx, createTempTableQuery)
	if err != nil {
		return fmt.Errorf("err creating temporary table: %w", err)
	}

	// Batch insert the invoice IDs to temporary table
	queueFn := func(b *pgx.Batch, invoiceID string) {
		stmt :=
			`
			INSERT INTO temporary_invoice_id_table(temp_invoice_id) VALUES ($1)
			ON CONFLICT
			DO NOTHING;
			`
		b.Queue(stmt, invoiceID)
	}

	batch := &pgx.Batch{}
	for _, invoiceID := range invoiceIDs {
		queueFn(batch, invoiceID)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for i := 0; i < len(invoiceIDs); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("no rows affected when creating action log")
		}
	}

	return nil
}

func (r *InvoiceRepo) FindInvoicesFromInvoiceIDTempTable(ctx context.Context, db database.QueryExecer) ([]*entities.Invoice, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.FindInvoicesFromInvoiceIDTempTable")
	defer span.End()

	e := &entities.Invoice{}
	fields, _ := e.FieldMap()

	stmt := fmt.Sprintf(`
		SELECT %s 
		FROM invoice i
		WHERE EXISTS (
			SELECT 1
			FROM temporary_invoice_id_table ti
			WHERE ti.temp_invoice_id = i.invoice_id
		)
	`, strings.Join(fields, ","))

	rows, err := db.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var invoices []*entities.Invoice
	for rows.Next() {
		invoice := new(entities.Invoice)
		_, values := invoice.FieldMap()

		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *InvoiceRepo) UpdateStatusFromInvoiceIDTempTable(ctx context.Context, db database.QueryExecer, status string) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.UpdateStatusFromInvoiceIDTempTable")
	defer span.End()

	query := `
			UPDATE invoice i 
			SET status = $1, updated_at = now()
			WHERE EXISTS (
				SELECT 1
				FROM temporary_invoice_id_table ti
				WHERE ti.temp_invoice_id = i.invoice_id
			)
		`

	cmdTag, err := db.Exec(ctx, query, status)
	if err != nil {
		return fmt.Errorf("err UpdateStatusFromInvoiceIDTempTable InvoiceRepo: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("err UpdateStatusFromInvoiceIDTempTable InvoiceRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *InvoiceRepo) UpdateMultipleWithFields(ctx context.Context, db database.QueryExecer, invoices []*entities.Invoice, fields []string) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.UpdateMultipleWithFields")
	defer span.End()

	primaryField := "invoice_id"
	newValueAlias := "nv"
	invoice := &entities.Invoice{}

	// Include the primary field in argument as the first field per row
	fieldWithPrimary := []string{primaryField}
	fieldWithPrimary = append(fieldWithPrimary, fields...)

	// This is the statement for SET clause e.g. (col1 = nv.col1, col2 = nv.col2)
	setStmt := generateBatchUpdateSetStmt(fields, newValueAlias)

	// This is the statement for VALUES clause e.g. ($1, $2, $3::timestamptz), ($4, $5, $6::timestamptz)
	valuesStmt := generateBatchUpdateValuesStmt(fieldWithPrimary, len(invoices), invoice) // primary field should be included

	// This is the statement for ALIAS clause e.g. nv(primary_field, col1, col2)
	aliasStmt := generateBatchUpdateAliasStmt(fields, primaryField, newValueAlias)
	stmt := fmt.Sprintf(`
			UPDATE %s i
			SET %s 
			FROM (
				VALUES 
					%s
			) AS %s 
			WHERE i.%s = %s.%s
		`, invoice.TableName(), setStmt, valuesStmt, aliasStmt, primaryField, newValueAlias, primaryField,
	)

	args := []interface{}{}
	for _, invoice := range invoices {
		args = append(args, database.GetScanFields(invoice, fieldWithPrimary)...)
	}

	_, err := db.Exec(ctx, stmt, args...)
	if err != nil {
		return err
	}
	return nil
}

func (r *InvoiceRepo) RetrieveInvoiceData(ctx context.Context, db database.QueryExecer, limit, offset pgtype.Int8, sqlFilter string) ([]*entities.InvoicePaymentMap, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.RetrieveInvoiceData")
	defer span.End()

	query := fmt.Sprintf(`
	WITH LatestPayment AS (
		SELECT
			p.invoice_id,
			p.payment_date,
			p.payment_due_date,
			p.payment_expiry_date,
			p.payment_method,
			p.payment_status,
			p.amount,
			p.is_exported,
			p.payment_id,
			p.payment_sequence_number,
			ROW_NUMBER() OVER (PARTITION BY p.invoice_id ORDER BY p.created_at DESC) AS rn
		FROM payment p
	)
	SELECT
		i.invoice_id,
		i.invoice_sequence_number,
		i.status,
		i.student_id,
		i.sub_total,
		i.total,
		i.outstanding_balance,
		i.amount_paid,
		i.type,
		i.created_at,
		i.deleted_at,
		p.payment_date,
		p.payment_due_date,
		p.payment_expiry_date,
		p.payment_method,
		p.payment_status,
		p.amount,
		p.is_exported,
		p.payment_id,
		p.payment_sequence_number,
		u.name
	FROM (
		SELECT
			invoice_id,
			invoice_sequence_number,
			status,
			student_id,
			sub_total,
			total,
			outstanding_balance,
			amount_paid,
			type,
			created_at,
			resource_path,
			deleted_at
		FROM invoice
	) AS i
	LEFT JOIN LatestPayment p ON i.invoice_id = p.invoice_id AND p.rn = 1
	INNER JOIN user_basic_info AS u ON i.student_id = u.user_id
	WHERE %s
	ORDER BY i.created_at DESC
	LIMIT $1
	OFFSET $2;
	`, sqlFilter)

	rows, err := db.Query(ctx, query, &limit, &offset)
	if err != nil {
		return nil, fmt.Errorf("err InvoiceRepo RetrieveInvoiceData: %w", err)
	}

	defer rows.Close()
	var result []*entities.InvoicePaymentMap
	for rows.Next() {
		e := &entities.InvoicePaymentMap{
			Invoice:       &entities.Invoice{},
			Payment:       &entities.Payment{},
			UserBasicInfo: &entities.UserBasicInfo{},
		}

		err := rows.Scan(
			&e.Invoice.InvoiceID,
			&e.Invoice.InvoiceSequenceNumber,
			&e.Invoice.Status,
			&e.Invoice.StudentID,
			&e.Invoice.SubTotal,
			&e.Invoice.Total,
			&e.Invoice.OutstandingBalance,
			&e.Invoice.AmountPaid,
			&e.Invoice.Type,
			&e.Invoice.CreatedAt,
			&e.Invoice.DeletedAt,
			&e.Payment.PaymentDate,
			&e.Payment.PaymentDueDate,
			&e.Payment.PaymentExpiryDate,
			&e.Payment.PaymentMethod,
			&e.Payment.PaymentStatus,
			&e.Payment.Amount,
			&e.Payment.IsExported,
			&e.Payment.PaymentID,
			&e.Payment.PaymentSequenceNumber,
			&e.UserBasicInfo.Name,
		)

		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		result = append(result, e)
	}
	return result, nil
}

func GenerateInvoiceDataWhereClause(invoiceFilter *invoice_pb.InvoiceDataForInvoiceFilter, paymentFilter *invoice_pb.InvoiceDataForPaymentFilter, studentName string) (string, error) {
	dateStrFormat := "2006-01-02"
	sqlFilter := "i.deleted_at IS NULL AND i.status != 'FAILED'"
	// for invoice filter
	if invoiceFilter != nil {
		if invoiceFilter.InvoiceTypes != nil || len(invoiceFilter.InvoiceTypes) > 0 {
			var invoiceTypes string
			for i, typeStr := range invoiceFilter.InvoiceTypes {
				if i > 0 {
					invoiceTypes += ", "
				}
				invoiceTypes += "'" + typeStr.String() + "'"
			}
			sqlFilter += fmt.Sprintf(" AND i.type IN (%s)", invoiceTypes)
		}

		if strings.TrimSpace(invoiceFilter.MinAmount) != "" {
			minTotalFloat, err := strconv.ParseFloat(strings.TrimSpace(invoiceFilter.MinAmount), 64)
			if err != nil {
				return "", err
			}

			sqlFilter += fmt.Sprintf(" AND i.total >= %d", int(minTotalFloat))
		}

		if strings.TrimSpace(invoiceFilter.MaxAmount) != "" {
			maxTotalFloat, err := strconv.ParseFloat(strings.TrimSpace(invoiceFilter.MaxAmount), 64)
			if err != nil {
				return "", err
			}

			sqlFilter += fmt.Sprintf(" AND i.total <= %d", int(maxTotalFloat))
		}

		if invoiceFilter.CreatedDateFrom != nil {
			sqlFilter += fmt.Sprintf(" AND DATE(i.created_at) >= '%s'", invoiceFilter.CreatedDateFrom.AsTime().Format(dateStrFormat))
		}

		if invoiceFilter.CreatedDateUntil != nil {
			sqlFilter += fmt.Sprintf(" AND DATE(i.created_at) <= '%s'", invoiceFilter.CreatedDateUntil.AsTime().Format(dateStrFormat))
		}

		if strings.TrimSpace(invoiceFilter.InvoiceStatus.String()) != "" && strings.TrimSpace(invoiceFilter.InvoiceStatus.String()) != invoice_pb.InvoiceStatus_ALL_STATUS.String() {
			sqlFilter += fmt.Sprintf(" AND i.status = '%s'", invoiceFilter.InvoiceStatus.String())
		}
	}
	// for payment filter
	if paymentFilter != nil {
		if paymentFilter.PaymentMethods != nil || len(paymentFilter.PaymentMethods) > 0 {
			var paymentMethods string
			for i, paymentMethodStr := range paymentFilter.PaymentMethods {
				if i > 0 {
					paymentMethods += ", "
				}
				paymentMethods += "'" + paymentMethodStr.String() + "'"
			}
			sqlFilter += fmt.Sprintf(" AND p.payment_method IN (%s)", paymentMethods)
		}

		if paymentFilter.DueDateFrom != nil {
			sqlFilter += fmt.Sprintf(" AND DATE(p.payment_due_date) >= '%s'", paymentFilter.DueDateFrom.AsTime().Format(dateStrFormat))
		}

		if paymentFilter.DueDateUntil != nil {
			sqlFilter += fmt.Sprintf(" AND DATE(p.payment_due_date) <= '%s'", paymentFilter.DueDateUntil.AsTime().Format(dateStrFormat))
		}

		if paymentFilter.ExpiryDateFrom != nil {
			sqlFilter += fmt.Sprintf(" AND DATE(p.payment_expiry_date) >= '%s'", paymentFilter.ExpiryDateFrom.AsTime().Format(dateStrFormat))
		}

		if paymentFilter.ExpiryDateUntil != nil {
			sqlFilter += fmt.Sprintf(" AND DATE(p.payment_expiry_date) <= '%s'", paymentFilter.ExpiryDateUntil.AsTime().Format(dateStrFormat))
		}

		if paymentFilter.PaymentStatuses != nil || len(paymentFilter.PaymentStatuses) > 0 {
			var paymentStatuses string
			for i, paymentStatusStr := range paymentFilter.PaymentStatuses {
				if i > 0 {
					paymentStatuses += ", "
				}
				paymentStatuses += "'" + paymentStatusStr.String() + "'"
			}
			sqlFilter += fmt.Sprintf(" AND p.payment_status IN (%s)", paymentStatuses)
		}
		// if this is set to true, no need to filter just retrieve all payment is exported status
		if !paymentFilter.IsExported {
			sqlFilter += " AND p.is_exported = false"
		}
	}

	if strings.TrimSpace(studentName) != "" {
		sqlFilter += fmt.Sprintf(" AND u.name ILIKE '%%%s%%'", strings.TrimSpace(studentName))
	}
	return sqlFilter, nil
}

func (r *InvoiceRepo) RetrieveInvoiceStatusCount(ctx context.Context, db database.QueryExecer, sqlFilter string) (map[string]int32, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceRepo.RetrieveInvoiceStatusCount")
	defer span.End()

	query := fmt.Sprintf(`
	SELECT
		i.status,
		COUNT(*) AS status_count
	FROM invoice i
	LEFT JOIN (
		SELECT
			ROW_NUMBER() OVER (PARTITION BY invoice_id ORDER BY created_at::timestamptz DESC) AS row_num,
			*
		FROM payment
	) p
		ON p.invoice_id = i.invoice_id AND p.row_num = 1
	INNER JOIN user_basic_info u
		ON u.user_id = i.student_id
	WHERE %s
	GROUP BY i.status;
	`, sqlFilter)

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("err InvoiceRepo RetrieveInvoiceStatusCount: %w", err)
	}

	statusCountMap := make(map[string]int32)

	for rows.Next() {
		var (
			status      string
			countStatus int32
		)
		if err := rows.Scan(&status, &countStatus); err != nil {
			return nil, fmt.Errorf("rows.Scan: %v", err)
		}
		statusCountMap[status] = countStatus
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %v", err)
	}

	return statusCountMap, nil
}
