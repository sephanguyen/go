package paymentsvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PaymentModifierService) BulkCancelPayment(ctx context.Context, req *invoice_pb.BulkCancelPaymentRequest) (*invoice_pb.BulkCancelPaymentResponse, error) {
	if strings.TrimSpace(req.BulkPaymentId) == "" {
		return nil, status.Error(codes.InvalidArgument, "bulk_payment_id cannot be empty")
	}

	// Find the bulk payment entity
	bulkPayment, err := s.BulkPaymentRepo.FindByBulkPaymentID(ctx, s.DB, req.BulkPaymentId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("BulkPaymentRepo.FindByBulkPaymentID err: %v", err))
	}

	// Check status of bulk payment entity
	if bulkPayment.BulkPaymentStatus.String != invoice_pb.BulkPaymentStatus_BULK_PAYMENT_PENDING.String() {
		return nil, status.Error(codes.InvalidArgument, "bulk payment is not in PENDING status")
	}

	// Find all payment belong to bulk payment entity
	payments, err := s.PaymentRepo.FindAllByBulkPaymentID(ctx, s.DB, req.BulkPaymentId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("PaymentRepo.FindAllByBulkPaymentID err: %v", err))
	}

	validPayments := []*entities.Payment{}
	for _, payment := range payments {
		// Validate payment (if there is exported payment, return error)
		if payment.IsExported.Bool {
			return nil, status.Error(codes.InvalidArgument, "at least one payment is already exported")
		}

		// Do not include payments that are not in PENDING status
		if payment.PaymentStatus.String != invoice_pb.PaymentStatus_PAYMENT_PENDING.String() {
			continue
		}

		validPayments = append(validPayments, payment)
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// Update the bulk payment status
		err = s.BulkPaymentRepo.UpdateBulkPaymentStatusByIDs(ctx, tx, invoice_pb.BulkPaymentStatus_BULK_PAYMENT_CANCELLED.String(), []string{req.BulkPaymentId})
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("BulkPaymentRepo.UpdateBulkPaymentStatusByIDs err: %v", err))
		}

		// No need to proceed on updating payments and creating action log when there is no valid payments
		if len(validPayments) == 0 {
			return nil
		}

		// Update the payment status
		err = s.PaymentRepo.UpdateStatusAndAmountByPaymentIDs(ctx, tx, getPaymentIDsFromPayments(validPayments), invoice_pb.PaymentStatus_PAYMENT_FAILED.String(), 0)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("PaymentRepo.UpdateStatusAndAmountByPaymentIDs err: %v", err))
		}

		// Create action logs
		for _, payment := range validPayments {
			actionDetails := &utils.InvoiceActionLogDetails{
				InvoiceID:             payment.InvoiceID.String,
				Action:                invoice_pb.InvoiceAction_PAYMENT_CANCELLED,
				ActionComment:         "",
				PaymentSequenceNumber: payment.PaymentSequenceNumber.Int,
			}

			if err := utils.CreateActionLogV2(ctx, tx, actionDetails, s.InvoiceActionLogRepo); err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("utils.CreateActionLogV2 err: %v", err))
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &invoice_pb.BulkCancelPaymentResponse{
		Successful: true,
	}, nil
}

func getPaymentIDsFromPayments(payments []*entities.Payment) []string {
	paymentIDs := []string{}

	for _, payment := range payments {
		paymentIDs = append(paymentIDs, payment.PaymentID.String)
	}

	return paymentIDs
}
