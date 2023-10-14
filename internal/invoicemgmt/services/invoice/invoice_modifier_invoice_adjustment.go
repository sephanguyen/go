package invoicesvc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InvoiceAdjustmentData struct {
	InvoiceAdjustmentToUpsert []*entities.InvoiceAdjustment
	InvoiceAdjustmentToDelete pgtype.TextArray
	InvoiceTotalAmount        float64
}

// Set maxRetry to 20 since this does not usually needs to retry
const invoiceAdjustmentMaxRetry = 20

func (s *InvoiceModifierService) UpsertInvoiceAdjustments(ctx context.Context, req *invoice_pb.UpsertInvoiceAdjustmentsRequest) (*invoice_pb.UpsertInvoiceAdjustmentsResponse, error) {
	// validate invoice adjustment requests
	err := validateInvoiceAdjustmentRequest(req)
	if err != nil {
		return nil, err
	}
	// validate first invoice adjustment details before generating entities
	err = validateInvoiceAdjustmentDetails(req.InvoiceAdjustmentDetails)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// find invoice to check if it's in Draft status and to calculate total and subtotal
	invoice, err := s.InvoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.DB, req.InvoiceId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error Invoice RetrieveInvoiceByInvoiceID: %v", err))
	}

	if invoice.Status.String != invoice_pb.InvoiceStatus_DRAFT.String() {
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("error invoice status: %v should be in draft", invoice.Status.String))
	}

	invoiceAdjustmentData, err := s.generateInvoiceAdjustment(ctx, req.InvoiceAdjustmentDetails, invoice)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// get the invoice total from the existing record
	getExactInvoiceTotalValue, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
	if err != nil {
		return nil, err
	}
	// validation of invoice adjustment detail amount is always returning 0 even the field is not existing in protobuf so we validate it only here
	// compare the total should be equal to the total of invoice + the adjustment amounts
	expectedInvoiceTotalAmount := getExactInvoiceTotalValue + invoiceAdjustmentData.InvoiceTotalAmount
	if req.InvoiceTotal != expectedInvoiceTotalAmount {
		return nil, status.Error(codes.Internal, fmt.Sprintf("expected invoice total amount %v received %v", expectedInvoiceTotalAmount, req.InvoiceTotal))
	}

	// get the invoice subtotal from the existing record
	getExactInvoiceSubTotalValue, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.SubTotal, "2")
	if err != nil {
		return nil, err
	}

	// compare the subtotal should be equal to the total of invoice + the adjustment amounts
	expectedInvoiceSubTotalAmount := getExactInvoiceSubTotalValue + invoiceAdjustmentData.InvoiceTotalAmount
	if req.InvoiceSubTotal != expectedInvoiceSubTotalAmount {
		return nil, status.Error(codes.Internal, fmt.Sprintf("expected invoice subtotal amount %v received %v", expectedInvoiceSubTotalAmount, req.InvoiceSubTotal))
	}

	// Retry when there is a duplicate error from invoice_adjustment_sequence_number
	err = utils.DoWithMaxRetry(func(attempt int) (bool, error) {
		err := s.upsertInvoiceAdjustments(ctx, req, invoiceAdjustmentData)
		if err == nil {
			return false, nil
		}

		if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
			return false, err
		}

		log.Printf("Retrying creating of invoice adjustment. Attempt: %d \n", attempt)
		return attempt < invoiceAdjustmentMaxRetry, fmt.Errorf("cannot create invoice adjustment: %v", err)
	}, invoiceAdjustmentMaxRetry)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &invoice_pb.UpsertInvoiceAdjustmentsResponse{
		Success: true,
	}, nil
}

