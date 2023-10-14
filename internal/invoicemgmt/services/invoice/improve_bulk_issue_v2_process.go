package invoicesvc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	payment_service "github.com/manabie-com/backend/internal/invoicemgmt/services/payment"
	seqnumberservice "github.com/manabie-com/backend/internal/invoicemgmt/services/sequence_number"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InvoiceModifierService) improvedProcessBulkIssuingWithPayments(ctx context.Context, db database.QueryExecer, bulkPaymentEntity *entities.BulkPayment, req *invoice_pb.BulkIssueInvoiceRequestV2) error {
	enablePaymentSequenceNumberManualSetting, err := s.UnleashClient.IsFeatureEnabled(constant.EnablePaymentSequenceNumberManualSetting, s.Env)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnablePaymentSequenceNumberManualSetting, err))
	}

	paymentSeqNumberService, err := s.retrievePaymentSequenceNumberService(ctx, enablePaymentSequenceNumberManualSetting, db)
	if err != nil {
		return err
	}

	if err := s.BulkPaymentRepo.Create(ctx, db, bulkPaymentEntity); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("error BulkPaymentRepo Create: %v", err))
	}

	// Validate invoices and retrieve student IDs
	invoices, studentIDs, err := s.validateInvoicesAndRetrieveStudents(ctx, db, req)
	if err != nil {
		return err
	}

	// Manage student payment methods
	studentPaymentMethodMap, err := s.manageStudentPaymentMethods(ctx, db, req.BulkIssuePaymentMethod, studentIDs)
	if err != nil {
		return err
	}

	// Update invoice status of invoices that exist in temp table
	err = s.InvoiceRepo.UpdateStatusFromInvoiceIDTempTable(ctx, db, invoice_pb.InvoiceStatus_ISSUED.String())
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("s.InvoiceRepo.UpdateStatusFromInvoiceIDTempTable err: %v", err))
	}

	// Generate payment entities
	paymentsToBeCreated, paymentIDs, err := genPaymentsFromBulkIssueInvoice(invoices, studentPaymentMethodMap, bulkPaymentEntity.BulkPaymentID, req)
	if err != nil {
		return err
	}

	if enablePaymentSequenceNumberManualSetting {
		err = paymentSeqNumberService.AssignSeqNumberToPayments(paymentsToBeCreated)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	// Create payments in batch
	err = s.PaymentRepo.CreateMultiple(ctx, db, paymentsToBeCreated)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("s.PaymentRepo.CreateMultiple err: %v", err))
	}

	// Get the payments by the created payment ID for the creation of action log if payment sequence number is not set manually
	createdPayments := paymentsToBeCreated
	if !enablePaymentSequenceNumberManualSetting {
		payments, err := s.PaymentRepo.FindByPaymentIDs(ctx, db, paymentIDs)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("s.PaymentRepo.FindByPaymentIDs err: %v", err))
		}

		if len(payments) != len(req.InvoiceIds) {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("there are %d payments that were not created", len(req.InvoiceIds)-len(payments)))
		}

		createdPayments = payments
	}

	// Generate list of action logs
	actionLogs, err := genBulkIssueActionLogsFromPayments(ctx, createdPayments)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// Create action log in bulk
	err = s.InvoiceActionLogRepo.CreateMultiple(ctx, db, actionLogs)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("s.InvoiceActionLogRepo.CreateMultiple err: %v", err))
	}

	return nil
}
func (s *InvoiceModifierService) retrievePaymentSequenceNumberService(ctx context.Context, enable bool, db database.QueryExecer) (seqnumberservice.IPaymentSequenceNumberService, error) {
	paymentSeqNumberService := s.SequenceNumberService.GetPaymentSequenceNumberService()
	if enable {
		err := paymentSeqNumberService.InitLatestSeqNumber(ctx, db)
		if err != nil {
			return paymentSeqNumberService, status.Error(codes.Internal, fmt.Sprintf("paymentSeqNumberService.InitLatestSeqNumber err: %v", err))
		}
	}

	return paymentSeqNumberService, nil
}

func (s *InvoiceModifierService) validateInvoicesAndRetrieveStudents(ctx context.Context, db database.QueryExecer, req *invoice_pb.BulkIssueInvoiceRequestV2) ([]*entities.Invoice, []string, error) {
	// Create temporary table to store the invoice IDs
	err := s.InvoiceRepo.InsertInvoiceIDsTempTable(ctx, db, req.InvoiceIds)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, fmt.Sprintf("s.InvoiceRepo.InsertInvoiceIDsTempTable err: %v", err))
	}

	// Find the invoices that exists in the temporary table
	invoices, err := s.InvoiceRepo.FindInvoicesFromInvoiceIDTempTable(ctx, db)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, fmt.Sprintf("s.InvoiceRepo.FindInvoicesFromInvoiceIDTempTable err: %v", err))
	}

	// Check if all given invoice IDs are existing
	if len(invoices) != len(req.InvoiceIds) {
		return nil, nil, status.Error(codes.InvalidArgument, fmt.Sprintf("there are %d invoices that does not exist", len(req.InvoiceIds)-len(invoices)))
	}

	// Validate invoices and get list of student IDs
	studentIDs := make([]string, len(invoices))
	for i, invoice := range invoices {
		if invoice.Status.String != invoice_pb.InvoiceStatus_DRAFT.String() {
			return nil, nil, status.Error(codes.InvalidArgument, fmt.Sprintf("error invalid invoice status: %v", invoice.Status.String))
		}

		err = validateInvoiceTotalToBulkIssue(invoice)
		if err != nil {
			return nil, nil, err
		}

		studentIDs[i] = invoice.StudentID.String
	}

	return invoices, studentIDs, nil
}

