package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"go.uber.org/multierr"
)

type InvoiceScheduleHistoryRepo struct {
}

func (r *InvoiceScheduleHistoryRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.InvoiceScheduleHistory) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceScheduleHistoryRepo.Create")
	defer span.End()
	now := time.Now()

	if err := multierr.Combine(
		e.InvoiceScheduleHistoryID.Set(idutil.ULIDNow()),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return "", fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	var invoiceScheduleHistoryID string
	err := database.InsertReturningAndExcept(ctx, e, db, []string{"resource_path"}, "invoice_schedule_history_id", &invoiceScheduleHistoryID)
	if err != nil {
		return "", fmt.Errorf("err insert InvoiceScheduleHistory: %w", err)
	}

	return invoiceScheduleHistoryID, nil
}

func (r *InvoiceScheduleHistoryRepo) UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.InvoiceScheduleHistory, fields []string) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceScheduleHistoryRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "invoice_schedule_history_id", fields)
	if err != nil {
		return fmt.Errorf("err updateWithFields InvoiceScheduleHistory: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err updateWithFields InvoiceScheduleHistory: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
