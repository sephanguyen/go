package paymentsvc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	seqnumberservice "github.com/manabie-com/backend/internal/invoicemgmt/services/sequence_number"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *PaymentModifierService) AddInvoicePayment(ctx context.Context, req *invoice_pb.AddInvoicePaymentRequest) (*invoice_pb.AddInvoicePaymentResponse, error) {
	invoice, err := s.validateAddInvoicePayment(ctx, req)
	if err != nil {
		return nil, err
	}

	paymentEntity, err := GenPaymentFromAddPaymentRequest(req, invoice)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Store the payment entity and action log to the DB
	if err := s.addPayment(ctx, req, paymentEntity); err != nil {
		return nil, err
	}

	return &invoice_pb.AddInvoicePaymentResponse{
		Successful: true,
	}, nil
}

func (s *PaymentModifierService) validateAddInvoicePayment(ctx context.Context, req *invoice_pb.AddInvoicePaymentRequest) (*entities.Invoice, error) {
	if err := validateAddInvoicePaymentRequest(req); err != nil {
		return nil, err
	}

	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.DB, req.InvoiceId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("InvoiceRepo.RetrieveInvoiceByInvoiceID err: %v", err))
	}

	// Check if invoice is ISSUED
	if invoice.Status.String != invoice_pb.InvoiceStatus_ISSUED.String() {
		return nil, status.Error(codes.InvalidArgument, "invoice status should be ISSUED")
	}

	exactTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Check if invoice total is positive
	if exactTotal <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invoice total cannot be less than or equal to 0")
	}

	exactOutstandingBalance, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.OutstandingBalance, "2")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// This is a temporary validation until partial payment is implemented
	// Check if the amount is equal to the invoice's outstanding balance
	if exactOutstandingBalance != req.Amount {
		return nil, status.Error(codes.InvalidArgument, "the given amount should be equal to the invoice outstanding balance")
	}

	// Check the latest payment status of the invoice
	if err := s.validateLatestPaymentStatus(ctx, invoice.InvoiceID.String); err != nil {
		return nil, err
	}

	if req.PaymentMethod.String() == invoice_pb.PaymentMethod_DIRECT_DEBIT.String() {
		// Check if student's bank detail is verified
		if err := s.checkStudentBankAccountVerification(ctx, invoice.StudentID.String); err != nil {
			return nil, err
		}
	}

	return invoice, nil
}

func (s *PaymentModifierService) addPayment(ctx context.Context, req *invoice_pb.AddInvoicePaymentRequest, paymentEntity *entities.Payment) error {
	// Retry when there is a duplicate error. Most likely from payment sequence number
	// Set maxRetry to 50 since there are some instance that it exceeded 10 retries
	enablePaymentSequenceNumberManualSetting, err := s.UnleashClient.IsFeatureEnabled(constant.EnablePaymentSequenceNumberManualSetting, s.Env)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("%v unleashClient.IsFeatureEnabled err: %v", constant.EnablePaymentSequenceNumberManualSetting, err))
	}

	maxRetry := 100
	err = utils.DoWithMaxRetry(func(attempt int) (bool, error) {
		err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			// store payment to the DB

			// Set the sequence number of payment manually
			paymentSeqNumberService := s.SequenceNumberService.GetPaymentSequenceNumberService()
			if enablePaymentSequenceNumberManualSetting {
				err := paymentSeqNumberService.InitLatestSeqNumber(ctx, tx)
				if err != nil {
					return status.Error(codes.Internal, fmt.Sprintf("paymentSeqNumberService.InitLatestSeqNumber err: %v", err))
				}

				err = paymentSeqNumberService.AssignSeqNumberToPayment(paymentEntity)
				if err != nil {
					return status.Error(codes.Internal, err.Error())
				}
			}

			if err := s.PaymentRepo.Create(ctx, tx, paymentEntity); err != nil {
				return err
			}

			createdPayment, err := s.PaymentRepo.GetLatestPaymentDueDateByInvoiceID(ctx, tx, req.InvoiceId)
			if err != nil {
				return err
			}

			// generate action logs details
			actionDetails := &utils.InvoiceActionLogDetails{
				InvoiceID:             createdPayment.InvoiceID.String,
				Action:                invoice_pb.InvoiceAction_PAYMENT_ADDED,
				PaymentSequenceNumber: createdPayment.PaymentSequenceNumber.Int,
				PaymentMethod:         createdPayment.PaymentMethod.String,
				ActionComment:         req.Remarks,
			}
			if err := utils.CreateActionLogV2(ctx, tx, actionDetails, s.InvoiceActionLogRepo); err != nil {
				return err
			}

			return nil
		})

		if err == nil {
			return false, nil
		}

		if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") && !strings.Contains(err.Error(), seqnumberservice.PaymentSeqNumberLockAcquiredErr) {
			return false, err
		}

		log.Printf("Retrying adding payment to invoice %v. Attempt: %d \n", req.InvoiceId, attempt)
		time.Sleep(100 * time.Millisecond)
		return attempt < maxRetry, fmt.Errorf("cannot add payment: %v", err)
	}, maxRetry)

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (s *PaymentModifierService) checkStudentBankAccountVerification(ctx context.Context, studentID string) error {
	bankAccount, err := s.BankAccountRepo.FindByStudentID(ctx, s.DB, studentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return status.Error(codes.InvalidArgument, "student should have verified bank account if the payment method is DIRECT DEBIT")
		}
		return status.Error(codes.Internal, fmt.Sprintf("BankAccountRepo.FindByStudentID err: %v", err))
	}

	if !bankAccount.IsVerified.Bool {
		return status.Error(codes.InvalidArgument, "bank account should be verified if the payment method is DIRECT DEBIT")
	}

	return nil
}

func validateAddInvoicePaymentRequest(req *invoice_pb.AddInvoicePaymentRequest) error {
	if strings.TrimSpace(req.InvoiceId) == "" {
		return status.Error(codes.InvalidArgument, "invoice ID cannot be empty")
	}

	if req.Amount <= 0 {
		return status.Error(codes.InvalidArgument, "amount cannot be less than or equal to 0")
	}

	_, allowed := constant.AddInvoicePaymentAllowedMethods[req.PaymentMethod.String()]
	if !allowed {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("payment method %v is not allowed", req.PaymentMethod.String()))
	}

	if err := utils.ValidateDueDateAndExpiryDate(req.DueDate, req.ExpiryDate); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	return nil
}
