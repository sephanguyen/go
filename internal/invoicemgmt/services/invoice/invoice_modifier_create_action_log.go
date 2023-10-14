package invoicesvc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InvoiceActionLogDetails struct {
	InvoiceID             string
	Action                invoice_pb.InvoiceAction
	PaymentSequenceNumber int32
	PaymentMethod         string
	ActionComment         string
}

// nolint:unused
func (s *InvoiceModifierService) createActionLog(ctx context.Context, db database.QueryExecer, actionLogDetails *InvoiceActionLogDetails) error {
	userID := interceptors.UserIDFromContext(ctx)
	var actionDetail string
	action := actionLogDetails.Action

	if len(actionLogDetails.InvoiceID) == 0 {
		return status.Error(codes.InvalidArgument, "invalid invoice id")
	}

	if actionLogDetails.Action != invoice_pb.InvoiceAction_INVOICE_FAILED && actionLogDetails.Action != invoice_pb.InvoiceAction_INVOICE_VOIDED && actionLogDetails.Action != invoice_pb.InvoiceAction_REMOVE_CREDIT_NOTE && actionLogDetails.Action != invoice_pb.InvoiceAction_EDIT_CREDIT_NOTE {
		if actionLogDetails.PaymentSequenceNumber == 0 {
			return status.Error(codes.InvalidArgument, "invalid payment sequence number")
		}
	}

	switch actionLogDetails.Action {
	case invoice_pb.InvoiceAction_INVOICE_VOIDED:
		actionDetail = ""
	case invoice_pb.InvoiceAction_INVOICE_ISSUED:
		if actionLogDetails.PaymentMethod == "" {
			return status.Error(codes.InvalidArgument, "invalid payment method")
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
		//backend will just passed Remove Credit Note but action log is still in Edit Credit Note
		action = invoice_pb.InvoiceAction_EDIT_CREDIT_NOTE
	default:
		return status.Error(codes.InvalidArgument, "invalid invoice action detail")
	}

	// Generate Action Log Entity
	actionLog := new(entities.InvoiceActionLog)
	database.AllNullEntity(actionLog)
	errs := []error{}

	if actionLogDetails.Action != invoice_pb.InvoiceAction_INVOICE_VOIDED {
		errs = append(errs, actionLog.PaymentSequenceNumber.Set(actionLogDetails.PaymentSequenceNumber))
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
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	// Create Action Log by storing on database
	if err := s.InvoiceActionLogRepo.Create(ctx, db, actionLog); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}
