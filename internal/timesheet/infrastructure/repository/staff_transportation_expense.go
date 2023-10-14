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

type StaffTransportationExpenseRepoImpl struct{}

func (r *StaffTransportationExpenseRepoImpl) UpsertMultiple(ctx context.Context, db database.QueryExecer, listStaffTransportExpenses []*entity.StaffTransportationExpense) error {
	ctx, span := interceptors.StartSpan(ctx, "StaffTransportationExpenseRepoImpl.UpsertMultiple")
	defer span.End()

	batch := &pgx.Batch{}
	now := time.Now()

	for _, TransportExpensesInfo := range listStaffTransportExpenses {
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

	for i := 0; i < len(listStaffTransportExpenses); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return err
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err upsert staff transportation expense: %d RowsAffected", cmdTag.RowsAffected())
		}
	}

	return nil
}

func (r *StaffTransportationExpenseRepoImpl) FindListTransportExpensesByStaffIDs(ctx context.Context, db database.QueryExecer, staffIDs pgtype.TextArray) ([]*entity.StaffTransportationExpense, error) {
	ctx, span := interceptors.StartSpan(ctx, "StaffTransportationExpenseRepoImpl.FindListTransportExpensesByStaffIDs")
	defer span.End()

	staffTransportExpenseE := &entity.StaffTransportationExpense{}
	listStaffTransportExpensesE := &entity.ListStaffTransportationExpense{}

	fields, _ := staffTransportExpenseE.FieldMap()
	stmt := fmt.Sprintf(`
	SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND staff_id = ANY($1::_TEXT);`, strings.Join(fields, ", "), staffTransportExpenseE.TableName())

	if err := database.Select(ctx, db, stmt, staffIDs).ScanAll(listStaffTransportExpensesE); err != nil {
		return nil, err
	}

	return *listStaffTransportExpensesE, nil
}

func (r *StaffTransportationExpenseRepoImpl) FindListTransportExpensesByStaffIDsAndLocation(ctx context.Context, db database.QueryExecer, staffIDs []string, location string) (map[string][]entity.StaffTransportationExpense, error) {
	ctx, span := interceptors.StartSpan(ctx, "StaffTransportationExpenseRepoImpl.FindListTransportExpensesByStaffIDsAndLocation")
	defer span.End()

	staffTransportExpenseE := &entity.StaffTransportationExpense{}
	fields, _ := staffTransportExpenseE.FieldMap()

	stmt := fmt.Sprintf(`
	SELECT %s
	FROM %s
	WHERE deleted_at IS NULL
	AND staff_id = ANY($1::_TEXT)
	AND location_id = $2;`, strings.Join(fields, ", "), staffTransportExpenseE.TableName())

	listStaffTransportExpensesE := &entity.ListStaffTransportationExpense{}
	if err := database.Select(ctx, db, stmt, staffIDs, location).ScanAll(listStaffTransportExpensesE); err != nil {
		return nil, err
	}

	result := map[string][]entity.StaffTransportationExpense{}
	for _, staffTE := range *listStaffTransportExpensesE {
		result[staffTE.StaffID.String] = append(result[staffTE.StaffID.String], *staffTE)
	}

	return result, nil
}
