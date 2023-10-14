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

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type NewCustomerCodeHistoryRepo struct {
}

func (r *NewCustomerCodeHistoryRepo) FindByStudentIDs(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]*entities.NewCustomerCodeHistory, error) {
	_, span := interceptors.StartSpan(ctx, "NewCustomerCodeHistoryRepo.FindByStudentIDs")
	defer span.End()

	e := entities.NewCustomerCodeHistory{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = ANY($1)", strings.Join(fields, ","), e.TableName())

	var ids pgtype.TextArray
	_ = ids.Set(studentIDs)

	rows, err := db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}

	results := []*entities.NewCustomerCodeHistory{}
	defer rows.Close()
	for rows.Next() {
		e := &entities.NewCustomerCodeHistory{}
		_, values := e.FieldMap()

		err := rows.Scan(values...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		results = append(results, e)
	}

	return results, nil
}

func (r *NewCustomerCodeHistoryRepo) FindByAccountNumbers(ctx context.Context, db database.QueryExecer, bankAccountNumbers []string) ([]*entities.NewCustomerCodeHistory, error) {
	_, span := interceptors.StartSpan(ctx, "NewCustomerCodeHistoryRepo.FindByAccountNumbers")
	defer span.End()

	e := entities.NewCustomerCodeHistory{}
	fields, _ := e.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE bank_account_number = ANY($1)", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(ctx, query, bankAccountNumbers)
	if err != nil {
		return nil, err
	}

	results := []*entities.NewCustomerCodeHistory{}
	defer rows.Close()
	for rows.Next() {
		e := &entities.NewCustomerCodeHistory{}
		_, values := e.FieldMap()

		err := rows.Scan(values...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		results = append(results, e)
	}

	return results, nil
}

func (r *NewCustomerCodeHistoryRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.NewCustomerCodeHistory) error {
	ctx, span := interceptors.StartSpan(ctx, "NewCustomerCodeHistoryRepo.Create")
	defer span.End()

	now := time.Now()
	id := idutil.ULIDNow()
	if err := multierr.Combine(
		e.NewCustomerCodeHistoryID.Set(id),
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert NewCustomerCodeHistoryRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert NewCustomerCodeHistoryRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *NewCustomerCodeHistoryRepo) UpdateWithFields(ctx context.Context, db database.QueryExecer, e *entities.NewCustomerCodeHistory, fieldsToUpdate []string) error {
	ctx, span := interceptors.StartSpan(ctx, "NewCustomerCodeHistoryRepo.UpdateWithFields")
	defer span.End()

	now := time.Now()
	if err := e.UpdatedAt.Set(now); err != nil {
		return fmt.Errorf("UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.UpdateFields(ctx, e, db.Exec, "new_customer_code_history_id", fieldsToUpdate)

	if err != nil {
		return fmt.Errorf("err updateWithFields NewCustomerCodeHistoryRepo: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err updateWithFields NewCustomerCodeHistoryRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
