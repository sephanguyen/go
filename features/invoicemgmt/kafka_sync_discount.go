package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	paymentEntities "github.com/manabie-com/backend/internal/payment/entities"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) aDiscountRecordIsInsertedInFatima(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateDiscount(ctx, s.FatimaDBTrace),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisDiscountRecordMustBeRecordedInInvoicemgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT discount_id, name, discount_type, discount_amount_type
		FROM discount 
		WHERE discount_id = $1
	`

	// Get the discount in fatima DB
	fatimaDiscount := &paymentEntities.Discount{}
	fatimaRow := s.FatimaDBTrace.QueryRow(ctx, stmt, stepState.DiscountID)
	err := fatimaRow.Scan(
		&fatimaDiscount.DiscountID, &fatimaDiscount.Name, &fatimaDiscount.DiscountType, &fatimaDiscount.DiscountAmountType,
	)
	if err != nil {
		return ctx, err
	}

	if err := try.Do(func(attempt int) (bool, error) {

		// Get the discount in invoicemgmt DB
		discount := &entities.Discount{}
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.DiscountID).Scan(
			&discount.DiscountID, &discount.Name, &discount.DiscountType, &discount.DiscountAmountType,
		)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		if discount.DiscountID.String == fatimaDiscount.DiscountID.String &&
			discount.Name.String == fatimaDiscount.Name.String &&
			discount.DiscountType.String == fatimaDiscount.DiscountType.String &&
			discount.DiscountAmountType.String == fatimaDiscount.DiscountAmountType.String {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("discount record not sync correctly on invoicemgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
