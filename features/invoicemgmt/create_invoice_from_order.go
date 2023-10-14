package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *suite) thereAreExistingStudents(ctx context.Context, studentCount int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Create Student
	for i := 0; i < studentCount; i++ {
		ctx, err := s.createStudent(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		stepState.StudentIds = append(stepState.StudentIds, stepState.StudentID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eachOfTheseStudentsHaveOrdersWithStatusAndReviewRequiredTag(ctx context.Context, orderCount int, orderStatus, reviewRequiredTag string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		err           error
		orderStatusPB payment_pb.OrderStatus
		isReviewed    bool
	)

	switch orderStatus {
	case "SUBMITTED":
		orderStatusPB = payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED
	case "VOIDED":
		orderStatusPB = payment_pb.OrderStatus_ORDER_STATUS_VOIDED
	case "PENDING":
		orderStatusPB = payment_pb.OrderStatus_ORDER_STATUS_PENDING
	case "REJECTED":
		orderStatusPB = payment_pb.OrderStatus_ORDER_STATUS_REJECTED
	case "INVOICED":
		orderStatusPB = payment_pb.OrderStatus_ORDER_STATUS_INVOICED
	}

	if reviewRequiredTag == "no-existing" {
		isReviewed = true
	}

	// Create Student
	for i := 0; i < len(stepState.StudentIds); i++ {
		stepState.StudentID = stepState.StudentIds[i]

		// Create Order
		for j := 0; j < orderCount; j++ {
			err = InsertEntities(
				stepState,
				s.EntitiesCreator.CreateOrder(ctx, s.FatimaDBTrace, orderStatusPB.String(), isReviewed),
			)

			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			stepState.OrderIDs = append(stepState.OrderIDs, stepState.OrderID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eachOfTheseOrdersHaveBillItemWithStatus(ctx context.Context, billItemCount int, status string) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)
	stepState := StepStateFromContext(ctx)

	var billItemStatus string
	switch status {
	case "BILLED":
		billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()
	case "PENDING":
		billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_PENDING.String()
	case "INVOICED":
		billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_INVOICED.String()
	case "CANCELLED":
		billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_CANCELLED.String()
	case "WAITING_APPROVAL":
		billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_WAITING_APPROVAL.String()
	}

	for _, orderID := range stepState.OrderIDs {
		stepState.OrderID = orderID

		for i := 0; i < billItemCount; i++ {
			err := InsertEntities(
				StepStateFromContext(ctx),
				s.EntitiesCreator.CreateTax(ctx, s.FatimaDBTrace),
				s.EntitiesCreator.CreateBillingSchedule(ctx, s.FatimaDBTrace, true),
				s.EntitiesCreator.CreateBillingSchedulePeriod(ctx, s.FatimaDBTrace),
				s.EntitiesCreator.CreateProduct(ctx, s.FatimaDBTrace),
				s.EntitiesCreator.CreateStudentProduct(ctx, s.FatimaDBTrace),
				s.EntitiesCreator.CreateBillItem(ctx, s.FatimaDBTrace, billItemStatus, payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String()),
			)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminSelectsTheOrderList(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	orderDetails := []*invoice_pb.OrderDetail{}
	for _, orderID := range stepState.OrderIDs {
		orderDetails = append(orderDetails, &invoice_pb.OrderDetail{
			OrderId: orderID,
		})
	}

	stepState.Request = &invoice_pb.CreateInvoiceFromOrderRequest{
		OrderDetails: orderDetails,
		InvoiceType:  invoice_pb.InvoiceType_MANUAL,
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) submitsTheCreateInvoiceFromOrderRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req, ok := stepState.Request.(*invoice_pb.CreateInvoiceFromOrderRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("request should be type *invoice_pb.CreateInvoiceFromOrderRequest and not %T", req)
	}

	stepState.Response, stepState.ResponseErr = invoice_pb.NewInvoiceServiceClient(s.InvoiceMgmtConn).CreateInvoiceFromOrder(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eachInvoiceHaveBillItems(ctx context.Context, billItemCount int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, invoiceID := range stepState.InvoiceIDs {
		var count int
		stmt := "SELECT COUNT(*) FROM invoice_bill_item WHERE invoice_id = $1"
		row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, invoiceID)

		if err := row.Scan(&count); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if count != billItemCount {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting the bill item count of invoice %s to be %d got %d", invoiceID, billItemCount, count)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theseBillItemsHaveBillingStatus(ctx context.Context, status string) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)
	stepState := StepStateFromContext(ctx)

	var billItemStatus string
	switch status {
	case "BILLED":
		billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()
	case "PENDING":
		billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_PENDING.String()
	case "INVOICED":
		billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_INVOICED.String()
	case "CANCELLED":
		billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_CANCELLED.String()
	case "WAITING_APPROVAL":
		billItemStatus = payment_pb.BillingStatus_BILLING_STATUS_WAITING_APPROVAL.String()
	}

	for _, orderID := range stepState.OrderIDs {
		billItemRepo := &repositories.BillItemRepo{}
		billItems, err := billItemRepo.FindByOrderID(ctx, s.InvoiceMgmtPostgresDBTrace, orderID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		for _, billItem := range billItems {
			if billItem.BillStatus.String != billItemStatus {
				return StepStateToContext(ctx, stepState), fmt.Errorf("Expecting bill item %d status to be %s got %s", billItem.BillItemSequenceNumber.Int, billItemStatus, billItem.BillStatus.String)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) billingItemsOfOrderHaveBillingDateAtDay(ctx context.Context, identification string, day int) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)
	stepState := StepStateFromContext(ctx)

	billingDate := time.Now().UTC().AddDate(0, 0, day)

	stmt := "UPDATE bill_item SET billing_date = $1 WHERE bill_item_sequence_number = $2 AND resource_path = $3"

	for _, orderID := range stepState.OrderIDs {
		billItemRepo := &repositories.BillItemRepo{}
		billItems, err := billItemRepo.FindByOrderID(ctx, s.InvoiceMgmtPostgresDBTrace, orderID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if len(billItems) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("No bill item found on order with ID %s", orderID)
		}

		// If `one`, only update the first bill item
		if identification == "one" {
			_, err := s.FatimaDBTrace.Exec(ctx, stmt, billingDate, billItems[0].BillItemSequenceNumber.Int, s.ResourcePath)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error updating bill item: %v", err)
			}

			continue
		}

		// Update all bill item of order
		for _, billItem := range billItems {
			_, err := s.FatimaDBTrace.Exec(ctx, stmt, billingDate, billItem.BillItemSequenceNumber.Int, s.ResourcePath)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error updating bill item: %v", err)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thereIsInvoiceDateScheduledAtDay(ctx context.Context, day int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	invoiceDate := time.Now().AddDate(0, 0, day)
	err = InsertEntities(
		StepStateFromContext(ctx),
		s.EntitiesCreator.CreateInvoiceSchedule(ctx, s.InvoiceMgmtPostgresDBTrace, invoiceDate, invoiceDate.Add(24*time.Hour), invoice_pb.InvoiceScheduleStatus_INVOICE_SCHEDULE_SCHEDULED.String()),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error on creating InvoiceSchedule: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eachInvoiceHasCorrectTotalAmountAndOutstandingBalance(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, invoiceID := range stepState.InvoiceIDs {

		invoiceStmt := "SELECT total, outstanding_balance FROM invoice WHERE invoice_id = $1"
		billItemStmt := `
			SELECT bi.final_price, bi.adjustment_price
			FROM invoice_bill_item ibi
			INNER JOIN bill_item bi
				ON ibi.bill_item_sequence_number = bi.bill_item_sequence_number
					AND bi.resource_path = ibi.resource_path
			WHERE invoice_id = $1
		`

		// Get the total amount of invoice
		invoice := &entities.Invoice{}
		row := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, invoiceStmt, invoiceID)

		if err := row.Scan(&invoice.Total, &invoice.OutstandingBalance); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		invoiceTotal, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.Total, "2")
		if err != nil {
			return StepStateToContext(ctx, stepState), status.Error(codes.InvalidArgument, err.Error())
		}

		invoiceOutstandingBalance, err := utils.GetFloat64ExactValueAndDecimalPlaces(invoice.OutstandingBalance, "2")
		if err != nil {
			return StepStateToContext(ctx, stepState), status.Error(codes.InvalidArgument, err.Error())
		}

		// Get the bill items of invoice
		rows, err := s.InvoiceMgmtPostgresDBTrace.Query(ctx, billItemStmt, invoiceID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		defer rows.Close()

		var billItemsTotal float64
		for rows.Next() {
			billItem := new(entities.BillItem)
			database.AllNullEntity(billItem)

			err := rows.Scan(&billItem.FinalPrice, &billItem.AdjustmentPrice)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			amount := billItem.FinalPrice
			if billItem.AdjustmentPrice.Status == pgtype.Present {
				amount = billItem.AdjustmentPrice
			}

			exactPriceWithDecimalPlaces, err := utils.GetFloat64ExactValueAndDecimalPlaces(amount, "2")
			if err != nil {
				return StepStateToContext(ctx, stepState), status.Error(codes.InvalidArgument, err.Error())
			}

			billItemsTotal += exactPriceWithDecimalPlaces
		}

		if invoiceTotal != invoiceOutstandingBalance {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invoice %s total is not equal to its outstanding balance: %v, Invoice Total: %v", invoiceID, invoiceOutstandingBalance, invoiceTotal)
		}

		if invoiceTotal != billItemsTotal {
			return StepStateToContext(ctx, stepState), fmt.Errorf("invoice %s total is not equal to its bill item total. Bill Item Total: %v, Invoice Total: %v", invoiceID, billItemsTotal, invoiceTotal)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) billingItemsOfOrderHaveAdjustmentPrice(ctx context.Context, identification string, adjustmentPrice int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := "UPDATE bill_item SET adjustment_price = $1, bill_type = $2 WHERE bill_item_sequence_number = $3 AND resource_path = $4"

	for _, orderID := range stepState.OrderIDs {
		billItemRepo := &repositories.BillItemRepo{}
		billItems, err := billItemRepo.FindByOrderID(ctx, s.InvoiceMgmtPostgresDBTrace, orderID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if len(billItems) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("No bill item found on order with ID %s", orderID)
		}

		// If `one`, only update the first bill item
		if identification == "one" {
			_, err := s.FatimaDBTrace.Exec(ctx, stmt, adjustmentPrice, payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String(), billItems[0].BillItemSequenceNumber.Int, s.ResourcePath)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error updating bill item: %v", err)
			}

			continue
		}

		// Update all bill item of order
		for _, billItem := range billItems {
			_, err := s.FatimaDBTrace.Exec(ctx, stmt, adjustmentPrice, payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String(), billItem.BillItemSequenceNumber.Int, s.ResourcePath)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("error updating bill item: %v", err)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
