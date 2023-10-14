package paymentsvc

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	seqnumberservice "github.com/manabie-com/backend/internal/invoicemgmt/services/sequence_number"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PaymentModifierService) BulkAddPayment(ctx context.Context, req *invoice_pb.BulkAddPaymentRequest) (*invoice_pb.BulkAddPaymentResponse, error) {
	if err := s.validateBulkAddPaymentReq(req); err != nil {
		return nil, err
	}

	// Retry when there is a duplicate error. Most likely from payment sequence number
	err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
		err := s.CreateBulkPayments(ctx, req)
		if err == nil {
			return false, nil
		}

		// Check if not duplicate constraint error
		if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") && !strings.Contains(err.Error(), seqnumberservice.PaymentSeqNumberLockAcquiredErr) {
			return false, err
		}

		log.Printf("Retrying bulk adding payment. Attempt: %d \n", attempt)
		time.Sleep(100 * time.Millisecond)
		return attempt < constant.BulkAddPaymentAndIssueMaxRetry, fmt.Errorf("cannot bulk add payment: %v", err)
	}, constant.BulkAddPaymentAndIssueMaxRetry)

	if err != nil {
		return nil, err
	}

	return &invoice_pb.BulkAddPaymentResponse{
		Successful: true,
	}, nil
}

func (s *PaymentModifierService) validateBulkAddPaymentReq(req *invoice_pb.BulkAddPaymentRequest) error {
	if len(req.InvoiceIds) == 0 {
		return status.Error(codes.InvalidArgument, "invoice ids cannot be empty")
	}

	if _, ok := invoice_pb.BulkPaymentMethod_value[req.BulkAddPaymentDetails.BulkPaymentMethod.String()]; !ok {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid %v BulkPaymentMethod value", req.BulkAddPaymentDetails.BulkPaymentMethod.String()))
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
	if req.BulkAddPaymentDetails.BulkPaymentMethod == invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT {
		if req.DirectDebitDates == nil {
			return status.Error(codes.InvalidArgument, "direct debit dates cannot be empty")
		}
		// validate direct debit dates
		err := utils.ValidateDueDateAndExpiryDate(req.DirectDebitDates.DueDate, req.DirectDebitDates.ExpiryDate)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}

	if len(req.BulkAddPaymentDetails.LatestPaymentStatus) == 0 {
		return status.Error(codes.InvalidArgument, "latest payment status cannot be empty")
	}

	for _, latestPaymentStatus := range req.BulkAddPaymentDetails.LatestPaymentStatus {
		if latestPaymentStatus != invoice_pb.PaymentStatus_PAYMENT_FAILED && latestPaymentStatus != invoice_pb.PaymentStatus_PAYMENT_NONE {
			return status.Error(codes.InvalidArgument, "latest payment status value should only have no payment or failed")
		}
	}

	if len(req.BulkAddPaymentDetails.InvoiceType) == 0 {
		return status.Error(codes.InvalidArgument, "invoice type cannot be empty")
	}

	for _, invoiceType := range req.BulkAddPaymentDetails.InvoiceType {
		if invoiceType != invoice_pb.InvoiceType_MANUAL && invoiceType != invoice_pb.InvoiceType_SCHEDULED {
			return status.Error(codes.InvalidArgument, "invoice type value should only have manual and scheduled")
		}
	}

	return nil
}

func (s *PaymentModifierService) CreateBulkPayments(ctx context.Context, req *invoice_pb.BulkAddPaymentRequest) error {
	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// create bulk payment entity to group the payments that will be created
		bulkPaymentEntity, err := genBulkPaymentFromRequest(req)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		if err := s.BulkPaymentRepo.Create(ctx, tx, bulkPaymentEntity); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error BulkPaymentRepo Create: %v", err))
		}

		return s.processBulkAddingPayments(ctx, tx, bulkPaymentEntity.BulkPaymentID, req)
	})

	return err
}

