package invoicesvc

import (
	"context"

	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InvoiceModifierService) RetrieveInvoiceInfo(ctx context.Context, req *invoice_pb.RetrieveInvoiceInfoRequest) (*invoice_pb.RetrieveInvoiceInfoResponse, error) {
	failedResp := &invoice_pb.RetrieveInvoiceInfoResponse{
		Successful: false,
	}

	successfulResp := &invoice_pb.RetrieveInvoiceInfoResponse{
		Successful: true,
	}

	// Retrieve invoice record
	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.DB, req.InvoiceIdString)

	if err != nil {
		return failedResp, status.Error(codes.Internal, err.Error())
	}

	// Only allow invoice with status other than DRAFT
	if invoice.Status.String == invoice_pb.InvoiceStatus_DRAFT.String() {
		return failedResp, status.Error(codes.InvalidArgument, "invoice should not be in DRAFT status")
	}

	// Retrieve payment record
	payment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, s.DB, req.InvoiceIdString)

	// Allow even if there's no payment record
	if err != nil && err != pgx.ErrNoRows {
		return failedResp, status.Error(codes.Internal, err.Error())
	}

	// Convert numeric into float64

	totalFloat, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	subTotalFloat, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.SubTotal, "2")

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Retrieve invoice bill items
	invoiceBillItems, err := s.InvoiceBillItemRepo.FindAllByInvoiceID(ctx, s.DB, req.InvoiceIdString)

	if err != nil {
		return failedResp, status.Error(codes.Internal, err.Error())
	}

	invoiceBillItemsArr := invoiceBillItems.ToArray()

	billItemsResult := make([]*invoice_pb.RetrieveInvoiceInfoBillItem, 0)

	// Retrieve the bill item record for each invoice bill item
	for _, invoiceBillItem := range invoiceBillItemsArr {
		billItem, err := s.BillItemRepo.FindByID(ctx, s.DB, invoiceBillItem.BillItemSequenceNumber.Int)

		if err != nil {
			return failedResp, status.Error(codes.Internal, err.Error())
		}

		// Set value to zero if no value assigned (nullable table fields)
		if billItem.DiscountAmountValue.Int == nil {
			billItem.DiscountAmountValue.Set(0)
		}

		if billItem.TaxAmount.Int == nil {
			billItem.TaxAmount.Set(0)
		}

		// Convert numeric into float64
		discountAmountValue, err := utils.GetFloat64ExactValueAndDecimalPlaces(billItem.DiscountAmountValue, "2")

		if err != nil {
			return failedResp, status.Error(codes.Internal, err.Error())
		}

		discountAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(billItem.DiscountAmount, "2")

		if err != nil {
			return failedResp, status.Error(codes.Internal, err.Error())
		}

		taxAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(billItem.TaxAmount, "2")

		if err != nil {
			return failedResp, status.Error(codes.Internal, err.Error())
		}

		finalPrice, err := utils.GetFloat64ExactValueAndDecimalPlaces(billItem.FinalPrice, "2")

		if err != nil {
			return failedResp, status.Error(codes.Internal, err.Error())
		}

		billItemResult := &invoice_pb.RetrieveInvoiceInfoBillItem{
			BillItemId:          billItem.BillItemSequenceNumber.Int,
			Description:         billItem.ProductDescription.String,
			DiscountAmountType:  billItem.DiscountAmountType.String,
			DiscountAmountValue: discountAmountValue,
			DiscountAmount:      discountAmount,
			TaxPercentage:       billItem.TaxPercentage.Int,
			TaxAmount:           taxAmount,
			Amount:              finalPrice + taxAmount,
		}

		if err != nil {
			return failedResp, status.Error(codes.Internal, err.Error())
		}

		billItemsResult = append(billItemsResult, billItemResult)
	}

	// Assign values to the response
	successfulResp.Total = totalFloat
	successfulResp.SubTotal = subTotalFloat
	successfulResp.CreatedDate = timestamppb.New(invoice.CreatedAt.Time)
	successfulResp.Status = invoice_pb.InvoiceStatus(invoice_pb.InvoiceStatus_value[invoice.Status.String])
	successfulResp.BillItems = billItemsResult

	successfulResp.DueDate = nil

	if payment != nil {
		successfulResp.DueDate = timestamppb.New(payment.PaymentDueDate.Time)
	}

	return successfulResp, err
}