func (s *InvoiceModifierService) upsertInvoiceAdjustments(ctx context.Context, req *invoice_pb.UpsertInvoiceAdjustmentsRequest, invoiceAdjustmentData *InvoiceAdjustmentData) error {
	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// upsert invoice adjustment
		if len(invoiceAdjustmentData.InvoiceAdjustmentToUpsert) != 0 {
			if err := s.InvoiceAdjustmentRepo.UpsertMultiple(ctx, tx, invoiceAdjustmentData.InvoiceAdjustmentToUpsert); err != nil {
				return fmt.Errorf("error InvoiceAdjustmentRepo UpsertMultiple: %v", err)
			}
		}

		// soft delete invoice adjustment
		if len(invoiceAdjustmentData.InvoiceAdjustmentToDelete.Elements) != 0 {
			if err := s.InvoiceAdjustmentRepo.SoftDeleteByIDs(ctx, tx, invoiceAdjustmentData.InvoiceAdjustmentToDelete); err != nil {
				return fmt.Errorf("error InvoiceAdjustmentRepo SoftDeleteByIDs: %v", err)
			}
		}

		// update subtotal and total of invoice
		if err := s.updateInvoiceTotalSubtotalAndOutstandingBalance(ctx, tx, req.InvoiceId, req.InvoiceTotal, req.InvoiceSubTotal); err != nil {
			return err
		}

		// create action log
		actionDetails := &utils.InvoiceActionLogDetails{
			InvoiceID:     req.InvoiceId,
			Action:        invoice_pb.InvoiceAction_INVOICE_ADJUSTED,
			ActionComment: "",
		}
		if err := utils.CreateActionLogV2(ctx, tx, actionDetails, s.InvoiceActionLogRepo); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *InvoiceModifierService) generateInvoiceAdjustment(ctx context.Context, invoiceAdjustmentDetails []*invoice_pb.InvoiceAdjustmentDetail, invoice *entities.Invoice) (*InvoiceAdjustmentData, error) {
	var (
		invoiceAdjustmentToUpsert []*entities.InvoiceAdjustment
		invoiceAdjustmentToDelete pgtype.TextArray
		totalAmount               float64
	)

	for _, invoiceAdjustmentDetail := range invoiceAdjustmentDetails {
		var err error

		invoiceAdjustment := new(entities.InvoiceAdjustment)
		database.AllNullEntity(invoiceAdjustment)

		switch invoiceAdjustmentDetail.InvoiceAdjustmentAction {
		case invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT:
			err = multierr.Combine(
				invoiceAdjustment.InvoiceAdjustmentID.Set(idutil.ULIDNow()),
				invoiceAdjustment.Description.Set(invoiceAdjustmentDetail.Description),
				invoiceAdjustment.Amount.Set(invoiceAdjustmentDetail.Amount),
				invoiceAdjustment.InvoiceID.Set(invoice.InvoiceID),
				invoiceAdjustment.StudentID.Set(invoice.StudentID),
				invoiceAdjustment.CreatedAt.Set(time.Now()),
			)
			invoiceAdjustmentToUpsert = append(invoiceAdjustmentToUpsert, invoiceAdjustment)
			totalAmount += invoiceAdjustmentDetail.Amount

		case invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT:
			invoiceAdjustment, err := s.InvoiceAdjustmentRepo.FindByID(ctx, s.DB, invoiceAdjustmentDetail.InvoiceAdjustmentId)
			if err != nil {
				return nil, fmt.Errorf("error InvoiceAdjustmentRepo FindByID: %v", err)
			}

			getExactInvoiceAdjustmentAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoiceAdjustment.Amount, "2")
			if err != nil {
				return nil, err
			}

			err = multierr.Combine(
				invoiceAdjustment.InvoiceAdjustmentID.Set(invoiceAdjustmentDetail.InvoiceAdjustmentId),
				invoiceAdjustment.Description.Set(invoiceAdjustmentDetail.Description),
				invoiceAdjustment.Amount.Set(invoiceAdjustmentDetail.Amount),
			)
			if err != nil {
				return nil, err
			}

			invoiceAdjustmentToUpsert = append(invoiceAdjustmentToUpsert, invoiceAdjustment)
			// deduct first the existing amount and add the updated amount
			totalAmount += (invoiceAdjustmentDetail.Amount - getExactInvoiceAdjustmentAmount)

		case invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT:
			// find by id to get the amount that to be deleted
			invoiceAdjustment, err := s.InvoiceAdjustmentRepo.FindByID(ctx, s.DB, invoiceAdjustmentDetail.InvoiceAdjustmentId)
			if err != nil {
				return nil, fmt.Errorf("error InvoiceAdjustmentRepo FindByID: %v", err)
			}

			getExactInvoiceAdjustmentAmount, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoiceAdjustment.Amount, "2")
			if err != nil {
				return nil, err
			}

			totalAmount -= getExactInvoiceAdjustmentAmount
			invoiceAdjustmentToDelete = database.AppendText(invoiceAdjustmentToDelete, database.Text(invoiceAdjustmentDetail.InvoiceAdjustmentId))
		}

		if err != nil {
			return nil, err
		}
	}
	return &InvoiceAdjustmentData{
		InvoiceAdjustmentToUpsert: invoiceAdjustmentToUpsert,
		InvoiceAdjustmentToDelete: invoiceAdjustmentToDelete,
		InvoiceTotalAmount:        totalAmount,
	}, nil
}

