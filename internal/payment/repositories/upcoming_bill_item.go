package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpcomingBillItemRepo struct{}

func (r *UpcomingBillItemRepo) GetUpcomingBillItemsForGenerate(
	ctx context.Context,
	db database.QueryExecer,
) (billItems []entities.UpcomingBillItem, err error) {
	e := entities.UpcomingBillItem{}
	fieldNames, _ := e.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s
				FROM "%s"
				WHERE 
				billing_date::DATE <= NOW()::DATE
				AND is_generated = false
				AND deleted_at IS NULL
				ORDER BY billing_date DESC;
			   `,
		strings.Join(fieldNames, ","),
		e.TableName(),
	)

	rows, err := db.Query(ctx, stmt)
	if err != nil {
		err = status.Errorf(codes.Internal, "Err while get upcoming bill item to generate")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var upComingBillItem entities.UpcomingBillItem
		_, fieldValues := upComingBillItem.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			err = status.Errorf(codes.Internal, "Err while scan upcoming bill item to generate")
			return
		}
		billItems = append(billItems, upComingBillItem)
	}
	return
}

func (r *UpcomingBillItemRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.UpcomingBillItem) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "UpcomingBillItemRepo.Create")
	defer span.End()

	now := time.Now()
	if err = multierr.Combine(
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
		e.DeletedAt.Set(nil),
	); err != nil {
		err = fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
		return
	}
	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)

	if err != nil {
		err = fmt.Errorf("err insert upcoming bill item: %w", err)
		return
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert upcomingBillItem: %d RowsAffected", cmdTag.RowsAffected())
	}

	return
}

func (r *UpcomingBillItemRepo) RemoveOldUpcomingBillItem(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
	productID string,

) (
	billingSchedulePeriodID string,
	billingDate time.Time,
	err error,
) {
	ctx, span := interceptors.StartSpan(ctx, "UpcomingBillItemRepo.RemoveOldUpcomingBillItem")
	defer span.End()
	upcomingBillItem := &entities.UpcomingBillItem{}
	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = now(), updated_at = now() 
						WHERE order_id = $1 AND product_id = $2 AND is_generated = false
						AND deleted_at IS NULL
						RETURNING billing_date, billing_schedule_period_id;`, upcomingBillItem.TableName())
	row := db.QueryRow(ctx, sql, orderID, productID)
	err = row.Scan(&billingDate, &billingSchedulePeriodID)
	if err != nil && err != pgx.ErrNoRows {
		return billingSchedulePeriodID, billingDate, fmt.Errorf("err db.Exec UpcomingBillItemRepo.RemoveOldUpcomingBillItem: %w", err)
	}

	return billingSchedulePeriodID, billingDate, nil
}

func (r *UpcomingBillItemRepo) AddUpcomingExecuteNote(
	ctx context.Context,
	db database.QueryExecer,
	upcomingBillItem entities.UpcomingBillItem,
	err error,
) (importErr error) {
	ctx, span := interceptors.StartSpan(ctx, "UpcomingBillItemRepo.AddUpcomingExecuteNote")
	defer span.End()
	sql := fmt.Sprintf(`UPDATE %s SET execute_note = $1, updated_at = now() 
						WHERE order_id = $2 AND product_id = $3 AND is_generated = false
						AND deleted_at IS NULL
						RETURNING billing_schedule_period_id;`, upcomingBillItem.TableName())
	row := db.QueryRow(ctx, sql, err.Error(), upcomingBillItem.OrderID, upcomingBillItem.ProductID)
	importErr = row.Scan(&upcomingBillItem.BillingSchedulePeriodID)
	if err != nil {
		importErr = fmt.Errorf("err db.Exec UpcomingBillItemRepo.AddUpcomingExecuteNote: %w", err)
	}
	return
}

