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
)

func (s *suite) bulkIssueInvoicesUsingV2EndpointWithPaymentMethod(ctx context.Context, signedInUser, paymentMethodStr, dueDate, expiryDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, signedInUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	dueDateTimestamp := getFormattedTimestampDate(dueDate)
	expiryDateTimestamp := getFormattedTimestampDate(expiryDate)

	paymentMethod := invoice_pb.BulkIssuePaymentMethod(invoice_pb.BulkIssuePaymentMethod_value[paymentMethodStr])

	invoiceTypes := make([]invoice_pb.InvoiceType, 0)

	for _, invoiceType := range stepState.InvoiceTypes {
		invoiceTypes = append(invoiceTypes, invoice_pb.InvoiceType(invoice_pb.InvoiceType_value[invoiceType]))
	}

	req := &invoice_pb.BulkIssueInvoiceRequestV2{
		InvoiceIds:             s.StepState.InvoiceIDs,
		BulkIssuePaymentMethod: paymentMethod,
		ConvenienceStoreDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueConvenienceStoreDates{
			DueDate:    dueDateTimestamp,
			ExpiryDate: expiryDateTimestamp,
		},
		DirectDebitDates: &invoice_pb.BulkIssueInvoiceRequestV2_BulkIssueDirectDebitDates{
			DueDate:    nil,
			ExpiryDate: nil,
		},
		InvoiceType: invoiceTypes,
	}

	if paymentMethod == invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT {
		req.DirectDebitDates.DueDate = dueDateTimestamp
		req.DirectDebitDates.ExpiryDate = expiryDateTimestamp
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).BulkIssueInvoiceV2(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoicesHasPaymentWithStatus(ctx context.Context, existence string, paymentStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, invoiceID := range stepState.InvoiceIDs {
		stepState.InvoiceID = invoiceID
		ctx, err := s.thisInvoiceHasPaymentWithStatus(ctx, existence, paymentStatus)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("InvoiceID: %v err: %v", invoiceID, err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) oneInvoiceHasZeroTotalAmount(ctx context.Context) (context.Context, error) {
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
			0,
			0,
			0,
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
