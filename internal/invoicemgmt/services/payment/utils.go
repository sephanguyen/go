package paymentsvc

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"go.uber.org/multierr"
)

func GenPaymentFromAddPaymentRequest(req *invoice_pb.AddInvoicePaymentRequest, invoice *entities.Invoice) (*entities.Payment, error) {
	e := new(entities.Payment)
	database.AllNullEntity(e)

	now := time.Now()
	if err := multierr.Combine(
		e.InvoiceID.Set(invoice.InvoiceID.String),
		e.PaymentMethod.Set(req.PaymentMethod.String()),
		e.PaymentDueDate.Set(database.TimestamptzFromPb(req.DueDate)),
		e.PaymentExpiryDate.Set(database.TimestamptzFromPb(req.ExpiryDate)),
		e.PaymentStatus.Set(invoice_pb.PaymentStatus_PAYMENT_PENDING.String()),
		e.StudentID.Set(invoice.StudentID.String),
		e.Amount.Set(req.Amount),
		e.IsExported.Set(false),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	return e, nil
}
