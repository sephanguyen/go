package paymentsvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PaymentModifierService) updatePaymentStatusWithZeroAmount(ctx context.Context, db database.QueryExecer, paymentID, newStatus string) error {
	e := new(entities.Payment)
	database.AllNullEntity(e)

	err := multierr.Combine(
		e.PaymentID.Set(paymentID),
		e.PaymentStatus.Set(newStatus),
		e.Amount.Set(0),
	)

	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	err = s.PaymentRepo.UpdateWithFields(ctx, db, e, []string{"payment_status", "amount", "updated_at"})
	if err != nil {
		return fmt.Errorf("error Payment UpdateWithFields: %v", err)
	}

	return nil
}

// validateLatestPaymentStatus checks if the latest payment has FAILED status
// Note that a non-existing or nil payment is valid
func (s *PaymentModifierService) validateLatestPaymentStatus(ctx context.Context, invoiceID string) error {
	// The GetLatestPaymentDueDateByInvoiceID using a query that is ordered by created_at DESC
	// Basically this fetch the latest payment of an invoice
	payment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, s.DB, invoiceID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return status.Error(codes.Internal, fmt.Sprintf("PaymentRepo.GetLatestPaymentDueDateByInvoiceID err: %v", err))
	}

	if payment != nil && payment.PaymentStatus.String != invoice_pb.PaymentStatus_PAYMENT_FAILED.String() {
		return status.Error(codes.InvalidArgument, "latest payment should have FAILED status")
	}

	return nil
}
