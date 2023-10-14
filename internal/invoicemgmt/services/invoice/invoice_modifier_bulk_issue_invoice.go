package invoicesvc

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Set maxRetry to 50 since there are some instance that it exceeded 10 retries
const bulkIssueInvoiceMaxRetry = 50

func (s *InvoiceModifierService) BulkIssueInvoice(ctx context.Context, req *invoice_pb.BulkIssueInvoiceRequest) (*invoice_pb.BulkIssueInvoiceResponse, error) {
	// input data validation
	if err := s.validateBulkIssueInvoiceReq(req); err != nil {
		return &invoice_pb.BulkIssueInvoiceResponse{
			Success: false,
		}, err
	}

	// Retry when there is a duplicate error. Most likely from payment sequence number
	err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
		err := s.bulkIssueInvoice(ctx, req)

		if err == nil {
			return false, nil
		}

		// Check if not duplicate constraint error
		if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			return false, err
		}

		log.Printf("Retrying bulk issuing invoice. Attempt: %d \n", attempt)
		return attempt < bulkIssueInvoiceMaxRetry, fmt.Errorf("cannot bulk issue invoice: %v", err)
	}, bulkIssueInvoiceMaxRetry)

	if err != nil {
		return &invoice_pb.BulkIssueInvoiceResponse{
			Success: false,
		}, err
	}

	return &invoice_pb.BulkIssueInvoiceResponse{
		Success: true,
	}, nil
}

func (s *InvoiceModifierService) bulkIssueInvoice(ctx context.Context, req *invoice_pb.BulkIssueInvoiceRequest) error {
	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for _, invoiceId := range req.InvoiceIds {
			invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, tx, invoiceId)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("error Invoice RetrieveInvoiceByInvoiceID: %v", err))
			}

			if invoice.Status.String != invoice_pb.InvoiceStatus_DRAFT.String() && invoice.Status.String != invoice_pb.InvoiceStatus_FAILED.String() {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("error Wrong Invoice Status"))
			}

			if invoice.Total.Int.Cmp(big.NewInt(0)) == -1 {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("error Should have positive total, negative total found"))
			}
			var paymentMethod invoice_pb.PaymentMethod
			switch req.BulkIssuePaymentMethod {
			case invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE:
				paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE
			case invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT:
				// check default student payment method
				studentPaymentDetail, err := s.StudentPaymentDetailRepo.FindByStudentID(ctx, tx, invoice.StudentID.String)
				if err != nil {
					return status.Error(codes.Internal, fmt.Sprintf("error StudentPaymentDetail FindByStudentID: %v", err))
				}
				paymentMethod = constant.PaymentMethodsConvertToEnums[studentPaymentDetail.PaymentMethod.String]
			}

			request := &invoice_pb.IssueInvoiceRequest{
				InvoiceIdString: invoiceId,
				PaymentMethod:   paymentMethod,
			}

			switch paymentMethod {
			case invoice_pb.PaymentMethod_CONVENIENCE_STORE:
				request.DueDate = req.ConvenienceStoreDates.DueDate
				request.ExpiryDate = req.ConvenienceStoreDates.ExpiryDate
			case invoice_pb.PaymentMethod_DIRECT_DEBIT:
				request.DueDate = req.DirectDebitDates.DueDate
				request.ExpiryDate = req.DirectDebitDates.ExpiryDate
			}

			paymentEntity, err := generatePaymentData(request, invoice)
			if err != nil {
				return err
			}

			if err := s.PaymentRepo.Create(ctx, tx, paymentEntity); err != nil {
				return err
			}

			if err := s.updateInvoiceStatusAndExportedTag(ctx, tx, invoiceId, invoice_pb.InvoiceStatus_ISSUED.String(), false); err != nil {
				return err
			}

			payment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, tx, invoiceId)
			if err != nil {
				return err
			}
			// generate action logs details
			actionDetails := &InvoiceActionLogDetails{
				InvoiceID:             invoiceId,
				Action:                invoice_pb.InvoiceAction_INVOICE_ISSUED,
				ActionComment:         "",
				PaymentSequenceNumber: payment.PaymentSequenceNumber.Int,
				PaymentMethod:         payment.PaymentMethod.String,
			}
			if err := s.createActionLog(ctx, tx, actionDetails); err != nil {

				return err
			}
		}

		return nil
	})

	return err
}

func (s *InvoiceModifierService) validateBulkIssueInvoiceReq(req *invoice_pb.BulkIssueInvoiceRequest) error {
	if len(req.InvoiceIds) == 0 {
		return status.Error(codes.InvalidArgument, "empty InvoiceIds")
	}

	if _, ok := invoice_pb.BulkIssuePaymentMethod_value[req.BulkIssuePaymentMethod.String()]; !ok {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid %v BulkIssuePaymentMethod value", req.BulkIssuePaymentMethod.String()))
	}

	if req.ConvenienceStoreDates == nil {
		return status.Error(codes.InvalidArgument, "convenience store dates cannot be empty")
	}

	// validate convenience store dates always visible on both payment methods
	err := utils.ValidateDueDateAndExpiryDate(req.ConvenienceStoreDates.DueDate, req.ConvenienceStoreDates.ExpiryDate)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	// check if payment method is default payment
	if req.BulkIssuePaymentMethod == invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT {
		if req.DirectDebitDates == nil {
			return status.Error(codes.InvalidArgument, "direct debit dates cannot be empty")
		}
		// validate direct debit dates
		err := utils.ValidateDueDateAndExpiryDate(req.DirectDebitDates.DueDate, req.DirectDebitDates.ExpiryDate)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}
	return nil
}
