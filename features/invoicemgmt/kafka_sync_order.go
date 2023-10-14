package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	paymentEntities "github.com/manabie-com/backend/internal/payment/entities"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) anOrderRecordIsInsertedIntoFatima(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Student creation not needed; no constraint for it
	stepState.StudentID = idutil.ULIDNow()

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateOrder(ctx, s.FatimaDBTrace, payment_pb.OrderStatus_ORDER_STATUS_SUBMITTED.String(), false),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisOrderRecordMustBeRecordedInInvoicemgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT order_id, student_id, location_id, order_sequence_number
		FROM "order" 
		WHERE order_id = $1
	`

	// Get order record from fatima DB
	fatimaOrder := &paymentEntities.Order{}
	fatimaRow := s.FatimaDBTrace.QueryRow(ctx, stmt, stepState.OrderID)
	err := fatimaRow.Scan(
		&fatimaOrder.OrderID, &fatimaOrder.StudentID, &fatimaOrder.LocationID, &fatimaOrder.OrderSequenceNumber,
	)
	if err != nil {
		return ctx, err
	}

	if err := try.Do(func(attempt int) (bool, error) {

		// Get order record from invoicemgmt DB
		order := &entities.Order{}
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.OrderID).Scan(
			&order.OrderID, &order.StudentID, &order.LocationID, &order.OrderSequenceNumber,
		)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		if order.OrderID.String == fatimaOrder.OrderID.String &&
			order.StudentID.String == fatimaOrder.StudentID.String &&
			order.LocationID.String == fatimaOrder.LocationID.String &&
			order.OrderSequenceNumber.Int == fatimaOrder.OrderSequenceNumber.Int {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("order record not sync correctly on invoicemgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
