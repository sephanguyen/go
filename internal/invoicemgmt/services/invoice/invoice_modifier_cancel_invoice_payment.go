package invoicesvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InvoiceModifierService) CancelInvoicePayment(ctx context.Context, req *invoice_pb.CancelInvoicePaymentRequest) (*invoice_pb.CancelInvoicePaymentResponse, error) {
	if err := validateCancelInvoiceRequest(ctx, s.DB, s, req); err != nil {
		return &invoice_pb.CancelInvoicePaymentResponse{
			Successful: false,
		}, err
	}

	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// Update invoice status to Failed
		if err := s.updateInvoiceStatus(ctx, tx, req.InvoiceId, invoice_pb.InvoiceStatus_FAILED.String()); err != nil {
			return fmt.Errorf("error Invoice Update: %v", err)
		}

		//  Get the payment of the corresponding invoiceID
		payment, err := s.getPaymentID(ctx, tx, req)
		if err != nil {
			return err
		}

		if payment == nil {
			return fmt.Errorf("error Payment: Payment is nil")
		}

		if payment.PaymentStatus.String != invoice_pb.PaymentStatus_PAYMENT_PENDING.String() {
			return fmt.Errorf("error Payment: Payment should be pending")
		}

		if err := s.updatePaymentStatus(ctx, tx, payment.PaymentID.String, invoice_pb.PaymentStatus_PAYMENT_FAILED.String()); err != nil {
			return fmt.Errorf("error Payment Update: %v", err)
		}

		// Create action log
		actionDetails := &InvoiceActionLogDetails{
			InvoiceID:             req.InvoiceId,
			Action:                invoice_pb.InvoiceAction_INVOICE_FAILED,
			ActionComment:         req.Remarks,
			PaymentSequenceNumber: payment.PaymentSequenceNumber.Int,
		}
		if err := s.createActionLog(ctx, tx, actionDetails); err != nil {
			return err
		}

		if err != nil && err != pgx.ErrNoRows {
			return fmt.Errorf("error InvoiceActionLogDetails: %v", err)
		}

		return nil
	})

	if err != nil {
		return &invoice_pb.CancelInvoicePaymentResponse{
			Successful: false,
		}, err
	}

	// Return success if no error
	return &invoice_pb.CancelInvoicePaymentResponse{
		Successful: true,
	}, nil
}

func (s *InvoiceModifierService) getPaymentID(ctx context.Context, tx pgx.Tx, req *invoice_pb.CancelInvoicePaymentRequest) (*entities.Payment, error) {
	payment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, tx, req.InvoiceId)

	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error GetLatestPaymentDueDateByInvoiceID: %v", err)
	}

	if payment != nil {
		return payment, nil
	}

	return nil, nil
}

func validateCancelInvoiceRequest(ctx context.Context, db database.QueryExecer, s *InvoiceModifierService, req *invoice_pb.CancelInvoicePaymentRequest) error {
	// check whether there are selected invoice
	if strings.TrimSpace(req.InvoiceId) == "" {
		return status.Error(codes.InvalidArgument, "invoiceID is required")
	}

	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, db, req.InvoiceId)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("error Invoice RetrieveInvoiceByInvoiceID: %v", err))
	}

	// Only allow Issued invoice status to be processed
	if invoice.Status.String != invoice_pb.InvoiceStatus_ISSUED.String() {
		return status.Error(codes.InvalidArgument, "Invoice should be in ISSUED status")
	}

	return nil
}
