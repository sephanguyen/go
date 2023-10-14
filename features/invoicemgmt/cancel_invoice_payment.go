package invoicemgmt

import (
	"context"
	"fmt"

	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) adminCancelsAnInvoiceWithRemarks(ctx context.Context, remarks string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.CancelInvoicePaymentRequest{
		InvoiceId: stepState.InvoiceID,
	}

	req.Remarks = remarks

	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).CancelInvoicePayment(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) actionLogHasFailedAction(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	actionLog, err := s.getLatestInvoiceActionLogByInvoiceID(ctx, stepState.InvoiceID)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if actionLog.Action.String != invoice_pb.InvoiceAction_INVOICE_FAILED.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error Action Log Action must be failed")
	}

	return StepStateToContext(ctx, stepState), nil

}
