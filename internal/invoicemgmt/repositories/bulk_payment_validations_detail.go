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
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type BulkPaymentValidationsDetailRepo struct {
}

func (r *BulkPaymentValidationsDetailRepo) RetrieveRecordsByBulkPaymentValidationsID(ctx context.Context, db database.QueryExecer, bulkPaymentValidationsID pgtype.Text) ([]*entities.BulkPaymentValidationsDetail, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentValidationsDetailRepo.FindByID")
	defer span.End()

	e := &entities.BulkPaymentValidationsDetail{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE bulk_payment_validations_id = $1 ORDER BY created_at DESC", strings.Join(fields, ","), e.TableName())
	rows, err := db.Query(ctx, query, &bulkPaymentValidationsID)
	if err != nil {
		return nil, fmt.Errorf("err retrieve records BulkPaymentValidationsDetailRepo: %w", err)
	}
	defer rows.Close()

	var result []*entities.BulkPaymentValidationsDetail

	for rows.Next() {
		bulkPaymentValidationDetail := &entities.BulkPaymentValidationsDetail{}
		_, values := bulkPaymentValidationDetail.FieldMap()
		if err := rows.Scan(values...); err != nil {
			return nil, err
		}
		result = append(result, bulkPaymentValidationDetail)
	}

	return result, nil
}

func (r *BulkPaymentValidationsDetailRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.BulkPaymentValidationsDetail) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentValidationsDetailRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return "", fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	if strings.TrimSpace(e.BulkPaymentValidationsDetailID.String) == "" {
		_ = e.BulkPaymentValidationsDetailID.Set(idutil.ULIDNow())
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return "", fmt.Errorf("err insert BulkPaymentValidationsDetailRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return "", fmt.Errorf("err insert BulkPaymentValidationsDetailRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return e.BulkPaymentValidationsDetailID.String, nil
}

func (r *BulkPaymentValidationsDetailRepo) FindByPaymentID(ctx context.Context, db database.QueryExecer, id string) (*entities.BulkPaymentValidationsDetail, error) {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentValidationsDetailRepo.FindByPaymentID")
	defer span.End()

	e := &entities.BulkPaymentValidationsDetail{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE payment_id = $1", strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, query, id).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (r *BulkPaymentValidationsDetailRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, validationDetails []*entities.BulkPaymentValidationsDetail) error {
	ctx, span := interceptors.StartSpan(ctx, "BulkPaymentValidationsDetailRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, actionLog *entities.BulkPaymentValidationsDetail) {
		fields := database.GetFieldNames(actionLog)
		fields = utils.RemoveStrFromSlice(fields, "resource_path")
		values := database.GetScanFields(actionLog, fields)

		placeHolders := database.GeneratePlaceholders(len(fields))

		stmt :=
			`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT
			DO NOTHING;
			`

		stmt = fmt.Sprintf(stmt, actionLog.TableName(), strings.Join(fields, ","), placeHolders)
		b.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}
	now := time.Now().UTC()
	for _, validationDetail := range validationDetails {
		err := multierr.Combine(
			validationDetail.UpdatedAt.Set(now),
			validationDetail.CreatedAt.Set(now),
		)
		if err != nil {
			return err
		}

		if strings.TrimSpace(validationDetail.BulkPaymentValidationsDetailID.String) == "" {
			_ = validationDetail.BulkPaymentValidationsDetailID.Set(idutil.ULIDNow())
		}

		queueFn(batch, validationDetail)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for i := 0; i < len(validationDetails); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("no rows affected when creating bulk payment validation details")
		}
	}

	return nil
}
