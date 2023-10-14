package invoicemgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
)

func (s *suite) addsInvoiceAdjustmentWithAmount(ctx context.Context, recordCount int, adjustAmountFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := s.generateUpsertInvoiceAdjustmentRequest(ctx)

	amountSlice := strings.Split(adjustAmountFormat, "&")
	for i := 0; i < recordCount; i++ {
		amountParse, err := strconv.ParseFloat(amountSlice[i], 64)
		if err != nil {
			return nil, fmt.Errorf("error on converting amount: %v", amountSlice[i])
		}
		req.InvoiceSubTotal += amountParse
		req.InvoiceTotal += amountParse
		invoiceAdjustmentDetail := &invoice_pb.InvoiceAdjustmentDetail{
			Description:             fmt.Sprintf("test-%v", i),
			Amount:                  amountParse,
			InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT,
		}

		req.InvoiceAdjustmentDetails = append(req.InvoiceAdjustmentDetails, invoiceAdjustmentDetail)
	}

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) applyTheAdjustmentOnTheInvoice(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	s.StepState.Response, s.StepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).UpsertInvoiceAdjustments(contextWithToken(ctx), stepState.Request.(*invoice_pb.UpsertInvoiceAdjustmentsRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceTotalSubtotalAndOutstandingBalanceAreCorrectlyUpdatedToAmount(ctx context.Context, expectedAmount float64) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	invoiceRepo := &repositories.InvoiceRepo{}
	invoice, err := invoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.InvoiceID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("invoiceRepo.RetrieveInvoiceByInvoiceID err: %v", err)
	}

	totalAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("utils.GetFloat64ExactValueAndDecimalPlaces err: %v on invoice total: %v", err, totalAmount)
	}
	if totalAmount != expectedAmount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected total amount: %v received: %v", expectedAmount, totalAmount)
	}

	subTotalAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.SubTotal, "2")
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("utils.GetFloat64ExactValueAndDecimalPlaces err: %v on invoice subtotal: %v", err, subTotalAmount)
	}

	if subTotalAmount != expectedAmount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected subtotal amount: %v received: %v", expectedAmount, subTotalAmount)
	}

	outstandingBalance, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.OutstandingBalance, "2")
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("utils.GetFloat64ExactValueAndDecimalPlaces err: %v on invoice outstanding_balance: %v", err, subTotalAmount)
	}

	if outstandingBalance != expectedAmount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected outstanding_balance amount: %v received: %v", expectedAmount, subTotalAmount)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) editsExistingInvoiceAdjustmentThatHasAmountUpdatedToAmount(ctx context.Context, recordToEditCount int, existingAmount, updatedAmount string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := s.generateUpsertInvoiceAdjustmentRequest(ctx)

	existingAmountSlice := strings.Split(existingAmount, "&")
	updatedAmountSlice := strings.Split(updatedAmount, "&")

	for i := 0; i < recordToEditCount; i++ {
		existingAmountParse, err := strconv.ParseFloat(strings.TrimSpace(existingAmountSlice[i]), 64)
		if err != nil {
			return nil, fmt.Errorf("error on converting amount: %v to float", existingAmountSlice[i])
		}
		invoiceAdjustmentID, ok := stepState.InvoiceAdjustMapAmount[existingAmountSlice[i]]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot find invoice adjustment id on amount: %v", existingAmountParse)
		}

		updatedAmountParse, err := strconv.ParseFloat(strings.TrimSpace(updatedAmountSlice[i]), 64)
		if err != nil {
			return nil, fmt.Errorf("error on converting amount: %v to float", updatedAmountSlice[i])
		}
		// deduct the existing amount and just add the updated amount info
		req.InvoiceSubTotal += (updatedAmountParse - existingAmountParse)
		req.InvoiceTotal += (updatedAmountParse - existingAmountParse)

		invoiceAdjustmentDetail := &invoice_pb.InvoiceAdjustmentDetail{
			InvoiceAdjustmentId:     invoiceAdjustmentID,
			Description:             fmt.Sprintf("test-%v", i),
			Amount:                  updatedAmountParse,
			InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT,
		}

		req.InvoiceAdjustmentDetails = append(req.InvoiceAdjustmentDetails, invoiceAdjustmentDetail)
	}

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereAreCreatedInvoiceAdjustmentWithAmount(ctx context.Context, recordCount int, adjustAmountFormat string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	adjustAmountSlice := strings.Split(adjustAmountFormat, "&")
	invoiceAdjustmentMap := make(map[string]string)
	for i := 0; i < recordCount; i++ {
		trimAmountSliceValue := strings.TrimSpace(adjustAmountSlice[i])
		amountParse, err := strconv.ParseFloat(trimAmountSliceValue, 64)
		stepState.InvoiceTotalAmount[0] += amountParse
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error on converting amount: %v to float", trimAmountSliceValue)
		}
		// using insert entities to get ids from invoice adjustment for existing record scenarios
		err = InsertEntities(
			stepState,
			s.EntitiesCreator.CreateInvoiceAdjustment(ctx, s.InvoiceMgmtPostgresDBTrace, stepState.InvoiceID, stepState.StudentIds[0], amountParse),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		ctx, err = s.thisInvoiceHasTotalAmount(ctx, stepState.InvoiceTotalAmount[0])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, ok := invoiceAdjustmentMap[trimAmountSliceValue]
		if ok {
			continue
		}
		invoiceAdjustmentMap[trimAmountSliceValue] = stepState.InvoiceAdjustmentIDs[i]
	}

	stepState.InvoiceAdjustMapAmount = invoiceAdjustmentMap

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) deletesExistingInvoiceAdjustmentThatHasAmount(ctx context.Context, recordToEditCount int, existingAmount string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := s.generateUpsertInvoiceAdjustmentRequest(ctx)

	existingAmountSlice := strings.Split(existingAmount, "&")

	for i := 0; i < recordToEditCount; i++ {
		existingAmountParse, err := strconv.ParseFloat(strings.TrimSpace(existingAmountSlice[i]), 64)
		if err != nil {
			return nil, fmt.Errorf("error on converting amount: %v to float", existingAmountSlice[i])
		}
		invoiceAdjustmentID, ok := stepState.InvoiceAdjustMapAmount[existingAmountSlice[i]]
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error cannot find invoice adjustment id on amount: %v", existingAmountParse)
		}
		// deduct the existing amount and just add the updated amount info
		req.InvoiceSubTotal -= existingAmountParse
		req.InvoiceTotal -= existingAmountParse
		invoiceAdjustmentDetail := &invoice_pb.InvoiceAdjustmentDetail{
			InvoiceAdjustmentId:     invoiceAdjustmentID,
			InvoiceAdjustmentAction: invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT,
		}

		req.InvoiceAdjustmentDetails = append(req.InvoiceAdjustmentDetails, invoiceAdjustmentDetail)
	}

	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) generateUpsertInvoiceAdjustmentRequest(ctx context.Context) *invoice_pb.UpsertInvoiceAdjustmentsRequest {
	stepState := StepStateFromContext(ctx)
	var req *invoice_pb.UpsertInvoiceAdjustmentsRequest
	req = &invoice_pb.UpsertInvoiceAdjustmentsRequest{
		InvoiceId:       stepState.InvoiceID,
		InvoiceSubTotal: stepState.InvoiceTotalAmount[0],
		InvoiceTotal:    stepState.InvoiceTotalAmount[0],
	}

	if stepState.Request != nil && stepState.Request.(*invoice_pb.UpsertInvoiceAdjustmentsRequest) != nil {
		req = stepState.Request.(*invoice_pb.UpsertInvoiceAdjustmentsRequest)
	}

	return req
}
