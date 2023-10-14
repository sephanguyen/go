package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) issuesInvoiceWithPaymentMethod(ctx context.Context, signedInUser string, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error
	ctx, err = s.signedAsAccount(ctx, signedInUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var paymentMethodInt int32
	var ok bool

	// Use generated map from proto to check if payment method name exists; if not, just use a random number
	if paymentMethodInt, ok = invoice_pb.PaymentMethod_value[paymentMethod]; !ok {
		paymentMethodInt = s.generateRandomNumber()
	}

	req := &invoice_pb.IssueInvoiceRequest{
		InvoiceIdString: s.StepState.InvoiceID,
		DueDate:         timestamppb.New(time.Now().Add(1 * time.Hour)),
		ExpiryDate:      timestamppb.New(time.Now().Add(1 * time.Hour)),
		PaymentMethod:   invoice_pb.PaymentMethod(paymentMethodInt),
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).IssueInvoice(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceHasDraftInvoiceStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.updateInvoiceStatus(StepStateToContext(ctx, stepState), invoice_pb.InvoiceStatus_DRAFT.String())

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceHasType(ctx context.Context, invoiceTypeFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var invoiceType invoice_pb.InvoiceType
	switch invoiceTypeFormat {
	case "SCHEDULED":
		invoiceType = invoice_pb.InvoiceType_SCHEDULED
	case "MANUAL":
		invoiceType = invoice_pb.InvoiceType_MANUAL
	}

	ctx, err := s.updateInvoiceType(StepStateToContext(ctx, stepState), invoiceType.String())

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsAnExistingInvoice(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.createInvoiceOfBillItem(StepStateToContext(ctx, stepState), invoice_pb.InvoiceStatus_DRAFT.String(), payment_pb.BillingStatus_BILLING_STATUS_BILLED.String())

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceStatusIsUpdatedToStatus(ctx context.Context, newStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var invoiceStatus string
	row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, "SELECT status FROM invoice WHERE invoice_id = $1", s.StepState.InvoiceID)

	if err := row.Scan(&invoiceStatus); err != nil {
		return ctx, fmt.Errorf("error finding invoice with invoice_id: %v: %w", s.StepState.InvoiceID, err)
	}

	if invoiceStatus != newStatus {
		return ctx, fmt.Errorf("invoice with invoice_id %v and status %v should have status of %v", s.StepState.InvoiceID, invoiceStatus, newStatus)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceExportedTagIsSetTo(ctx context.Context, isExportedStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	isExported := isExportedStr == "true"
	var actualIsExported bool
	row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, "SELECT is_exported FROM invoice WHERE invoice_id = $1", s.StepState.InvoiceID)

	if err := row.Scan(&actualIsExported); err != nil {
		return ctx, fmt.Errorf("error finding invoice with invoice_id: %v: %w", s.StepState.InvoiceID, err)
	}

	if isExported != actualIsExported {
		return ctx, fmt.Errorf("invoice with invoice_id %v exported tag should be set to %v", s.StepState.InvoiceID, isExported)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) paymentExportedTagIsSetTo(ctx context.Context, isExportedStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	isExported := isExportedStr == "true"

	var actualIsExported bool
	row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, "SELECT is_exported FROM payment WHERE invoice_id = $1", stepState.InvoiceID)

	if err := row.Scan(&actualIsExported); err != nil {
		return ctx, fmt.Errorf("error finding payment with payment_id: %v: %w", stepState.PaymentID, err)
	}

	if isExported != actualIsExported {
		return ctx, fmt.Errorf("invoice with payment_id %v exported tag should be set to %v", stepState.PaymentID, isExported)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) paymentHistoryIsRecordedWithPendingStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	paymentRecordCnt, err := s.getPaymentHistoryRecordCount(ctx, s.StepState.InvoiceID, invoice_pb.PaymentStatus_PAYMENT_PENDING.String())

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if paymentRecordCnt == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no payment record with invoice_id %v exists", s.StepState.InvoiceID)
	}

	if paymentRecordCnt > 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected 1 payment record inserted but got %v for invoice_id %v", paymentRecordCnt, s.StepState.InvoiceID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceIDIsNonexisting(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Overwrites the existing InvoiceID to simulate a non-existing invoice
	s.StepState.InvoiceID = idutil.ULIDNow()

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) noPaymentHistoryIsRecorded(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	paymentRecordCnt, err := s.getPaymentHistoryRecordCount(ctx, s.StepState.InvoiceID, invoice_pb.PaymentStatus_PAYMENT_PENDING.String())

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if paymentRecordCnt > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("payment record with invoice_id %v exists", s.StepState.InvoiceID)
	}

	return StepStateToContext(ctx, stepState), nil
}
