package invoicemgmt

import (
	"context"

	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) issuesInvoiceUsingV2EndpointWithPaymentMethodAndDates(ctx context.Context, signedInUser, paymentMethodStr, dueDate, expiryDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var err error
	ctx, err = s.signedAsAccount(ctx, signedInUser)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	paymentMethod, err := getPaymentMethodFromStr(paymentMethodStr)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	dueDateTimestamp := getFormattedTimestampDate(dueDate)
	expiryDateTimestamp := getFormattedTimestampDate(expiryDate)

	req := &invoice_pb.IssueInvoiceRequestV2{
		InvoiceId:     s.StepState.InvoiceID,
		Remarks:       "Issue this invoice",
		PaymentMethod: paymentMethod,
		DueDate:       dueDateTimestamp,
		ExpiryDate:    expiryDateTimestamp,
		Amount:        float64(stepState.InvoiceTotal),
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).IssueInvoiceV2(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}
