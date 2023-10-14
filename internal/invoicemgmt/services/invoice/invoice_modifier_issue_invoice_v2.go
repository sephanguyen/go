package invoicesvc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	payment_service "github.com/manabie-com/backend/internal/invoicemgmt/services/payment"
	seqnumberservice "github.com/manabie-com/backend/internal/invoicemgmt/services/sequence_number"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InvoiceModifierService) IssueInvoiceV2(ctx context.Context, req *invoice_pb.IssueInvoiceRequestV2) (*invoice_pb.IssueInvoiceResponseV2, error) {
	enableSingleIssueInvoiceWithPayment, err := s.UnleashClient.IsFeatureEnabled(constant.EnableSingleIssueInvoiceWithPayment, s.Env)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnableSingleIssueInvoiceWithPayment, err))
	}

	if enableSingleIssueInvoiceWithPayment {
		return s.IssueInvoiceV2WithPayment(ctx, req)
	}

	if strings.TrimSpace(req.InvoiceId) == "" {
		return nil, status.Error(codes.InvalidArgument, "invoice ID cannot be empty")
	}

	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.DB, req.InvoiceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if invoice.Status.String != invoice_pb.InvoiceStatus_DRAFT.String() && invoice.Status.String != invoice_pb.InvoiceStatus_FAILED.String() {
		return nil, status.Error(codes.InvalidArgument, "invalid invoice status")
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// If the invoice total is 0, update the status to PAID
		var invoiceStatus invoice_pb.InvoiceStatus
		switch {
		case invoice.Total.Int.Cmp(big.NewInt(0)) == 0:
			invoiceStatus = invoice_pb.InvoiceStatus_PAID
		default:
			invoiceStatus = invoice_pb.InvoiceStatus_ISSUED
		}

		// update invoice status
		if err := s.updateInvoiceStatus(ctx, tx, invoice.InvoiceID.String, invoiceStatus.String()); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// generate action logs details
		actionDetails := &utils.InvoiceActionLogDetails{
			InvoiceID:     req.InvoiceId,
			Action:        invoice_pb.InvoiceAction_INVOICE_ISSUED,
			ActionComment: req.Remarks,
		}
		if err := utils.CreateActionLogV2(ctx, tx, actionDetails, s.InvoiceActionLogRepo); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &invoice_pb.IssueInvoiceResponseV2{
		Successful: true,
	}, nil
}

func (s *InvoiceModifierService) IssueInvoiceV2WithPayment(ctx context.Context, req *invoice_pb.IssueInvoiceRequestV2) (*invoice_pb.IssueInvoiceResponseV2, error) {
	if strings.TrimSpace(req.InvoiceId) == "" {
		return nil, status.Error(codes.InvalidArgument, "invoice ID cannot be empty")
	}

	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.DB, req.InvoiceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.validateInvoiceIssueWithPayment(ctx, invoice, req)
	if err != nil {
		return nil, err
	}

	// Retry when there is a duplicate error. Most likely from payment sequence number
	err = utils.DoWithMaxRetry(func(attempt int) (bool, error) {
		err = s.issueInvoiceWithPayment(ctx, invoice, req)
		if err == nil {
			return false, nil
		}

		if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") && !strings.Contains(err.Error(), seqnumberservice.PaymentSeqNumberLockAcquiredErr) {
			return false, err
		}

		log.Printf("Retrying issuing single invoice. Attempt: %d \n", attempt)
		time.Sleep(100 * time.Millisecond)
		return attempt < issueInvoiceMaxRetry, fmt.Errorf("cannot issue invoice: %v", err)
	}, issueInvoiceMaxRetry)

	if err != nil {
		return nil, err
	}

	return &invoice_pb.IssueInvoiceResponseV2{
		Successful: true,
	}, nil
}

func isInvoiceAmountGreaterThanZero(invoice *entities.Invoice) (bool, error) {
	exactTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
	if err != nil {
		return false, err
	}

	return exactTotal > 0, nil
}

