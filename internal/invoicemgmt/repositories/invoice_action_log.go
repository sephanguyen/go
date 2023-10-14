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

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type InvoiceActionLogRepo struct {
}

// Create invoice action log record
func (r *InvoiceActionLogRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.InvoiceActionLog) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceActionLogRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.InvoiceActionID.Set(idutil.ULIDNow()),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert InvoiceActionLogRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert InvoiceActionLogRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *InvoiceActionLogRepo) GetLatestRecordByInvoiceID(ctx context.Context, db database.QueryExecer, invoiceID string) (*entities.InvoiceActionLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceActionLogRepo.GetLatestRecordByInvoiceID")
	defer span.End()

	e := &entities.InvoiceActionLog{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE invoice_id = $1 ORDER BY created_at DESC LIMIT 1", strings.Join(fields, ","), e.TableName())

	if err := database.Select(ctx, db, query, invoiceID).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (r *InvoiceActionLogRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, actionLogs []*entities.InvoiceActionLog) error {
	ctx, span := interceptors.StartSpan(ctx, "InvoiceActionLogRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, actionLog *entities.InvoiceActionLog) {
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
	for _, actionLog := range actionLogs {
		err := multierr.Combine(
			actionLog.InvoiceActionID.Set(idutil.ULIDNow()),
			actionLog.UpdatedAt.Set(now),
			actionLog.CreatedAt.Set(now),
		)
		if err != nil {
			return err
		}

		queueFn(batch, actionLog)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for i := 0; i < len(actionLogs); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("no rows affected when creating action log")
		}
	}

	return nil
}
