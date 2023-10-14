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

type BillingScheduleRepo struct{}

func (r *BillingScheduleRepo) GetByIDForUpdate(ctx context.Context, db database.QueryExecer, billingScheduleID string) (entities.BillingSchedule, error) {
	result := entities.BillingSchedule{}
	fieldNames, fieldValues := result.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE billing_schedule_id = $1 FOR NO KEY UPDATE`,
		strings.Join(fieldNames, ","),
		result.TableName(),
	)
	row := db.QueryRow(ctx, stmt, billingScheduleID)
	err := row.Scan(fieldValues...)
	if err != nil {
		return entities.BillingSchedule{}, fmt.Errorf("row.Scan: %w", err)
	}
	return result, nil
}

// Create creates BillingScheduleRepo entity
func (r *BillingScheduleRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.BillingSchedule) error {
	ctx, span := interceptors.StartSpan(ctx, "BillingScheduleRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.BillingScheduleID.Set(idutil.ULIDNow()),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert BillingSchedule: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert BillingSchedule: %d RowsAffected", cmdTag.RowsAffected())
	}
	return nil
}

// Update updates BillingSchedule entity
func (r *BillingScheduleRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.BillingSchedule) error {
	ctx, span := interceptors.StartSpan(ctx, "BillingScheduleRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "billing_schedule_id", []string{"name", "remarks", "is_archived", "updated_at"})
	if err != nil {
		return fmt.Errorf("err update BillingSchedule: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update BillingSchedule: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
