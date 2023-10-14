package paymentsvc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	invoicemgmt_entities "github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PaymentModifierService) ApproveInvoicePaymentV2(ctx context.Context, req *invoice_pb.ApproveInvoicePaymentV2Request) (*invoice_pb.ApproveInvoicePaymentV2Response, error) {
	failedResp := &invoice_pb.ApproveInvoicePaymentV2Response{
		Successful: false,
	}

	err := validateApprovePaymentRequest(req)
	if err != nil {
		return failedResp, err
	}

	invoice, err := retrieveValidInvoiceFromApprovePayment(ctx, s.DB, s, req.InvoiceId)
	if err != nil {
		return failedResp, err
	}

	payment, err := retrieveValidPaymentFromApprovePayment(ctx, s, req.InvoiceId)
	if err != nil {
		return failedResp, err
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// for now the total and amount paid is equal this means it is fully paid after one approval
		// it will be change on partial payment
		invoice.AmountPaid = invoice.Total
		invoice.Status = database.Text(invoice_pb.InvoiceStatus_PAID.String())
		invoice.OutstandingBalance = database.Numeric(0)

		err = s.InvoiceRepo.UpdateWithFields(ctx, tx, invoice, []string{"status", "updated_at", "outstanding_balance", "amount_paid"})
		if err != nil {
			return fmt.Errorf("error Invoice UpdateWithFields: %v", err)
		}

		err = multierr.Combine(
			payment.PaymentStatus.Set(database.Text(invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String())),
			payment.PaymentDate.Set(database.TimestamptzFromPb(req.PaymentDate)),
			payment.ReceiptDate.Set(database.Timestamptz(time.Now())),
		)
		if err != nil {
			return fmt.Errorf("multierr.Combine: %w", err)
		}

		err = s.PaymentRepo.UpdateWithFields(ctx, tx, payment, []string{"payment_date", "payment_status", "receipt_date", "updated_at"})
		if err != nil {
			return fmt.Errorf("error Payment UpdateWithFields: %v", err)
		}

		// Create action log
		actionDetails := &utils.InvoiceActionLogDetails{
			InvoiceID:             req.InvoiceId,
			Action:                invoice_pb.InvoiceAction_PAYMENT_APPROVED,
			ActionComment:         req.Remarks,
			PaymentSequenceNumber: payment.PaymentSequenceNumber.Int,
		}

		if err := utils.CreateActionLogV2(ctx, tx, actionDetails, s.InvoiceActionLogRepo); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return failedResp, status.Error(codes.Internal, err.Error())
	}

	return &invoice_pb.ApproveInvoicePaymentV2Response{
		Successful: true,
	}, nil
}

func validateApprovePaymentRequest(req *invoice_pb.ApproveInvoicePaymentV2Request) error {
	if len(strings.TrimSpace(req.InvoiceId)) == 0 {
		return status.Error(codes.InvalidArgument, "invoice id cannot be empty")
	}

	if req.PaymentDate == nil {
		return status.Error(codes.InvalidArgument, "payment date cannot be empty")
	}

	return nil
}

func retrieveValidInvoiceFromApprovePayment(ctx context.Context, db database.QueryExecer, s *PaymentModifierService, invoiceID string) (*invoicemgmt_entities.Invoice, error) {
	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, db, invoiceID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error Invoice RetrieveInvoiceByInvoiceID: %v", err))
	}

	// Only allow an invoice with ISSUED status
	if invoice.Status.String != invoice_pb.InvoiceStatus_ISSUED.String() {
		return nil, status.Error(codes.InvalidArgument, "Invoice should be in ISSUED status")
	}

	// Only allow an invoice total with greater than 0
	if invoice.Total.Int.Int64() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Invoice total should be greater than 0")
	}

	return invoice, nil
}

func retrieveValidPaymentFromApprovePayment(ctx context.Context, s *PaymentModifierService, invoiceID string) (*invoicemgmt_entities.Payment, error) {
	payment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, s.DB, invoiceID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error Payment GetLatestPaymentDueDateByInvoiceID: %v", err))
	}

	// Only allow invoice with PENDING payment status
	if payment.PaymentStatus.String != invoice_pb.PaymentStatus_PAYMENT_PENDING.String() {
		return nil, status.Error(codes.InvalidArgument, "Payment should be in PENDING status")
	}

	// Only allow payment with Bank Transfer And Cash payment method
	if !constant.ApprovePaymentAllowedMethods[payment.PaymentMethod.String] {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("payment method %v is not allowed only cash and bank transfer accepted", payment.PaymentMethod.String))
	}

	return payment, nil
}