func validateInvoiceAdjustmentRequest(req *invoice_pb.UpsertInvoiceAdjustmentsRequest) error {
	// validate invoice id
	if len(strings.TrimSpace(req.InvoiceId)) == 0 {
		return status.Error(codes.InvalidArgument, "invoice id is required")
	}

	// validate invoice adjustment detail
	if len(req.InvoiceAdjustmentDetails) == 0 {
		return status.Error(codes.InvalidArgument, "invoice adjustment detail is empty")
	}

	return nil
}

func validateInvoiceAdjustmentDetails(invoiceAdjustmentDetails []*invoice_pb.InvoiceAdjustmentDetail) error {
	for _, invoiceAdjustmentDetail := range invoiceAdjustmentDetails {
		switch invoiceAdjustmentDetail.InvoiceAdjustmentAction {
		case invoice_pb.InvoiceAdjustmentAction_CREATE_ADJUSTMENT:
			err := validateInvoiceAdjustmentID(invoiceAdjustmentDetail.InvoiceAdjustmentId)
			// expecting no invoice adjustment id on creating new record
			if err == nil {
				return fmt.Errorf("invalid invoice adjustment id: %v should be null when creating new record", invoiceAdjustmentDetail.InvoiceAdjustmentId)
			}

			err = validateAdjustmentDescription(invoiceAdjustmentDetail.Description)
			if err != nil {
				return err
			}

		case invoice_pb.InvoiceAdjustmentAction_EDIT_ADJUSTMENT:
			err := validateInvoiceAdjustmentID(invoiceAdjustmentDetail.InvoiceAdjustmentId)
			if err != nil {
				return err
			}

			err = validateAdjustmentDescription(invoiceAdjustmentDetail.Description)
			if err != nil {
				return err
			}

		case invoice_pb.InvoiceAdjustmentAction_DELETE_ADJUSTMENT:
			err := validateInvoiceAdjustmentID(invoiceAdjustmentDetail.InvoiceAdjustmentId)
			if err != nil {
				return err
			}
		default:
			return errors.New("invalid invoice adjustment action")
		}
	}

	return nil
}

func validateInvoiceAdjustmentID(invoiceAdjustmentID string) error {
	// validate invoice adjustment detail
	if len(strings.TrimSpace(invoiceAdjustmentID)) == 0 {
		return status.Error(codes.InvalidArgument, "invoice adjustment id is empty")
	}

	return nil
}

func validateAdjustmentDescription(description string) error {
	if len(strings.TrimSpace(description)) == 0 {
		return status.Error(codes.InvalidArgument, "invoice adjustment detail description is empty")
	}

	return nil
}

func (s *InvoiceModifierService) updateInvoiceTotalSubtotalAndOutstandingBalance(ctx context.Context, db database.QueryExecer, invoiceID string, total, subtotal float64) error {
	e := new(entities.Invoice)
	database.AllNullEntity(e)

	err := multierr.Combine(
		e.InvoiceID.Set(invoiceID),
		e.Total.Set(total),
		e.SubTotal.Set(subtotal),
		// since invoice is in draft status, outstanding balance and total should be the same
		e.OutstandingBalance.Set(total),
		e.UpdatedAt.Set(time.Now()),
	)

	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	if err := s.InvoiceRepo.UpdateWithFields(ctx, db, e, []string{"total", "sub_total", "outstanding_balance", "updated_at"}); err != nil {
		return fmt.Errorf("error InvoiceRepo UpdateWithFields: %w", err)
	}

	return nil
}