func (s *InvoiceModifierService) manageStudentPaymentMethods(ctx context.Context, db database.QueryExecer, paymentMethod invoice_pb.BulkIssuePaymentMethod, studentIDs []string) (map[string]string, error) {
	studentPaymentMethodMap := make(map[string]string)

	switch paymentMethod {
	case invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_CONVENIENCE_STORE:
		// Set all student payments' payment method to CONVENIENCE_STORE
		for _, studentID := range studentIDs {
			studentPaymentMethodMap[studentID] = invoice_pb.PaymentMethod_CONVENIENCE_STORE.String()
		}
	case invoice_pb.BulkIssuePaymentMethod_BULK_ISSUE_DEFAULT_PAYMENT:
		studentPaymentDetails, err := s.StudentPaymentDetailRepo.FindFromInvoiceIDTempTable(ctx, db)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("s.StudentPaymentDetailRepo.FindFromInvoiceIDTempTable err: %v", err))
		}

		if len(studentPaymentDetails) != len(studentIDs) {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("there are %d students that does not have student payment detail", len(studentIDs)-len(studentPaymentDetails)))
		}

		for _, spd := range studentPaymentDetails {
			if spd.PaymentMethod.String == "" {
				return nil, status.Error(codes.Internal, fmt.Sprintf("bulk issue student: %v payment method in student payment detail is empty", spd.StudentID.String))
			}

			studentPaymentMethodMap[spd.StudentID.String] = spd.PaymentMethod.String
		}
	default:
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid bulk issue payment method %v", paymentMethod))
	}

	return studentPaymentMethodMap, nil
}

func genPaymentsFromBulkIssueInvoice(invoices []*entities.Invoice, studentPaymentMethodMap map[string]string, bulkPaymentID pgtype.Text, req *invoice_pb.BulkIssueInvoiceRequestV2) ([]*entities.Payment, []string, error) {
	paymentIDs := make([]string, len(invoices))
	paymentEntities := make([]*entities.Payment, len(invoices))

	for i, invoice := range invoices {
		paymentMethodStr, ok := studentPaymentMethodMap[invoice.StudentID.String]
		if !ok {
			return nil, nil, status.Error(codes.Internal, fmt.Sprintf("bulk issue student: %v student has no student payment detail", invoice.StudentID.String))
		}
		paymentMethod := invoice_pb.PaymentMethod(invoice_pb.PaymentMethod_value[paymentMethodStr])

		invoiceTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
		if err != nil {
			return nil, nil, status.Error(codes.Internal, err.Error())
		}

		// create add payment request as common for reusing generating of payment
		addPaymentReq := getAddPaymentReqDatesFromBulkIssueReq(&invoice_pb.AddInvoicePaymentRequest{
			InvoiceId:     invoice.InvoiceID.String,
			PaymentMethod: paymentMethod,
			Amount:        invoiceTotal,
		}, req)

		paymentEntity, err := payment_service.GenPaymentFromAddPaymentRequest(addPaymentReq, invoice)
		if err != nil {
			return nil, nil, status.Error(codes.Internal, err.Error())
		}

		// populate the bulk payment id for grouping them in bulk
		paymentEntity.BulkPaymentID = bulkPaymentID
		paymentEntity.PaymentID = database.Text(idutil.ULIDNow())

		paymentIDs[i] = paymentEntity.PaymentID.String
		paymentEntities[i] = paymentEntity
	}

	return paymentEntities, paymentIDs, nil
}

func genBulkIssueActionLogsFromPayments(ctx context.Context, payments []*entities.Payment) ([]*entities.InvoiceActionLog, error) {
	actionLogs := make([]*entities.InvoiceActionLog, len(payments))
	index := 0
	for _, payment := range payments {
		actionDetails := &utils.InvoiceActionLogDetails{
			InvoiceID:             payment.InvoiceID.String,
			Action:                invoice_pb.InvoiceAction_INVOICE_BULK_ISSUED,
			PaymentSequenceNumber: payment.PaymentSequenceNumber.Int,
			PaymentMethod:         payment.PaymentMethod.String,
		}

		actionLogEntity, err := utils.GenActionLogEntity(ctx, actionDetails)
		if err != nil {
			return nil, err
		}

		actionLogs[index] = actionLogEntity

		index++
	}

	return actionLogs, nil
}
