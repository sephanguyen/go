package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"go.uber.org/multierr"
)

type InvoiceActionLogDetails struct {
	InvoiceID                string
	Action                   invoice_pb.InvoiceAction
	PaymentSequenceNumber    int32
	PaymentMethod            string
	ActionComment            string
	BulkPaymentValidationsID string
}

type InvoiceActionLogRepo interface {
	Create(ctx context.Context, db database.QueryExecer, e *entities.InvoiceActionLog) error
}

func GenActionLogEntityV1(ctx context.Context, actionLogDetails *InvoiceActionLogDetails) (*entities.InvoiceActionLog, error) {
	userID := interceptors.UserIDFromContext(ctx)

	var actionDetail string
	action := actionLogDetails.Action

	if len(actionLogDetails.InvoiceID) == 0 {
		return nil, fmt.Errorf("invalid invoice id")
	}

	if actionLogDetails.Action != invoice_pb.InvoiceAction_INVOICE_FAILED && actionLogDetails.Action != invoice_pb.InvoiceAction_INVOICE_VOIDED && actionLogDetails.Action != invoice_pb.InvoiceAction_REMOVE_CREDIT_NOTE && actionLogDetails.Action != invoice_pb.InvoiceAction_EDIT_CREDIT_NOTE {
		if actionLogDetails.PaymentSequenceNumber == 0 {
			return nil, fmt.Errorf("invalid payment sequence number")
		}
	}

	switch actionLogDetails.Action {
	case invoice_pb.InvoiceAction_INVOICE_VOIDED:
		actionDetail = ""
	case invoice_pb.InvoiceAction_INVOICE_ISSUED:
		if actionLogDetails.PaymentMethod == "" {
			return nil, fmt.Errorf("invalid payment method")
		}
		actionDetail = actionLogDetails.PaymentMethod
	case invoice_pb.InvoiceAction_INVOICE_PAID:
		actionDetail = invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String()
	case invoice_pb.InvoiceAction_INVOICE_REFUNDED:
		actionDetail = invoice_pb.PaymentStatus_PAYMENT_REFUNDED.String()
	case invoice_pb.InvoiceAction_INVOICE_FAILED:
		actionDetail = invoice_pb.PaymentStatus_PAYMENT_FAILED.String()
	case invoice_pb.InvoiceAction_EDIT_CREDIT_NOTE:
		actionDetail = "Add Credit Note"
	case invoice_pb.InvoiceAction_REMOVE_CREDIT_NOTE:
		actionDetail = "Remove Credit Note"
		// backend will just passed Remove Credit Note but action log is still in Edit Credit Note
		action = invoice_pb.InvoiceAction_EDIT_CREDIT_NOTE
	case invoice_pb.InvoiceAction_PAYMENT_UPDATED:
		actionDetail = invoice_pb.InvoiceAction_PAYMENT_UPDATED.String()
	default:
		return nil, fmt.Errorf("invalid invoice action detail")
	}

	// Generate Action Log Entity
	actionLog := new(entities.InvoiceActionLog)
	database.AllNullEntity(actionLog)
	errs := []error{}

	if actionLogDetails.Action != invoice_pb.InvoiceAction_INVOICE_VOIDED {
		errs = append(errs, actionLog.PaymentSequenceNumber.Set(actionLogDetails.PaymentSequenceNumber))
	}

	if strings.TrimSpace(actionLogDetails.BulkPaymentValidationsID) != "" {
		errs = append(errs, actionLog.BulkPaymentValidationsID.Set(actionLogDetails.BulkPaymentValidationsID))
	}

	if err := multierr.Combine(
		actionLog.InvoiceID.Set(actionLogDetails.InvoiceID),
		actionLog.ActionComment.Set(actionLogDetails.ActionComment),
		actionLog.Action.Set(action),
		actionLog.ActionDetail.Set(actionDetail),
		actionLog.UserID.Set(userID),
	); err != nil {
		errs = append(errs, err)
	}

	if err := multierr.Combine(errs...); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}
	return actionLog, nil
}

func CreateActionLog(ctx context.Context, db database.QueryExecer, actionLogDetails *InvoiceActionLogDetails, actionLogRepo InvoiceActionLogRepo) error {
	actionLog, err := GenActionLogEntityV1(ctx, actionLogDetails)
	if err != nil {
		return err
	}

	// Create Action Log by storing on database
	if err := actionLogRepo.Create(ctx, db, actionLog); err != nil {
		return err
	}

	return nil
}
