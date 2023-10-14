package invoicesvc

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InvoiceModifierService) RefundInvoice(ctx context.Context, req *invoice_pb.RefundInvoiceRequest) (*invoice_pb.RefundInvoiceResponse, error) {
	invoice, err := s.validateRefundInvoice(ctx, req)
	if err != nil {
		return nil, err
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Update invoice outstanding_balance, amount_refunded and status
		if err := s.updateInvoiceToRefunded(ctx, tx, invoice, req.Amount); err != nil {
			return err
		}

		// Get the equivalent payment method for action log use
		paymentMethod, err := getPaymentMethodFromRefundMethod(req.RefundMethod)
		if err != nil {
			return err
		}

		// Generate action logs details
		actionDetails := &utils.InvoiceActionLogDetails{
			InvoiceID:     req.InvoiceId,
			PaymentMethod: paymentMethod.String(),
			Action:        invoice_pb.InvoiceAction_INVOICE_REFUNDED,
			ActionComment: req.Remarks,
		}
		if err := utils.CreateActionLogV2(ctx, tx, actionDetails, s.InvoiceActionLogRepo); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &invoice_pb.RefundInvoiceResponse{
		Successful: true,
	}, nil
}

func (s *InvoiceModifierService) validateRefundInvoice(ctx context.Context, req *invoice_pb.RefundInvoiceRequest) (*entities.Invoice, error) {
	if err := validateRefundInvoiceRequest(req); err != nil {
		return nil, err
	}

	invoice, err := s.validateInvoiceToBeRefunded(ctx, req)
	if err != nil {
		return nil, err
	}

	return invoice, nil
}

func (s *InvoiceModifierService) validateInvoiceToBeRefunded(ctx context.Context, req *invoice_pb.RefundInvoiceRequest) (*entities.Invoice, error) {
	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.DB, req.InvoiceId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("InvoiceRepo.RetrieveInvoiceByInvoiceID err: %v", err))
	}

	// Check if invoice is ISSUED
	if invoice.Status.String != invoice_pb.InvoiceStatus_ISSUED.String() {
		return nil, status.Error(codes.InvalidArgument, "invoice status should be ISSUED")
	}

	exactTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Check if invoice total is negative
	if exactTotal >= 0 {
		return nil, status.Error(codes.InvalidArgument, "invoice total should be negative")
	}

	exactOutstandingBalance, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.OutstandingBalance, "2")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Check if outstanding balance is negative
	if exactOutstandingBalance >= 0 {
		return nil, status.Error(codes.InvalidArgument, "invoice outstanding balance should be negative")
	}

	// Check if the amount is equal to the invoice's outstanding balance
	if exactOutstandingBalance != req.Amount {
		return nil, status.Error(codes.InvalidArgument, "the given amount should be equal to the invoice outstanding balance")
	}

	return invoice, nil
}

func (s *InvoiceModifierService) updateInvoiceToRefunded(ctx context.Context, db database.QueryExecer, invoice *entities.Invoice, amount float64) error {
	err := multierr.Combine(
		computeInvoiceBalanceAndRefundAmount(invoice, amount),
		invoice.Status.Set(invoice_pb.InvoiceStatus_REFUNDED),
	)
	if err != nil {
		return err
	}

	if err := s.InvoiceRepo.UpdateWithFields(ctx, db, invoice, []string{"status", "amount_refunded", "outstanding_balance", "updated_at"}); err != nil {
		return err
	}

	return nil
}

func computeInvoiceBalanceAndRefundAmount(invoice *entities.Invoice, amount float64) error {
	exactOutstandingBalance, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.OutstandingBalance, "2")
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// To compute the new outstanding balance, get the difference of the current balance and the amount to refund
	// It is expected that both of these values are negative
	newOutstandingBalance := exactOutstandingBalance - amount
	newRefundedAmount := amount

	return multierr.Combine(
		invoice.OutstandingBalance.Set(newOutstandingBalance),
		invoice.AmountRefunded.Set(newRefundedAmount),
	)
}

func validateRefundInvoiceRequest(req *invoice_pb.RefundInvoiceRequest) error {
	if strings.TrimSpace(req.InvoiceId) == "" {
		return status.Error(codes.InvalidArgument, "invoice ID cannot be empty")
	}

	if req.Amount >= 0 {
		return status.Error(codes.InvalidArgument, "amount should be negative value")
	}

	_, allowed := constant.RefundInvoiceAllowedMethods[req.RefundMethod.String()]
	if !allowed {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("refund method %v is not allowed", req.RefundMethod.String()))
	}

	return nil
}

func getPaymentMethodFromRefundMethod(refundMethod invoice_pb.RefundMethod) (invoice_pb.PaymentMethod, error) {
	var paymentMethod invoice_pb.PaymentMethod

	switch refundMethod {
	case invoice_pb.RefundMethod_REFUND_CASH:
		paymentMethod = invoice_pb.PaymentMethod_CASH
	case invoice_pb.RefundMethod_REFUND_BANK_TRANSFER:
		paymentMethod = invoice_pb.PaymentMethod_BANK_TRANSFER
	default:
		return paymentMethod, fmt.Errorf("refund method %v is not yet supported", refundMethod.String())
	}

	return paymentMethod, nil
}
