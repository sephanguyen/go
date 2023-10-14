package invoicesvc

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	payment_service "github.com/manabie-com/backend/internal/invoicemgmt/services/payment"
	seqnumberservice "github.com/manabie-com/backend/internal/invoicemgmt/services/sequence_number"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InvoiceModifierService) BulkIssueInvoiceV2(ctx context.Context, req *invoice_pb.BulkIssueInvoiceRequestV2) (*invoice_pb.BulkIssueInvoiceResponseV2, error) {
	if err := s.validateBulkIssueReqV2(req); err != nil {
		return nil, err
	}

	// Retry when there is a duplicate error. Most likely from payment sequence number
	err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
		err := s.CreateBulkIssuePayments(ctx, req)
		if err == nil {
			return false, nil
		}

		// Check if not duplicate constraint error
		if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") && !strings.Contains(err.Error(), seqnumberservice.PaymentSeqNumberLockAcquiredErr) {
			return false, err
		}

		log.Printf("Retrying bulk issue invoice. Attempt: %d \n", attempt)
		time.Sleep(100 * time.Millisecond)
		return attempt < constant.BulkAddPaymentAndIssueMaxRetry, fmt.Errorf("cannot bulk issue invoice: %v", err)
	}, constant.BulkAddPaymentAndIssueMaxRetry)

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &invoice_pb.BulkIssueInvoiceResponseV2{
		Success: true,
	}, nil
}

func (s *InvoiceModifierService) CreateBulkIssuePayments(ctx context.Context, req *invoice_pb.BulkIssueInvoiceRequestV2) error {
	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		bulkPaymentEntity, err := genBulkPaymentFromBulkIssue(req)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		enableBulkIssueInvoiceImprovement, err := s.UnleashClient.IsFeatureEnabled(constant.EnableImproveBulkIssueInvoice, s.Env)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableImproveBulkIssueInvoice, err))
		}

		switch enableBulkIssueInvoiceImprovement {
		case true:
			err = s.improvedProcessBulkIssuingWithPayments(ctx, tx, bulkPaymentEntity, req)
			log.Println("test", err)
		default:
			err = s.processBulkIssuingWithPayments(ctx, tx, bulkPaymentEntity, req)
		}

		return err
	})

	return err
}

func (s *InvoiceModifierService) processBulkIssuingWithPayments(ctx context.Context, tx database.QueryExecer, bulkPaymentEntity *entities.BulkPayment, req *invoice_pb.BulkIssueInvoiceRequestV2) error {
	if err := s.BulkPaymentRepo.Create(ctx, tx, bulkPaymentEntity); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("error BulkPaymentRepo Create: %v", err))
	}

	for _, invoiceID := range req.InvoiceIds {
		invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, tx, invoiceID)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error InvoiceRepo RetrieveInvoiceByInvoiceID: %v", err))
		}
		if invoice.Status.String != invoice_pb.InvoiceStatus_DRAFT.String() {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("error invalid invoice status: %v", invoice.Status.String))
		}

		err = validateInvoiceTotalToBulkIssue(invoice)
		if err != nil {
			return err
		}

		invoiceTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		paymentMethod, err := s.getStudentPaymentMethodFromBulkIssueRequest(ctx, tx, req.BulkIssuePaymentMethod, invoice.StudentID.String)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// create add payment request as common for reusing generating of payment
		addPaymentReq := &invoice_pb.AddInvoicePaymentRequest{
			InvoiceId:     invoiceID,
			PaymentMethod: paymentMethod,
			Amount:        invoiceTotal,
		}

		addPaymentReq = getAddPaymentReqDatesFromBulkIssueReq(addPaymentReq, req)

		paymentEntity, err := payment_service.GenPaymentFromAddPaymentRequest(addPaymentReq, invoice)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		// populate the bulk payment id for grouping them in bulk
		paymentEntity.BulkPaymentID = bulkPaymentEntity.BulkPaymentID

		if err := s.PaymentRepo.Create(ctx, tx, paymentEntity); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error PaymentRepo Create: %v", err))
		}

		createdPayment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, tx, invoiceID)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error PaymentRepo GetLatestPaymentDueDateByInvoiceID: %v", err))
		}

		if err := s.updateInvoiceStatus(ctx, tx, invoiceID, invoice_pb.InvoiceStatus_ISSUED.String()); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// generate action logs details
		actionDetails := &utils.InvoiceActionLogDetails{
			InvoiceID:             createdPayment.InvoiceID.String,
			Action:                invoice_pb.InvoiceAction_INVOICE_BULK_ISSUED,
			PaymentSequenceNumber: createdPayment.PaymentSequenceNumber.Int,
			PaymentMethod:         createdPayment.PaymentMethod.String,
		}
		if err := utils.CreateActionLogV2(ctx, tx, actionDetails, s.InvoiceActionLogRepo); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error InvoiceActionLogRepo Create: %v", err))
		}
	}
	return nil
}

