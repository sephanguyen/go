package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type InvoiceAdjustmentRepo struct{}

func (r *InvoiceAdjustmentRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, invoiceAdjustments []*entities.InvoiceAdjustment) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceAdjustmentRepo.UpsertMultiple")
	defer span.End()

	batch := &pgx.Batch{}
	now := time.Now()

	for _, invoiceAdjustment := range invoiceAdjustments {
		err := invoiceAdjustment.UpdatedAt.Set(now)
		if err != nil {
			return err
		}
		initialFields := database.GetFieldNames(invoiceAdjustment)
		fields := utils.RemoveStrFromSlice(initialFields, "resource_path")
		fields = utils.RemoveStrFromSlice(fields, "invoice_adjustment_sequence_number")
		values := database.GetScanFields(invoiceAdjustment, fields)

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s ;",
			invoiceAdjustment.TableName(),
			strings.Join(fields, ","),
			placeHolders,
			invoiceAdjustment.PrimaryKey(),
			invoiceAdjustment.UpdateOnConflictQuery(),
		)

		batch.Queue(stmt, values...)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(invoiceAdjustments); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return err
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err upsert multiple invoice adjustment: %d RowsAffected", cmdTag.RowsAffected())
		}
	}

	return nil
}

func (r *InvoiceAdjustmentRepo) SoftDeleteByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceAdjustmentRepo.SoftDeleteByIDs")
	defer span.End()

	e := &entities.InvoiceAdjustment{}

	now := time.Now()

	stmt := fmt.Sprintf(`
		UPDATE %s SET deleted_at = $1, updated_at = $2
		WHERE invoice_adjustment_id = ANY($3::_TEXT)
		AND deleted_at IS NULL;`, e.TableName())

	cmdTag, err := db.Exec(ctx, stmt, now, now, ids)
	if err != nil {
		return fmt.Errorf("err delete InvoiceAdjustmentRepo: %w", err)
	}

	if cmdTag.RowsAffected() != int64(len(ids.Elements)) {
		return fmt.Errorf("err delete InvoiceAdjustmentRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *InvoiceAdjustmentRepo) FindByID(ctx context.Context, db database.QueryExecer, id string) (*entities.InvoiceAdjustment, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceAdjustmentRepo.FindByID")
	defer span.End()

	e := &entities.InvoiceAdjustment{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_adjustment_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, query, id).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (r *InvoiceAdjustmentRepo) FindByInvoiceIDs(ctx context.Context, db database.QueryExecer, invoiceIDs []string) ([]*entities.InvoiceAdjustment, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceAdjustmentRepo.FindByInvoiceIDs")
	defer span.End()

	e := &entities.InvoiceAdjustment{}
	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_id = ANY($1)", strings.Join(fields, ","), e.TableName())

	var ids pgtype.TextArray
	_ = ids.Set(invoiceIDs)

	rows, err := db.Query(ctx, stmt, ids)
	if err != nil {
		return nil, err
	}

	res := []*entities.InvoiceAdjustment{}
	defer rows.Close()
	for rows.Next() {
		e := new(entities.InvoiceAdjustment)
		database.AllNullEntity(e)

		_, fieldValues := e.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		res = append(res, e)
	}

	return res, nil
}
