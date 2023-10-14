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

type BulkPaymentRequestFileRepo struct {
}

func (r *BulkPaymentRequestFileRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequestFile) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentRequestFileRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return "", fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	if strings.TrimSpace(e.BulkPaymentRequestFileID.String) == "" {
		_ = e.BulkPaymentRequestFileID.Set(idutil.ULIDNow())
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return "", fmt.Errorf("err insert BulkPaymentRequestFileRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return "", fmt.Errorf("err insert BulkPaymentRequestFileRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return e.BulkPaymentRequestFileID.String, nil
}

func (r *BulkPaymentRequestFileRepo) FindByPaymentFileID(ctx context.Context, db database.QueryExecer, id string) (*entities.BulkPaymentRequestFile, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentRequestFileRepo.FindByPaymentFileID")
	defer span.End()

	e := &entities.BulkPaymentRequestFile{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE bulk_payment_request_file_id = $1", strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, query, id).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (r *BulkPaymentRequestFileRepo) UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentRequestFile, fieldsToUpdate []string) error {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentRequestFileRepo.UpdateWithFields")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "bulk_payment_request_file_id", fieldsToUpdate)

	if err != nil {
		return fmt.Errorf("err updateWithFields BulkPaymentRequestFileRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err updateWithFields BulkPaymentRequestFileRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
