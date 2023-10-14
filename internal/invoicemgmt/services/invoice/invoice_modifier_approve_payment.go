package invoicesvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	invoicemgmt_entities "github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InvoiceModifierService) ApproveInvoicePayment(ctx context.Context, req *invoice_pb.ApproveInvoicePaymentRequest) (*invoice_pb.ApproveInvoicePaymentResponse, error) {
	failedResp := &invoice_pb.ApproveInvoicePaymentResponse{
		Successful: false,
	}

	successfulResp := &invoice_pb.ApproveInvoicePaymentResponse{
		Successful: true,
	}

	// Validate the request
	invoice, payment, err := validatePaymentRequest(ctx, s.DB, s, req)
	if err != nil {
		return failedResp, err
	}

	// Retrieve all invoice bill items
	invoiceBillItems, err := s.InvoiceBillItemRepo.FindAllByInvoiceID(ctx, s.DB, req.InvoiceId)
	if err != nil {
		return failedResp, status.Error(codes.Internal, fmt.Sprintf("error InvoiceBillItem FindAllByInvoiceID: %v", err))
	}

	if len(invoiceBillItems.ToArray()) == 0 {
		return failedResp, status.Error(codes.Internal, "No invoice bill items; cannot approve payment")
	}

	// Check invoice total to determine if the status is PAID or REFUNDED
	invoiceStatus, err := s.getStatusBasedOnInvoiceTotal(invoice.Total)
	if err != nil {
		return failedResp, err
	}

	invoice.Status = database.Text(invoiceStatus)
	invoice.OutstandingBalance = database.Numeric(0)

	// Generate action logs details
	actionDetails := s.generateActionLogDetails(invoice, payment, req.Remarks)

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		updateFields := []string{"status", "updated_at", "outstanding_balance"}
		switch invoice.Status.String {
		case invoice_pb.InvoiceStatus_PAID.String():
			updateFields = append(updateFields, "amount_paid")
			invoice.AmountPaid = invoice.Total
		case invoice_pb.InvoiceStatus_REFUNDED.String():
			updateFields = append(updateFields, "amount_refunded")
			invoice.AmountRefunded = invoice.Total
		}

		err = s.InvoiceRepo.UpdateWithFields(ctx, tx, invoice, updateFields)
		if err != nil {
			return fmt.Errorf("error Invoice UpdateWithFields: %v", err)
		}

		payment.PaymentStatus = database.Text(invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String())
		payment.PaymentDate = database.TimestamptzFromPb(req.PaymentDate)

		err = s.PaymentRepo.UpdateWithFields(ctx, tx, payment, []string{"payment_date", "payment_status", "updated_at"})
		if err != nil {
			return fmt.Errorf("error Payment UpdateWithFields: %v", err)
		}

		if err := s.createActionLog(ctx, tx, actionDetails); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return failedResp, status.Error(codes.Internal, err.Error())
	}

	return successfulResp, nil
}

func validatePaymentRequest(ctx context.Context, db database.QueryExecer, s *InvoiceModifierService, req *invoice_pb.ApproveInvoicePaymentRequest) (*invoicemgmt_entities.Invoice, *invoicemgmt_entities.Payment, error) {
	if len(strings.TrimSpace(req.InvoiceId)) == 0 {
		return nil, nil, status.Error(codes.InvalidArgument, "InvoiceId is required")
	}

	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, db, req.InvoiceId)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, fmt.Sprintf("error Invoice RetrieveInvoiceByInvoiceID: %v", err))
	}

	// Only allow an invoice with ISSUED status
	if invoice.Status.String != invoice_pb.InvoiceStatus_ISSUED.String() {
		return nil, nil, status.Error(codes.InvalidArgument, "Invoice should be in ISSUED status")
	}

	payment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, db, req.InvoiceId)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, fmt.Sprintf("error Payment GetLatestPaymentDueDateByInvoiceID: %v", err))
	}

	// Only allow invoice with PENDING payment status
	if payment.PaymentStatus.String != invoice_pb.PaymentStatus_PAYMENT_PENDING.String() {
		return nil, nil, status.Error(codes.InvalidArgument, "Payment should be in PENDING status")
	}

	return invoice, payment, nil
}

func (s *InvoiceModifierService) getStatusBasedOnInvoiceTotal(invoiceTotal pgtype.Numeric) (string, error) {
	getExactInvoiceTotalValue, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoiceTotal, "2")
	if err != nil {
		return "", err
	}

	invoiceStatus := invoice_pb.InvoiceStatus_PAID.String()
	if getExactInvoiceTotalValue < 0 {
		invoiceStatus = invoice_pb.InvoiceStatus_REFUNDED.String()
	}

	return invoiceStatus, nil
}
