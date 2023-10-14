package invoicemgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) studentHasInvoiceRecords(ctx context.Context, invoices string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if invoices != "no existing" {
		splitInvoices := strings.Split(invoices, "-")
		var err error
		for _, invoice := range splitInvoices {
			err = s.createInvoiceBasedOnStatus(ctx, invoice)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}

	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentIsAtTheInvoiceListScreen(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &invoice_pb.RetrieveInvoiceRecordsRequest{
		Paging: &cpb.Paging{
			Limit: invoiceConst.DefaultPageLimit,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		},
	}
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentSelectsThisExistingStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Request.(*invoice_pb.RetrieveInvoiceRecordsRequest).StudentId = s.StepState.StudentID
	req := stepState.Request.(*invoice_pb.RetrieveInvoiceRecordsRequest)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).RetrieveInvoiceRecords(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) recordsFoundWithDefaultLimitAreDisplayedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	invoiceRecords := stepState.Response.(*invoice_pb.RetrieveInvoiceRecordsResponse).InvoiceRecords
	if len(invoiceRecords) != invoiceConst.DefaultPageLimit {
		return ctx, fmt.Errorf("unexpected invoice records response count %d", len(invoiceRecords))
	}

	// checking display due date
	for _, invoiceRecord := range invoiceRecords {
		// invoice that is not draft should have due date
		if invoiceRecord.DueDate == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected payment due date but got null")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) noInvoiceDraftRecordsFound(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	invoiceRecords := stepState.Response.(*invoice_pb.RetrieveInvoiceRecordsResponse).InvoiceRecords
	// checking draft invoice if exist
	for _, invoiceRecord := range invoiceRecords {
		if invoiceRecord.InvoiceStatus.String() == "DRAFT" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected draft invoice found")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) noRecordsFoundDisplayedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	if stepState.Response != nil && int32(len(stepState.Response.(*invoice_pb.RetrieveInvoiceRecordsResponse).InvoiceRecords)) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected invoice records for this student")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentScrollsDownToDisplayAllRecords(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// the initial response is added on the response count record
	countRecord := int32(len(stepState.Response.(*invoice_pb.RetrieveInvoiceRecordsResponse).InvoiceRecords))
	// get the previous request
	req := stepState.Request.(*invoice_pb.RetrieveInvoiceRecordsRequest)
	// assign the initial response next page for pagination
	stepStateResponse := stepState.Response
	for {
		req.Paging = stepStateResponse.(*invoice_pb.RetrieveInvoiceRecordsResponse).NextPage
		stepState.RequestSentAt = time.Now()
		stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).RetrieveInvoiceRecords(contextWithToken(ctx), req)
		stepStateResponse = stepState.Response
		if countRecord == countRecord+int32(len(stepState.Response.(*invoice_pb.RetrieveInvoiceRecordsResponse).InvoiceRecords)) {
			break
		}
		countRecord += int32(len(stepState.Response.(*invoice_pb.RetrieveInvoiceRecordsResponse).InvoiceRecords))
	}
	stepState.RetrieveRecordCount = countRecord

	req.Paging.Limit = 10
	req.Paging.Offset = &cpb.Paging_OffsetInteger{
		OffsetInteger: 0,
	}
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).RetrieveInvoiceRecords(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) allRecordsFoundAreDisplayedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	if int32(len(stepState.Response.(*invoice_pb.RetrieveInvoiceRecordsResponse).InvoiceRecords)) != stepState.RetrieveRecordCount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve all records: unexpected row affected")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) parentHasAnotherExistingStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.thisParentHasAnExistingStudent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) loginsLearnerApp(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, role)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
