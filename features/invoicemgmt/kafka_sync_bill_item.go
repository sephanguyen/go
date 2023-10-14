package invoicemgmt

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	paymentEntities "github.com/manabie-com/backend/internal/payment/entities"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) adminInsertsABillItemRecordToFatimaWithStatus(ctx context.Context, status string) (context.Context, error) {
	return s.createStudentWithBillItem(ctx, status, payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String())
}

func (s *suite) invoicemgmtBillItemTableWillBeUpdated(ctx context.Context) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	stepState := StepStateFromContext(ctx)

	stmt := `SELECT order_id, bill_item_sequence_number, student_product_id, product_id FROM bill_item WHERE bill_item_sequence_number = $1`

	fatimaBillItem := &paymentEntities.BillItem{}
	fatimaRow := s.FatimaDBTrace.QueryRow(ctx, stmt, s.StepState.BillItemSequenceNumber)
	err := fatimaRow.Scan(
		&fatimaBillItem.OrderID, &fatimaBillItem.BillItemSequenceNumber, &fatimaBillItem.StudentProductID, &fatimaBillItem.ProductID,
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bill item record not found in fatima")
	}

	if err := try.Do(func(attempt int) (bool, error) {

		billItem := &entities.BillItem{}
		row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, s.StepState.BillItemSequenceNumber)
		err := row.Scan(
			&billItem.OrderID, &billItem.BillItemSequenceNumber, &billItem.StudentProductID, &billItem.ProductID,
		)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		if fatimaBillItem.OrderID.String == billItem.OrderID.String &&
			fatimaBillItem.BillItemSequenceNumber.Int == billItem.BillItemSequenceNumber.Int &&
			fatimaBillItem.StudentProductID.String == billItem.StudentProductID.String &&
			fatimaBillItem.ProductID.String == billItem.ProductID.String {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("bill item record inserted not synced correctly on invoicemgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminDeletesThisBillItemRecordOnFatima(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stmt := `DELETE FROM bill_item WHERE bill_item_sequence_number = $1 AND resource_path = $2 AND student_id = $3`

	cmdTag, err := s.FatimaDBTrace.Exec(ctx, stmt, s.StepState.BillItemSequenceNumber, stepState.ResourcePath, stepState.StudentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error delete bill item sequence number: %d record in fatima", s.StepState.BillItemSequenceNumber)
	}

	if cmdTag.RowsAffected() == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no rows affected on delete bill item sequence number: %d record in fatima", s.StepState.BillItemSequenceNumber)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisBillItemOnInvoicemgmtWillBeDeleted(ctx context.Context) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	stepState := StepStateFromContext(ctx)

	stmt := `SELECT order_id FROM bill_item WHERE bill_item_sequence_number = $1 AND resource_path = $2 AND student_id = $3`

	if err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
		var invoiceBillItemOrderID string

		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, s.StepState.BillItemSequenceNumber, stepState.ResourcePath, stepState.StudentID).Scan(&invoiceBillItemOrderID)

		if err != nil && errors.Is(err, pgx.ErrNoRows) && strings.Trim(invoiceBillItemOrderID, " ") == "" {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)

		if attempt%10 == 0 {
			log.Printf("Selecting deleted bill item %d. Attempt: %d", s.StepState.BillItemSequenceNumber, attempt)
		}

		return attempt < 100, fmt.Errorf("bill item record deleted not synced correctly on invoicemgmt")
	}, 100); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