func (s *PaymentModifierService) processBulkAddingPayments(ctx context.Context, tx database.QueryExecer, bulkPaymentID pgtype.Text, req *invoice_pb.BulkAddPaymentRequest) error {
	enablePaymentSequenceNumberManualSetting, err := s.UnleashClient.IsFeatureEnabled(constant.EnablePaymentSequenceNumberManualSetting, s.Env)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnablePaymentSequenceNumberManualSetting, err))
	}

	paymentSeqNumberService := s.SequenceNumberService.GetPaymentSequenceNumberService()
	if enablePaymentSequenceNumberManualSetting {
		err := paymentSeqNumberService.InitLatestSeqNumber(ctx, tx)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("paymentSeqNumberService.InitLatestSeqNumber err: %v", err))
		}
	}

	for _, invoiceID := range req.InvoiceIds {
		invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, tx, invoiceID)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error InvoiceRepo RetrieveInvoiceByInvoiceID: %v", err))
		}
		// only issued invoice is valid for now as there will be no partial payment
		if invoice.Status.String != invoice_pb.InvoiceStatus_ISSUED.String() {
			return status.Error(codes.Internal, fmt.Sprintf("error invalid invoice status: %v", invoice.Status.String))
		}

		invoiceTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		paymentMethod, err := s.getStudentPaymentMethodFromBulkPaymentRequest(ctx, tx, req.BulkAddPaymentDetails.BulkPaymentMethod, invoice.StudentID.String)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// create add payment request as common for reusing generating of payment
		addPaymentReq := &invoice_pb.AddInvoicePaymentRequest{
			InvoiceId:     invoiceID,
			PaymentMethod: paymentMethod,
			Amount:        invoiceTotal,
		}

		addPaymentReq = getAddPaymentReqDatesFromPaymentMethod(addPaymentReq, req)

		paymentEntity, err := GenPaymentFromAddPaymentRequest(addPaymentReq, invoice)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
		// populate the bulk payment id for grouping them in bulk
		paymentEntity.BulkPaymentID = bulkPaymentID

		if enablePaymentSequenceNumberManualSetting {
			err = paymentSeqNumberService.AssignSeqNumberToPayment(paymentEntity)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}

		if err := s.PaymentRepo.Create(ctx, tx, paymentEntity); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error PaymentRepo Create: %v", err))
		}

		createdPayment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, tx, invoiceID)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error PaymentRepo GetLatestPaymentDueDateByInvoiceID: %v", err))
		}
		// generate action logs details
		actionDetails := &utils.InvoiceActionLogDetails{
			InvoiceID:             createdPayment.InvoiceID.String,
			Action:                invoice_pb.InvoiceAction_PAYMENT_ADDED,
			PaymentSequenceNumber: createdPayment.PaymentSequenceNumber.Int,
			PaymentMethod:         createdPayment.PaymentMethod.String,
		}
		if err := utils.CreateActionLogV2(ctx, tx, actionDetails, s.InvoiceActionLogRepo); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("error InvoiceActionLogRepo Create: %v", err))
		}
	}
	return nil
}

func (s *PaymentModifierService) getStudentPaymentMethodFromBulkPaymentRequest(ctx context.Context, db database.QueryExecer, bulkPaymentMethod invoice_pb.BulkPaymentMethod, studentID string) (invoice_pb.PaymentMethod, error) {
	var paymentMethod invoice_pb.PaymentMethod

	switch bulkPaymentMethod {
	case invoice_pb.BulkPaymentMethod_BULK_PAYMENT_CONVENIENCE_STORE:
		paymentMethod = invoice_pb.PaymentMethod_CONVENIENCE_STORE
	case invoice_pb.BulkPaymentMethod_BULK_PAYMENT_DEFAULT_PAYMENT:
		// check default student payment method
		studentPaymentDetail, err := s.StudentPaymentDetailRepo.FindByStudentID(ctx, db, studentID)
		if err != nil {
			return paymentMethod, fmt.Errorf("error StudentPaymentDetailRepo FindByStudentID: %v", err)
		}
		if strings.TrimSpace(studentPaymentDetail.PaymentMethod.String) == "" {
			return paymentMethod, fmt.Errorf("bulk add student: %v payment method in student payment detail is empty", studentID)
		}
		paymentMethod = constant.PaymentMethodsConvertToEnums[studentPaymentDetail.PaymentMethod.String]
	}
	return paymentMethod, nil
}

func genBulkPaymentFromRequest(req *invoice_pb.BulkAddPaymentRequest) (*entities.BulkPayment, error) {
	e := new(entities.BulkPayment)
	database.AllNullEntity(e)
	id := idutil.ULIDNow()
	if err := multierr.Combine(
		e.BulkPaymentID.Set(id),
		e.BulkPaymentStatus.Set(invoice_pb.BulkPaymentStatus_BULK_PAYMENT_PENDING),
		e.PaymentMethod.Set(req.BulkAddPaymentDetails.BulkPaymentMethod.String()),
		e.InvoiceStatus.Set(invoice_pb.InvoiceStatus_ISSUED),
		e.InvoiceType.Set(req.BulkAddPaymentDetails.InvoiceType),
		e.PaymentStatus.Set(req.BulkAddPaymentDetails.LatestPaymentStatus),
	); err != nil {
		return nil, fmt.Errorf("multierr.Combine: %w", err)
	}

	return e, nil
}

func getAddPaymentReqDatesFromPaymentMethod(addPaymentReq *invoice_pb.AddInvoicePaymentRequest, bulkPaymentReq *invoice_pb.BulkAddPaymentRequest) *invoice_pb.AddInvoicePaymentRequest {
	switch addPaymentReq.PaymentMethod {
	case invoice_pb.PaymentMethod_CONVENIENCE_STORE:
		addPaymentReq.DueDate = bulkPaymentReq.ConvenienceStoreDates.DueDate
		addPaymentReq.ExpiryDate = bulkPaymentReq.ConvenienceStoreDates.ExpiryDate
	case invoice_pb.PaymentMethod_DIRECT_DEBIT:
		addPaymentReq.DueDate = bulkPaymentReq.DirectDebitDates.DueDate
		addPaymentReq.ExpiryDate = bulkPaymentReq.DirectDebitDates.ExpiryDate
	}

	return addPaymentReq
}
