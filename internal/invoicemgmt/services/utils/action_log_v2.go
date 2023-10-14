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
	"golang.org/x/exp/slices"
)

var actionNotRequiresPaymentV2 = []string{
	invoice_pb.InvoiceAction_INVOICE_ISSUED.String(),
	invoice_pb.InvoiceAction_INVOICE_FAILED.String(),
	invoice_pb.InvoiceAction_INVOICE_VOIDED.String(),
	invoice_pb.InvoiceAction_REMOVE_CREDIT_NOTE.String(),
	invoice_pb.InvoiceAction_EDIT_CREDIT_NOTE.String(),
	invoice_pb.InvoiceAction_INVOICE_ADJUSTED.String(),
	invoice_pb.InvoiceAction_PAYMENT_CANCELLED.String(),
	invoice_pb.InvoiceAction_INVOICE_REFUNDED.String(),
}

func GenActionLogEntity(ctx context.Context, actionLogDetails *InvoiceActionLogDetails) (*entities.InvoiceActionLog, error) {
	userID := interceptors.UserIDFromContext(ctx)
	var actionDetail string
	action := actionLogDetails.Action

	if len(actionLogDetails.InvoiceID) == 0 {
		return nil, fmt.Errorf("invalid invoice id")
	}

	// If the action is not in the list of actions that does not requires payment, validate the payment seq number
	if !slices.Contains(actionNotRequiresPaymentV2, actionLogDetails.Action.String()) {
		if actionLogDetails.PaymentSequenceNumber == 0 {
			return nil, fmt.Errorf("invalid payment sequence number")
		}
	}
	switch actionLogDetails.Action {
	case invoice_pb.InvoiceAction_INVOICE_VOIDED, invoice_pb.InvoiceAction_INVOICE_ISSUED, invoice_pb.InvoiceAction_INVOICE_ADJUSTED:
		actionDetail = ""
	case invoice_pb.InvoiceAction_PAYMENT_ADDED, invoice_pb.InvoiceAction_INVOICE_REFUNDED, invoice_pb.InvoiceAction_INVOICE_BULK_ISSUED:
		if actionLogDetails.PaymentMethod == "" {
			return nil, fmt.Errorf("invalid payment method")
		}
		actionDetail = actionLogDetails.PaymentMethod
	case invoice_pb.InvoiceAction_PAYMENT_APPROVED, invoice_pb.InvoiceAction_INVOICE_PAID:
		actionDetail = invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL.String()
	case invoice_pb.InvoiceAction_PAYMENT_CANCELLED, invoice_pb.InvoiceAction_INVOICE_FAILED:
		actionDetail = invoice_pb.PaymentStatus_PAYMENT_FAILED.String()
	case invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED:
		actionDetail = invoice_pb.InvoiceAction_PAYMENT_VALIDATE_FAILED.String()
	case invoice_pb.InvoiceAction_PAYMENT_VALIDATE_SUCCESS:
		actionDetail = invoice_pb.InvoiceAction_PAYMENT_VALIDATE_SUCCESS.String()
	case invoice_pb.InvoiceAction_PAYMENT_UPDATED:
		actionDetail = invoice_pb.InvoiceAction_PAYMENT_UPDATED.String()

	default:
		return nil, fmt.Errorf("invalid invoice action detail")
	}

	// Generate Action Log Entity
	actionLog := new(entities.InvoiceActionLog)
	database.AllNullEntity(actionLog)
	errs := []error{}

	if actionLogDetails.Action != invoice_pb.InvoiceAction_INVOICE_VOIDED &&
		actionLogDetails.Action != invoice_pb.InvoiceAction_INVOICE_ISSUED &&
		actionLogDetails.Action != invoice_pb.InvoiceAction_INVOICE_ADJUSTED {
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
		return nil, fmt.Errorf("multierr.Combine: %v", err)
	}

	return actionLog, nil
}

// nolint:unused
func CreateActionLogV2(ctx context.Context, db database.QueryExecer, actionLogDetails *InvoiceActionLogDetails, actionLogRepo InvoiceActionLogRepo) error {
	actionLog, err := GenActionLogEntity(ctx, actionLogDetails)
	if err != nil {
		return err
	}

	// Create Action Log by storing on database
	if err := actionLogRepo.Create(ctx, db, actionLog); err != nil {
		return err
	}

	return nil
}