func (s *InvoiceModifierService) issueInvoiceWithPayment(ctx context.Context, invoice *entities.Invoice, req *invoice_pb.IssueInvoiceRequestV2) error {
	enablePaymentSequenceNumberManualSetting, err := s.UnleashClient.IsFeatureEnabled(constant.EnablePaymentSequenceNumberManualSetting, s.Env)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnablePaymentSequenceNumberManualSetting, err))
	}

	invoiceTotalGreaterZero, err := isInvoiceAmountGreaterThanZero(invoice)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		paymentSeqNumberService := s.SequenceNumberService.GetPaymentSequenceNumberService()
		if enablePaymentSequenceNumberManualSetting {
			err := paymentSeqNumberService.InitLatestSeqNumber(ctx, tx)
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("paymentSeqNumberService.InitLatestSeqNumber err: %v", err))
			}
		}

		// If the invoice total is 0, update the status to PAID
		invoiceStatus := invoice_pb.InvoiceStatus_ISSUED
		if invoice.Total.Int.Cmp(big.NewInt(0)) == 0 {
			invoiceStatus = invoice_pb.InvoiceStatus_PAID
		}

		// update invoice status
		if err := s.updateInvoiceStatus(ctx, tx, invoice.InvoiceID.String, invoiceStatus.String()); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// generate action logs details for issuing the invoice
		actionDetails := &utils.InvoiceActionLogDetails{
			InvoiceID:     req.InvoiceId,
			Action:        invoice_pb.InvoiceAction_INVOICE_ISSUED,
			ActionComment: req.Remarks,
		}
		if err := utils.CreateActionLogV2(ctx, tx, actionDetails, s.InvoiceActionLogRepo); err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// Create payment and action log for invoice total that is greater than zero
		if invoiceTotalGreaterZero {
			addPaymentReq := &invoice_pb.AddInvoicePaymentRequest{
				InvoiceId:     invoice.InvoiceID.String,
				PaymentMethod: req.PaymentMethod,
				Amount:        req.Amount,
				DueDate:       req.DueDate,
				ExpiryDate:    req.ExpiryDate,
			}

			paymentEntity, err := payment_service.GenPaymentFromAddPaymentRequest(addPaymentReq, invoice)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			if enablePaymentSequenceNumberManualSetting {
				err = paymentSeqNumberService.AssignSeqNumberToPayment(paymentEntity)
				if err != nil {
					return status.Error(codes.Internal, err.Error())
				}
			}

			err = s.PaymentRepo.Create(ctx, tx, paymentEntity)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			createdPayment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, tx, invoice.InvoiceID.String)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			// generate action logs details
			paymentActionDetails := &utils.InvoiceActionLogDetails{
				InvoiceID:             req.InvoiceId,
				Action:                invoice_pb.InvoiceAction_PAYMENT_ADDED,
				ActionComment:         "",
				PaymentMethod:         createdPayment.PaymentMethod.String,
				PaymentSequenceNumber: createdPayment.PaymentSequenceNumber.Int,
			}
			if err := utils.CreateActionLogV2(ctx, tx, paymentActionDetails, s.InvoiceActionLogRepo); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}

		return nil
	})
}

func (s *InvoiceModifierService) validateInvoiceIssueWithPayment(ctx context.Context, invoice *entities.Invoice, req *invoice_pb.IssueInvoiceRequestV2) error {
	if invoice.Status.String != invoice_pb.InvoiceStatus_DRAFT.String() && invoice.Status.String != invoice_pb.InvoiceStatus_FAILED.String() {
		return status.Error(codes.InvalidArgument, "invalid invoice status")
	}

	invoiceTotalGreaterZero, err := isInvoiceAmountGreaterThanZero(invoice)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// Validate other payment fields if invoice total is greater than zero
	if invoiceTotalGreaterZero {
		if _, ok := invoice_pb.PaymentMethod_value[req.PaymentMethod.String()]; !ok {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("invalid %v issue payment method value", req.PaymentMethod.String()))
		}

		err := utils.ValidateDueDateAndExpiryDate(req.DueDate, req.ExpiryDate)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		// Validate bank account verification
		if req.PaymentMethod == invoice_pb.PaymentMethod_DIRECT_DEBIT {
			err = s.checkBankAccountVerification(ctx, invoice.StudentID.String)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *InvoiceModifierService) checkBankAccountVerification(ctx context.Context, studentID string) error {
	bankAccount, err := s.BankAccountRepo.FindByStudentID(ctx, s.DB, studentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return status.Error(codes.InvalidArgument, "student has no bank account registered")
		}
		return status.Error(codes.Internal, fmt.Sprintf("s.BankAccountRepo.FindByStudentID err: %v", err))
	}

	if !bankAccount.IsVerified.Bool {
		return status.Error(codes.InvalidArgument, "student bank account is not verified")
	}

	return nil
}
