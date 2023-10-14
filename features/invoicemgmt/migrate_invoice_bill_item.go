package invoicemgmt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manabie-com/backend/cmd/server/invoicemgmt"
	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

func (s *suite) thereAreExistingInvoiceAndBillItemThatHaveSameReference(ctx context.Context, count int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for i := 0; i < count; i++ {
		// Create student
		ctx, err := s.createStudent(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		time.Sleep(invoiceConst.KafkaSyncSleepDuration)

		// Create the invoice
		err = InsertEntities(
			StepStateFromContext(ctx),
			s.EntitiesCreator.CreateMigratedInvoice(ctx, s.InvoiceMgmtPostgresDBTrace, "ISSUED"),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		// Create the bill item
		_, err = s.createBillItemBasedOnStatusAndType(ctx, "INVOICED", payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		// Update the reference of bill_item
		stmt := "UPDATE bill_item SET reference = $1 WHERE bill_item_sequence_number = $2 AND resource_path = $3"
		if _, err := s.FatimaDBTrace.Exec(ctx, stmt, stepState.InvoiceReferenceID2, stepState.BillItemSequenceNumber, stepState.ResourcePath); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.InvoiceIDs = append(stepState.InvoiceIDs, stepState.InvoiceID)
		stepState.InvoiceIDInvoiceReferenceMap[stepState.InvoiceID] = stepState.InvoiceReferenceID
		stepState.InvoiceIDInvoiceReference2Map[stepState.InvoiceID] = stepState.InvoiceReferenceID2
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminRunsTheMigrateInvoiceBillItemScript(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Need to re create to set the current user ID to admin
	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	zLogger := logger.NewZapLogger("panic", true)
	err = invoicemgmt.MigrateInvoiceBillItem(
		ctx,
		s.InvoiceMgmtPostgresDBTrace,
		zLogger.Sugar(),
		stepState.OrganizationID,
		stepState.CurrentUserID,
	)
	stepState.ResponseErr = err

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminRunsTheMigrateInvoiceBillItemScriptWith(ctx context.Context, condition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Need to re create to set the current user ID to admin
	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	var (
		userID string
		orgID  string
	)

	switch condition {
	case "empty-orgID":
		orgID = " "
		userID = stepState.CurrentUserID
	case "empty-userID":
		orgID = stepState.ResourcePath
		userID = " "
	case "invalid-orgID":
		orgID = "invalid-test-id"
		userID = stepState.CurrentUserID
	}

	zLogger := logger.NewZapLogger("panic", true)
	err = invoicemgmt.MigrateInvoiceBillItem(
		ctx,
		s.InvoiceMgmtPostgresDBTrace,
		zLogger.Sugar(),
		orgID,
		userID,
	)
	stepState.ResponseErr = err

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) migrateInvoiceBillItemScriptHasNoError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting nil error got %v", stepState.ResponseErr)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) migrateInvoiceBillItemScriptReturnsError(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr == nil {
		return StepStateToContext(ctx, stepState), errors.New("expecting error got nil")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invoiceBillItemsWereSuccessfullyMigrated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	repo := &repositories.InvoiceBillItemRepo{}
	for _, invoiceID := range stepState.InvoiceIDs {
		invoiceBillitem, err := repo.FindAllByInvoiceID(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error finding invoice_bill_item err: %v", err)
		}

		if len(invoiceBillitem.ToArray()) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invoice bill item of invoice %v is not migrated", invoiceID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) migratedInvoiceBillItemHaveTheSameReference(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	repo := &repositories.InvoiceBillItemRepo{}

	invoiceRepo := &repositories.InvoiceRepo{}
	billItemRepo := &repositories.BillItemRepo{}
	for _, invoiceID := range stepState.InvoiceIDs {
		invoiceBillitem, err := repo.FindAllByInvoiceID(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error finding invoice_bill_item err: %v", err)
		}

		invoice, err := invoiceRepo.RetrieveInvoiceByInvoiceID(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceID)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error finding invoice err: %v", err)
		}

		for _, e := range invoiceBillitem.ToArray() {
			billItem, err := billItemRepo.FindByID(ctx, s.InvoiceMgmtPostgresDBTrace, e.BillItemSequenceNumber.Int)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error finding bill_item err: %v", err)
			}

			if billItem.Reference.String != invoice.InvoiceReferenceID2.String {
				return StepStateToContext(ctx, stepState), fmt.Errorf("invoice and bill item reference are not equal %v - %v",
					billItem.Reference.String, invoice.InvoiceReferenceID2.String)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
