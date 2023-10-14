package invoicemgmt

import (
	"context"
	"fmt"

	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) issueInvoiceUsingV2Endpoint(ctx context.Context, signedInUser string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error
	ctx, err = s.signedAsAccount(ctx, signedInUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := &invoice_pb.IssueInvoiceRequestV2{
		InvoiceId: s.StepState.InvoiceID,
		Remarks:   "Issue this invoice",
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).IssueInvoiceV2(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisInvoiceHasPaymentWithStatus(ctx context.Context, existence string, paymentStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var status invoice_pb.PaymentStatus
	switch paymentStatus {
	case "PENDING":
		status = invoice_pb.PaymentStatus_PAYMENT_PENDING
	case "FAILED":
		status = invoice_pb.PaymentStatus_PAYMENT_FAILED
	case "SUCCESSFUL":
		status = invoice_pb.PaymentStatus_PAYMENT_SUCCESSFUL
	}

	stmt := `SELECT COUNT(*) FROM payment WHERE invoice_id = $1 AND payment_status = $2`
	row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.InvoiceID, status.String())

	var paymentCount int
	if err := row.Scan(&paymentCount); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if existence == "no existing" {
		if paymentCount != 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting that there is no existing payment got %d", paymentCount)
		}
	} else {
		if paymentCount == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting that there is existing payment got 0")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
