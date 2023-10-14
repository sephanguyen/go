package invoicemgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"go.uber.org/multierr"
)

func (s *suite) thereIsExistingBulkPaymentValidationRecordFor(ctx context.Context, paymentMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// insert bulk payment validation records
	err := InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreateBulkPaymentValidations(ctx, s.InvoiceMgmtPostgresDBTrace, paymentMethod),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.PaymentMethod = paymentMethod

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisRecordConsistsOfPaymentValidated(ctx context.Context, paymentValidatedCount int, validationStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// reset values of payment ids when adding additional records
	stepState.PaymentIDs = []string{}

	if paymentValidatedCount > 0 {
		ctx, err := s.thereAreExistingPayments(ctx, paymentValidatedCount, validationStatus, stepState.PaymentMethod)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		ctx, err = s.createValidationPaymentsDetailsFromPayments(ctx, validationStatus)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	// update bulk payment validation successful and failed count
	stmt := `UPDATE bulk_payment_validations SET successful_payments = $1 WHERE bulk_payment_validations_id = $2`
	if validationStatus != "FAILED" {
		stmt = `UPDATE bulk_payment_validations SET failed_payments = $1 WHERE bulk_payment_validations_id = $2`
	}

	if _, err := s.InvoiceMgmtPostgresDBTrace.Exec(ctx, stmt, paymentValidatedCount, stepState.BulkPaymentValidationsID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error updateBillItemFinalPriceValueByStudentID: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminIsAtPaymentValidationScreen(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Request = &invoice_pb.DownloadBulkPaymentValidationsDetailRequest{}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) selectsTheExistingBulkPaymentValidationRecordToDownload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*invoice_pb.DownloadBulkPaymentValidationsDetailRequest)
	req.BulkPaymentValidationsId = stepState.BulkPaymentValidationsID
	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).DownloadBulkPaymentValidationsDetail(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createValidationPaymentsDetailsFromPayments(ctx context.Context, validationStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	paymentRepo := &repositories.PaymentRepo{}
	bulkPaymentValidationDetailRepo := &repositories.BulkPaymentValidationsDetailRepo{}

	for _, paymentID := range stepState.PaymentIDs {
		resultCode := "R0" //default success validation already transferred
		if validationStatus != "FAILED" {
			resultCode = "R1" //default failed validation
		}
		bulkPaymentValidationDetail := &entities.BulkPaymentValidationsDetail{}
		database.AllNullEntity(bulkPaymentValidationDetail)

		payment, err := paymentRepo.FindByPaymentID(ctx, s.InvoiceMgmtPostgresDBTrace, paymentID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("paymentRepo.FindByPaymentID err: %v", err)
		}

		switch payment.PaymentMethod.String {
		case invoice_pb.PaymentMethod_DIRECT_DEBIT.String():
			resultCode = "D-" + resultCode
		case invoice_pb.PaymentMethod_CONVENIENCE_STORE.String():
			resultCode = "C-" + resultCode
		}

		err = multierr.Combine(
			bulkPaymentValidationDetail.BulkPaymentValidationsID.Set(stepState.BulkPaymentValidationsID),
			bulkPaymentValidationDetail.PaymentID.Set(paymentID),
			bulkPaymentValidationDetail.InvoiceID.Set(payment.InvoiceID),
			bulkPaymentValidationDetail.ValidatedResultCode.Set(resultCode),
			bulkPaymentValidationDetail.PaymentStatus.Set(payment.PaymentStatus),
		)

		_, err = bulkPaymentValidationDetailRepo.Create(ctx, s.InvoiceMgmtPostgresDBTrace, bulkPaymentValidationDetail)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("bulkPaymentRequestFilePaymentRepo.Create err: %v", err)
		}

		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error generating bulk payment validation entity: %v", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) hasResponsePaymentDataWithCorrectRecords(ctx context.Context, totalValidatedRecords int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if s.StepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to send payment validated data err: %v", s.StepState.ResponseErr)
	}
	resp, ok := stepState.Response.(*invoice_pb.DownloadBulkPaymentValidationsDetailResponse)

	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no download bulk payment validation detail response err")
	}

	importPaymentValidationDetail := resp.PaymentValidationDetail

	if len(importPaymentValidationDetail) != totalValidatedRecords {
		return StepStateToContext(ctx, stepState), fmt.Errorf("records count not match expected %d got %d on bulk payment validations id %v", totalValidatedRecords, len(importPaymentValidationDetail), stepState.BulkPaymentValidationsID)
	}

	for _, paymentValidationDetail := range importPaymentValidationDetail {
		var count int
		stmt := `
			SELECT
				COUNT(b.bulk_payment_validations_detail_id)
			FROM bulk_payment_validations_detail b
			INNER JOIN payment p
				ON b.payment_id = p.payment_id
			INNER JOIN invoice i
				ON b.invoice_id = i.invoice_id
			WHERE p.payment_sequence_number = $1 
				AND i.invoice_sequence_number = $2 
				AND b.resource_path = $3
				AND b.bulk_payment_validations_id = $4
				AND p.student_id = $5
		`
		resourcePath := golibs.ResourcePathFromCtx(ctx)

		e := &entities.PaymentInvoiceMap{}
		e.Payment = &entities.Payment{}
		e.Invoice = &entities.Invoice{}

		if err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, paymentValidationDetail.PaymentSequenceNumber, paymentValidationDetail.InvoiceSequenceNumber, resourcePath, stepState.BulkPaymentValidationsID, paymentValidationDetail.StudentId).Scan(&count); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if count == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid record on response err")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) hasResponseValidationDate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := stepState.Response.(*invoice_pb.DownloadBulkPaymentValidationsDetailResponse)

	if resp.ValidationDate == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected non-nil validation date")
	}

	return StepStateToContext(ctx, stepState), nil
}
