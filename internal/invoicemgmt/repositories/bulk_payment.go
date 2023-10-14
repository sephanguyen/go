package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"go.uber.org/multierr"
)

type BulkPaymentRepo struct {
}

func (r *BulkPaymentRepo) UpdateBulkPaymentStatusByIDs(ctx context.Context, db database.QueryExecer, status string, bulkPaymentIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentRepo.UpdateBulkPaymentStatusByIDs")
	defer span.End()

	e := &entities.BulkPayment{}

	query := fmt.Sprintf(`UPDATE %s SET updated_at = NOW(), bulk_payment_status = $1 WHERE bulk_payment_id = ANY($2)`, e.TableName())

	cmdTag, err := db.Exec(ctx, query, status, bulkPaymentIDs)
	if err != nil {
		return fmt.Errorf("err UpdateBulkPaymentStatusByIDs BulkPaymentRepo: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("err UpdateBulkPaymentStatusByIDs BulkPaymentRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *BulkPaymentRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPayment) error {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert BulkPaymentRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert BulkPaymentRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *BulkPaymentRepo) FindByBulkPaymentID(ctx context.Context, db database.QueryExecer, bulkPaymentID string) (*entities.BulkPayment, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentRepo.FindByBulkPaymentID")
	defer span.End()

	e := &entities.BulkPayment{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE bulk_payment_id = $1", strings.Join(fields, ","), e.TableName())

	err := database.Select(ctx, db, query, &bulkPaymentID).ScanOne(e)
	if err != nil {
		return nil, err
	}

	return e, nil
}