func (r *UpcomingBillItemRepo) UpdateCurrentUpcomingBillItemStatus(
	ctx context.Context,
	db database.QueryExecer,
	upcomingBillItem entities.UpcomingBillItem,
) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "UpcomingBillItemRepo.UpdateCurrentUpcomingBillItemStatus")
	defer span.End()
	sql := fmt.Sprintf(`UPDATE %s SET is_generated = true, updated_at = now() 
						WHERE order_id = $1 AND product_id = $2 AND billing_schedule_period_id = $3 AND is_generated = false
						AND deleted_at IS NULL
						RETURNING billing_schedule_period_id;`, upcomingBillItem.TableName())
	row := db.QueryRow(ctx, sql, upcomingBillItem.OrderID, upcomingBillItem.ProductID, upcomingBillItem.BillingSchedulePeriodID)
	err = row.Scan(&upcomingBillItem.BillingSchedulePeriodID)
	if err != nil {
		err = fmt.Errorf("err db.Exec UpcomingBillItemRepo.UpdateCurrentUpcomingBillItemStatus: %w", err)
	}
	return
}

func (r *UpcomingBillItemRepo) VoidUpcomingBillItemsByOrderID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "UpcomingBillItemRepo.VoidUpcomingBillItemsByOrderID")
	defer span.End()
	upcomingBillItem := &entities.UpcomingBillItem{}
	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = now(), updated_at = now() 
						WHERE order_id = $1 AND is_generated = false
						AND deleted_at IS NULL;`, upcomingBillItem.TableName())
	_, err = db.Exec(ctx, sql, orderID)

	if err != nil {
		return fmt.Errorf("err db.Exec UpcomingBillItemRepo.VoidUpcomingBillItemsByOrderID: %w", err)
	}
	return
}

func (r *UpcomingBillItemRepo) SetLastUpcomingBillItem(
	ctx context.Context,
	db database.QueryExecer,
	upcomingBillItem entities.UpcomingBillItem,
) (err error) {
	ctx, span := interceptors.StartSpan(ctx, "UpcomingBillItemRepo.SetLastUpcomingBillItem")
	defer span.End()
	executeNote := "last bill item"
	sql := fmt.Sprintf(`UPDATE %s SET is_generated = true, updated_at = now(), execute_note = $1 
						WHERE order_id = $2 AND product_id = $3 AND billing_schedule_period_id = $4 AND is_generated = false
						AND deleted_at IS NULL
						RETURNING billing_schedule_period_id;`, upcomingBillItem.TableName())
	row := db.QueryRow(ctx, sql, executeNote, upcomingBillItem.OrderID, upcomingBillItem.ProductID, upcomingBillItem.BillingSchedulePeriodID)
	err = row.Scan(&upcomingBillItem.BillingSchedulePeriodID)
	if err != nil {
		err = fmt.Errorf("err db.Exec UpcomingBillItemRepo.SetLastUpcomingBillItem: %w", err)
	}
	return
}

func (r *UpcomingBillItemRepo) GetUpcomingBillItemByOrderIDProductIDBillingSchedulePeriodID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
	productID string,
	billingSchedulePeriodID string,
) (upcomingBillItems []entities.UpcomingBillItem, err error) {
	e := entities.UpcomingBillItem{}
	fieldNames, _ := e.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s
				FROM "%s"
				WHERE 
					order_id = $1
				AND product_id = $2
				AND billing_schedule_period_id = $3
				AND is_generated = false
				AND deleted_at IS NULL;
			   `,
		strings.Join(fieldNames, ","),
		e.TableName(),
	)

	rows, err := db.Query(ctx, stmt, orderID, productID, billingSchedulePeriodID)
	if err != nil {
		err = status.Errorf(codes.Internal, "Err while get upcoming bill item to update")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var upComingBillItem entities.UpcomingBillItem
		_, fieldValues := upComingBillItem.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			err = status.Errorf(codes.Internal, "Err while scan upcoming bill item to update")
			return
		}
		upcomingBillItems = append(upcomingBillItems, upComingBillItem)
	}
	return
}
