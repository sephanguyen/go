package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"go.uber.org/multierr"
)

type BulkPaymentRequestRepo struct {
}

func (r *BulkPaymentRequestRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequest) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentRequestRepo.Create")
	defer span.End()

	bulkPaymentRequestID := idutil.ULIDNow()

	now := time.Now()
	if err := multierr.Combine(
		e.BulkPaymentRequestID.Set(bulkPaymentRequestID),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return "", fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return "", fmt.Errorf("err insert BulkPaymentRequestRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return "", fmt.Errorf("err insert BulkPaymentRequestRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return bulkPaymentRequestID, nil
}

func (r *BulkPaymentRequestRepo) Update(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequest) error {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentRequestRepo.Update")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "bulk_payment_request_id", []string{"error_details", "updated_at"})

	if err != nil {
		return fmt.Errorf("err update BulkPaymentRequestRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err update BulkPaymentRequestRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *BulkPaymentRequestRepo) FindByPaymentRequestID(ctx context.Context, db database.QueryExecer, id string) (*entities.BulkPaymentRequest, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentRequestRepo.FindByPaymentRequestID")
	defer span.End()

	e := &entities.BulkPaymentRequest{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE bulk_payment_request_id = $1", strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, query, id).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}
