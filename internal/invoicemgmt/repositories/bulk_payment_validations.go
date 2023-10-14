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

type BulkPaymentValidationsRepo struct {
}

func (r *BulkPaymentValidationsRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentValidations) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentValidationsRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return "", fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	if strings.TrimSpace(e.BulkPaymentValidationsID.String) == "" {
		_ = e.BulkPaymentValidationsID.Set(idutil.ULIDNow())
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return "", fmt.Errorf("err insert BulkPaymentValidationsRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return "", fmt.Errorf("err insert BulkPaymentValidationsRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return e.BulkPaymentValidationsID.String, nil
}

func (r *BulkPaymentValidationsRepo) UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentValidations, fieldsToUpdate []string) error {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentValidationsRepo.UpdateWithFields")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "bulk_payment_validations_id", fieldsToUpdate)

	if err != nil {
		return fmt.Errorf("err updateWithFields BulkPaymentValidationsRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err updateWithFields BulkPaymentValidationsRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *BulkPaymentValidationsRepo) FindByID(ctx context.Context, db database.QueryExecer, bulkPaymentValidationsID string) (*entities.BulkPaymentValidations, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentValidationsRepo.FindByID")
	defer span.End()

	e := &entities.BulkPaymentValidations{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE bulk_payment_validations_id = $1", strings.Join(fields, ","), e.TableName())

	err := database.Select(ctx, db, query, bulkPaymentValidationsID).ScanOne(e)

	if err != nil {
		return nil, fmt.Errorf("err FindByID BulkPaymentValidations: %w", err)
	}

	return e, nil
}
