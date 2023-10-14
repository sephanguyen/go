package paymentsvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PaymentModifierService) CancelInvoicePaymentV2(ctx context.Context, req *invoice_pb.CancelInvoicePaymentV2Request) (*invoice_pb.CancelInvoicePaymentV2Response, error) {
	if err := s.validateCancelInvoiceRequest(ctx, s.DB, req); err != nil {
		return nil, err
	}

	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		//  Get the payment of the corresponding invoiceID
		payment, err := s.getInvoiceLatestPayment(ctx, tx, req.InvoiceId)
		if err != nil {
			return err
		}
		if payment == nil {
			return status.Error(codes.Internal, "error Payment: Payment is nil")
		}

		if payment.PaymentStatus.String != invoice_pb.PaymentStatus_PAYMENT_PENDING.String() {
			return status.Error(codes.Internal, "error Payment: Payment status should be pending")
		}

		if payment.PaymentMethod.String == invoice_pb.PaymentMethod_DIRECT_DEBIT.String() && payment.IsExported.Bool {
			return status.Error(codes.Internal, "error Payment: Payment method direct debit should not be exported")
		}

		if err := s.updatePaymentStatusWithZeroAmount(ctx, tx, payment.PaymentID.String, invoice_pb.PaymentStatus_PAYMENT_FAILED.String()); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// check if payments belong to bulk
		err = s.validateAndCancelBulkPaymentRecord(ctx, tx, payment)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// Create action log
		actionDetails := &utils.InvoiceActionLogDetails{
			InvoiceID:             req.InvoiceId,
			Action:                invoice_pb.InvoiceAction_PAYMENT_CANCELLED,
			ActionComment:         req.Remarks,
			PaymentSequenceNumber: payment.PaymentSequenceNumber.Int,
		}

		if err := utils.CreateActionLogV2(ctx, tx, actionDetails, s.InvoiceActionLogRepo); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &invoice_pb.CancelInvoicePaymentV2Response{
		Successful: true,
	}, nil
}

func (s *PaymentModifierService) getInvoiceLatestPayment(ctx context.Context, db database.QueryExecer, invoiceID string) (*entities.Payment, error) {
	payment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, db, invoiceID)

	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("error GetLatestPaymentDueDateByInvoiceID: %v", err)
	}

	if payment != nil {
		return payment, nil
	}

	return nil, nil
}

func (s *PaymentModifierService) validateCancelInvoiceRequest(ctx context.Context, db database.QueryExecer, req *invoice_pb.CancelInvoicePaymentV2Request) error {
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
		return status.Error(codes.InvalidArgument, "invoice should be in ISSUED status")
	}

	return nil
}

func (s *PaymentModifierService) validateAndCancelBulkPaymentRecord(ctx context.Context, db database.QueryExecer, payment *entities.Payment) error {
	if payment.BulkPaymentID.Status == pgtype.Present && strings.TrimSpace(payment.BulkPaymentID.String) != "" {
		// check if there's other payments that are not failed
		paymentCount, err := s.PaymentRepo.CountOtherPaymentsByBulkPaymentIDNotInStatus(ctx, db, payment.BulkPaymentID.String, payment.PaymentID.String, invoice_pb.PaymentStatus_PAYMENT_FAILED.String())
		if err != nil {
			return fmt.Errorf("PaymentRepo.CountOtherPaymentsByBulkPaymentIDNotInStatus err: %v", err)
		}

		// if there's no other payments that are not failed in status aside from the current payment; cancel the bulk payment record
		if paymentCount == 0 {
			if err := s.BulkPaymentRepo.UpdateBulkPaymentStatusByIDs(ctx, db, invoice_pb.BulkPaymentStatus_BULK_PAYMENT_CANCELLED.String(), []string{payment.BulkPaymentID.String}); err != nil {
				return fmt.Errorf("error BulkPaymentRepo UpdateBulkPaymentStatusByIDs: %v", err)
			}
		}
	}
	return nil
}