func (s *InvoiceModifierService) getStudentPaymentMethodFromBulkIssueRequest(ctx context.Context, db database.QueryExecer, bulkIssuePaymentMethod invoice_pb.BulkIssuePaymentMethod, studentID string) (invoice_pb.PaymentMethod, error) {
	var paymentMethod invoice_pb.PaymentMethod

	switch bulkIssuePaymentMethod {
	case invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE:
		paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE
	case invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT:
		studentPaymentDetail, err := s.StudentPaymentDetailRepo.FindByStudentID(ctx, db, studentID)
		if err != nil {
			return paymentMethod, fmt.Errorf("error StudentPaymentDetailRepo FindByStudentID: %v", err)
		}

		if studentPaymentDetail.PaymentMethod.String == "" {
			return paymentMethod, fmt.Errorf("bulk issue student: %v payment method in student payment detail is empty", studentID)
		}

		paymentMethod = constant.PaymentMethodsConvertToEnums[studentPaymentDetail.PaymentMethod.String]
	}
	return paymentMethod, nil
}

func (s *InvoiceModifierService) validateBulkIssueReqV2(req *invoice_pb.BulkIssueInvoiceRequestV2) error {
	if len(req.InvoiceIds) == 0 {
		return status.Error(codes.InvalidArgument, "invoice ids cannot be empty")
	}

	if _, ok := invoice_pb.BulkIssuePaymentMethod_value[req.BulkIssuePaymentMethod.String()]; !ok {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid %v bulk issue payment method value", req.BulkIssuePaymentMethod.String()))
	}

	if req.ConvenienceStoreDates == nil {
		return status.Error(codes.InvalidArgument, "convenience store dates cannot be empty")
	}

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

	if len(req.InvoiceType) == 0 {
		return status.Error(codes.InvalidArgument, "invoice type cannot be empty")
	}

	for _, invoiceType := range req.InvoiceType {
		if invoiceType != invoice_pb.InvoiceType_MANUAL && invoiceType != invoice_pb.InvoiceType_SCHEDULED {
			return status.Error(codes.InvalidArgument, "invoice type value should only have manual and scheduled")
		}
	}

	return nil
}

func genBulkPaymentFromBulkIssue(req *invoice_pb.BulkIssueInvoiceRequestV2) (*entities.BulkPayment, error) {
	var paymentMethod invoice_pb.BulkPaymentMethod

	switch req.BulkIssuePaymentMethod {
	case invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE:
		paymentMethod = invoice_pb.BulkPaymentMethod_BULK_PAYMENT_CONVENIENCE_STORE
	case invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT:
		paymentMethod = invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT
	default:
		return nil, fmt.Errorf("invalid payment method from bulk issue: %v", req.BulkIssuePaymentMethod.String())
	}
	e := new(entities.BulkPayment)
	database.AllNullEntity(e)
	id := idutil.ULIDNow()
	if err := multierr.Combine(
		e.BulkPaymentID.Set(id),
		e.BulkPaymentStatus.Set(invoice_pb.BulkPaymentStatus_BULK_PAYMENT_PENDING),
		e.PaymentMethod.Set(paymentMethod),
		e.InvoiceStatus.Set(invoice_pb.InvoiceStatus_ISSUED),
		e.InvoiceType.Set(req.InvoiceType),
		e.PaymentStatus.Set(invoice_pb.PaymentStatus_PAYMENT_NONE.String()),
	); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	return e, nil
}

func getAddPaymentReqDatesFromBulkIssueReq(addPaymentReq *invoice_pb.AddInvoicePaymentRequest, bulkIssueReq *invoice_pb.BulkIssueInvoiceRequestV2) *invoice_pb.AddInvoicePaymentRequest {
	switch addPaymentReq.PaymentMethod {
	case invoice_pb.PaymentMethod_CONVENIENCE_STORE:
		addPaymentReq.DueDate = bulkIssueReq.ConvenienceStoreDates.DueDate
		addPaymentReq.ExpiryDate = bulkIssueReq.ConvenienceStoreDates.ExpiryDate
	case invoice_pb.PaymentMethod_DIRECT_DEBIT:
		addPaymentReq.DueDate = bulkIssueReq.DirectDebitDates.DueDate
		addPaymentReq.ExpiryDate = bulkIssueReq.DirectDebitDates.ExpiryDate
	}

	return addPaymentReq
}

func validateInvoiceTotalToBulkIssue(invoice *entities.Invoice) error {
	// Check if invoice total is negative amount
	if invoice.Total.Int.Cmp(big.NewInt(0)) == -1 {
		return status.Error(codes.InvalidArgument, "error Should have positive total, negative total found")
	}

	// Check if invoice total is zero amount
	if invoice.Total.Int.Cmp(big.NewInt(0)) == 0 {
		return status.Error(codes.InvalidArgument, "error Should have positive total, zero total amount found")
	}

	return nil
}
