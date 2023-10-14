package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"go.uber.org/multierr"
)

type BillingRatioRepo struct{}

func (r *BillingRatioRepo) GetFirstRatioByBillingSchedulePeriodIDAndFromTime(ctx context.Context, db database.QueryExecer, billingSchedulePeriodID string, from time.Time) (entities.BillingRatio, error) {
	billingRatio := &entities.BillingRatio{}
	fieldNames, fieldValues := billingRatio.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			billing_schedule_period_id = $1
			AND is_archived = false
			AND start_date <= $2
			AND end_date >= $3
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		billingRatio.TableName(),
	)
	row := db.QueryRow(ctx, stmt, billingSchedulePeriodID, from, from)
	err := row.Scan(fieldValues...)
	return *billingRatio, err
}

// Create creates BillingRatio entity
func (r *BillingRatioRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.BillingRatio) error {
	ctx, span := interceptors.StartSpan(ctx, "BillingRatioRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.BillingRatioID.Set(idutil.ULIDNow()),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert BillingRatio: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert BillingRatio: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

// Update updates BillingRatio entity
func (r *BillingRatioRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.BillingRatio) error {
	ctx, span := interceptors.StartSpan(ctx, "BillingRatioRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "billing_ratio_id", []string{"start_date", "end_date", "billing_schedule_period_id", "billing_ratio_numerator", "billing_ratio_denominator", "is_archived", "updated_at"})

	if err != nil {
		return fmt.Errorf("err update BillingRatio: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update BillingRatio: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}

func (r *BillingRatioRepo) GetNextRatioByBillingSchedulePeriodIDAndPrevious(
	ctx context.Context,
	db database.QueryExecer,
	ratioOfProRatedBillingItem entities.BillingRatio,
) (entities.BillingRatio, error) {
	billingRatio := &entities.BillingRatio{}
	fieldNames, fieldValues := billingRatio.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			billing_schedule_period_id = $1
			AND is_archived = false
			AND start_date > $2
		ORDER BY start_date ASC
		LIMIT 1
		FOR NO KEY UPDATE
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		billingRatio.TableName(),
	)
	row := db.QueryRow(ctx, stmt, ratioOfProRatedBillingItem.BillingSchedulePeriodID, ratioOfProRatedBillingItem.EndDate)
	err := row.Scan(fieldValues...)
	return *billingRatio, err
}
