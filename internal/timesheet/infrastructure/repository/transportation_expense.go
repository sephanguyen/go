package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type TransportationExpenseRepoImpl struct{}

func (r *TransportationExpenseRepoImpl) UpsertMultiple(ctx context.Context, db database.QueryExecer, listTransportExpenses []*entity.TransportationExpense) error {
	ctx, span := interceptors.StartSpan(ctx, "TransportationExpenseRepoImpl.UpsertMultiple")
	defer span.End()

	batch := &pgx.Batch{}
	now := time.Now()

	for _, TransportExpensesInfo := range listTransportExpenses {
		err := multierr.Combine(
			TransportExpensesInfo.UpdatedAt.Set(now),
			TransportExpensesInfo.CreatedAt.Set(now),
		)
		if err != nil {
			return err
		}

		fields, values := TransportExpensesInfo.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s ;",
			TransportExpensesInfo.TableName(),
			strings.Join(fields, ","),
			placeHolders,
			TransportExpensesInfo.UpsertConflictField(),
			TransportExpensesInfo.UpdateOnConflictQuery(),
		)

		batch.Queue(stmt, values...)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(listTransportExpenses); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return err
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err upsert transportation expense: %d RowsAffected", cmdTag.RowsAffected())
		}
	}

	return nil
}

func (r *TransportationExpenseRepoImpl) FindListTransportExpensesByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs pgtype.TextArray) ([]*entity.TransportationExpense, error) {
	ctx, span := interceptors.StartSpan(ctx, "TransportationExpenseRepoImpl.Retrieve")
	defer span.End()

	transportExpenseE := &entity.TransportationExpense{}
	listTransportExpensesE := &entity.ListTransportationExpenses{}

	values, _ := transportExpenseE.FieldMap()
	stmt := fmt.Sprintf(`
	SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND timesheet_id = ANY($1::_TEXT);`, strings.Join(values, ", "), transportExpenseE.TableName())

	if err := database.Select(ctx, db, stmt, timesheetIDs).ScanAll(listTransportExpensesE); err != nil {
		return nil, err
	}

	return *listTransportExpensesE, nil
}

func (r *TransportationExpenseRepoImpl) SoftDeleteByTimesheetID(ctx context.Context, db database.QueryExecer, timesheetID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "TransportationExpenseRepoImpl.SoftDeleteByTimesheetID")
	defer span.End()

	e := &entity.TransportationExpense{}

	stmt := fmt.Sprintf(`
		UPDATE %s SET deleted_at = NOW()
		WHERE timesheet_id = $1 AND deleted_at IS NULL;
	`, e.TableName())

	_, err := db.Exec(ctx, stmt, &timesheetID)
	if err != nil {
		return fmt.Errorf("err delete SoftDeleteByTimesheetID: %w", err)
	}

	return nil
}

func (r *TransportationExpenseRepoImpl) SoftDeleteMultipleByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "TransportationExpenseRepoImpl.SoftDeleteMultipleByTimesheetIDs")
	defer span.End()

	e := &entity.TransportationExpense{}

	stmt := fmt.Sprintf(`
		UPDATE %s SET deleted_at = NOW()
		WHERE deleted_at IS NULL
		AND timesheet_id = ANY($1::_TEXT);
	`, e.TableName())

	_, err := db.Exec(ctx, stmt, &timesheetIDs)
	if err != nil {
		return fmt.Errorf("err delete SoftDeleteMultipleByTimesheetIDs: %w", err)
	}

	return nil
}
