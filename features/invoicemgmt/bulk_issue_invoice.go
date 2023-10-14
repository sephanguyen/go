package invoicemgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) issuesInvoicesInBulkWithPaymentMethodAnd(ctx context.Context, signedInUser, status, dueDate, expiryDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, signedInUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	dueDateTimestamp := getFormattedTimestampDate(dueDate)
	expiryDateTimestamp := getFormattedTimestampDate(expiryDate)

	var paymentMethod invoice_pb.BulkIssuePaymentMethod
	switch status {
	case "INVALID_PAYMENT_METHOD":
		paymentMethod = 99
	default:
		paymentMethod = invoice_pb.BulkIssuePaymentMethod(invoice_pb.BulkIssuePaymentMethod_value[status])
	}

	req := &invoice_pb.BulkIssueInvoiceRequest{
		InvoiceIds:             s.StepState.InvoiceIDs,
		BulkIssuePaymentMethod: paymentMethod,
		ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
			DueDate:    dueDateTimestamp,
			ExpiryDate: expiryDateTimestamp,
		},
		DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueDirectDebitDates{
			DueDate:    nil,
			ExpiryDate: nil,
		},
	}

	if status == "BULK_ISSUE_DEFAULT_PAYMENT" {
		req.DirectDebitDates.DueDate = dueDateTimestamp
		req.DirectDebitDates.ExpiryDate = expiryDateTimestamp
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).BulkIssueInvoice(contextWithToken(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoicesStatusIsUpdatedToStatus(ctx context.Context, newStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, invoiceID := range stepState.InvoiceIDs {
		var invoiceStatus string
		row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, "SELECT status FROM invoice WHERE invoice_id = $1", invoiceID)

		if err := row.Scan(&invoiceStatus); err != nil {
			return ctx, fmt.Errorf("error finding invoice with invoice_id: %v: %w", invoiceID, err)
		}

		if invoiceStatus != newStatus {
			return ctx, fmt.Errorf("invoice with invoice_id %v and status %v should have status of %v", invoiceID, invoiceStatus, newStatus)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) oneInvoiceIDIsAddedToTheRequestButIsNonexisting(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Overwrites the existing InvoiceID to simulate a non-existing invoice
	stepState.InvoiceIDs = append(stepState.InvoiceIDs, idutil.ULIDNow())

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) oneInvoiceHasNegativeTotalAmount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.createStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	var negativeTotalInvoice string
	stmt := `INSERT INTO invoice(
		invoice_id,
		type, 
		status, 
		student_id, 
		sub_total, 
		total,
		outstanding_balance,
		amount_paid,
		amount_refunded,
		created_at, 
		updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now()) RETURNING invoice_id`

	if err := try.Do(func(attempt int) (bool, error) {
		row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt,
			idutil.ULIDNow(),
			invoice_pb.PaymentMethod_DIRECT_DEBIT.String(),
			invoice_pb.InvoiceStatus_DRAFT.String(),
			stepState.StudentID,
			-100,
			-100,
			-100,
			0,
			0,
		)
		err := row.Scan(&negativeTotalInvoice)

		if err == nil {
			return false, nil
		}

		if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			return false, fmt.Errorf("repo.Create: %w", err)
		}

		time.Sleep(invoiceConst.DuplicateSleepDuration)
		return attempt < 10, fmt.Errorf("cannot create negative invoice, err %v", err)
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.InvoiceIDs = append(stepState.InvoiceIDs, negativeTotalInvoice)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) issuesInvoicesInBulkWithPaymentMethodAndDueDateAfterExpiryDate(ctx context.Context, signedInUser string, bulkPaymentMethod, defaultPaymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, signedInUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	paymentMethod := invoice_pb.BulkIssuePaymentMethod(invoice_pb.BulkIssuePaymentMethod_value[bulkPaymentMethod])

	req := &invoice_pb.BulkIssueInvoiceRequest{
		InvoiceIds:             s.StepState.InvoiceIDs,
		BulkIssuePaymentMethod: paymentMethod,
		ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueConvenieceStoreDates{
			DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
		},
		DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequest_BulkIssueDirectDebitDates{
			DueDate:    timestamppb.New(time.Now().Add(1 * time.Hour)),
			ExpiryDate: timestamppb.New(time.Now().Add(2 * time.Hour)),
		},
	}

	switch defaultPaymentMethod {
	case "DIRECT_DEBIT":
		req.DirectDebitDates.DueDate = timestamppb.New(time.Now().Add(2 * time.Hour))
		req.DirectDebitDates.ExpiryDate = timestamppb.New(time.Now().Add(1 * time.Hour))
	default:
		req.ConvenienceStoreDates.DueDate = timestamppb.New(time.Now().Add(2 * time.Hour))
		req.ConvenienceStoreDates.ExpiryDate = timestamppb.New(time.Now().Add(1 * time.Hour))
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).BulkIssueInvoice(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) actionLogRecordForEachInvoiceIsRecordedWithActionLogType(ctx context.Context, expectedActionLogType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, invoiceID := range stepState.InvoiceIDs {
		actionLog, err := s.getLatestInvoiceActionLogByInvoiceID(ctx, invoiceID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("actionLogRecordIsRecordedWithActionLogType error: %v", err.Error())
		}
		if expectedActionLogType != actionLog.Action.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected %v action log type but got %v for invoice_id %v", expectedActionLogType, actionLog.Action.String, s.StepState.InvoiceID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theseInvoiceForStudentsHaveDefaultPaymentMethod(ctx context.Context, defaultPaymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if defaultPaymentMethod != "" {
		for _, studentID := range stepState.StudentIds {
			ctx, err := s.createStudentPaymentDetail(ctx, defaultPaymentMethod, studentID)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereArePendingPaymentRecordsForStudentsCreatedWithPaymentMethodAnd(ctx context.Context, paymentMethod, dueDate, expiryDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	dueDateTime := getFormattedTimestampDate(dueDate).AsTime()
	expiryDateTime := getFormattedTimestampDate(expiryDate).AsTime()
	dueDateTimeStr := dueDateTime.Format("2006-01-02")
	expiryDateTimeStr := expiryDateTime.Format("2006-01-02")

	for _, invoiceID := range stepState.InvoiceIDs {
		var paymentRecordCnt int32

		stmt := fmt.Sprintf("SELECT COUNT(*) FROM payment WHERE invoice_id = '%v' AND payment_status = '%v' AND to_char(payment_due_date, 'YYYY-MM-DD') = '%v' AND to_char(payment_expiry_date, 'YYYY-MM-DD') = '%v'", invoiceID, invoice_pb.PaymentStatus_PAYMENT_PENDING.String(), dueDateTimeStr, expiryDateTimeStr)

		row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt)

		if err := row.Scan(&paymentRecordCnt); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error finding payment with invoice_id '%v' and status %v: %w", invoiceID, invoice_pb.PaymentStatus_PAYMENT_PENDING.String(), err)
		}
		if paymentRecordCnt == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("no payment record with invoice_id %v exists", s.StepState.InvoiceID)
		}
		if paymentRecordCnt > 1 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected 1 payment record inserted but got %v for invoice_id %v", paymentRecordCnt, s.StepState.InvoiceID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoicesExportedTagIsSetTo(ctx context.Context, isExportedStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	isExported := isExportedStr == "true"

	for _, invoiceID := range stepState.InvoiceIDs {
		var actualIsExported bool
		row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, "SELECT is_exported FROM invoice WHERE invoice_id = $1", invoiceID)

		if err := row.Scan(&actualIsExported); err != nil {
			return ctx, fmt.Errorf("error finding invoice with invoice_id: %v: %w", invoiceID, err)
		}

		if actualIsExported != isExported {
			return ctx, fmt.Errorf("invoice with invoice_id %v exported tag should be set to %v", invoiceID, isExported)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) paymentsExportedTagIsSetTo(ctx context.Context, isExportedStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	isExported := isExportedStr == "true"

	for _, invoiceID := range stepState.InvoiceIDs {
		var actualIsExported bool
		row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, "SELECT is_exported FROM payment WHERE invoice_id = $1  ORDER BY created_at DESC LIMIT 1", invoiceID)

		if err := row.Scan(&actualIsExported); err != nil {
			return ctx, fmt.Errorf("error finding payment with invoice_id: %v: %w", invoiceID, err)
		}

		if actualIsExported != isExported {
			return ctx, fmt.Errorf("payment with invoice_id %v exported tag should be set to %v", invoiceID, isExported)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
