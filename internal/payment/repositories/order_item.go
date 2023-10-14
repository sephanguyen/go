package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type OrderItemRepo struct{}

// Create creates OrderItem entity
func (r *OrderItemRepo) Create(ctx context.Context, tx database.QueryExecer, e entities.OrderItem) error {
	ctx, span := interceptors.StartSpan(ctx, "OrderItemRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, &e, []string{"resource_path"}, tx.Exec)
	if err != nil {
		return fmt.Errorf("err insert OrderItem: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert OrderItem: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

// GetStudentProductIDsForVoidOrderByOrderID Get student product id by order id
func (r *OrderItemRepo) GetStudentProductIDsForVoidOrderByOrderID(
	ctx context.Context,
	db database.QueryExecer,
	orderID string,
) (studentProductIDs []string, err error) {
	table := entities.OrderItem{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE order_id = $1`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)
	rows, err := db.Query(ctx, stmt, orderID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		orderItem := new(entities.OrderItem)
		_, fieldValues := orderItem.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			err = fmt.Errorf(constant.RowScanError, err)
			return
		}
		studentProductIDs = append(studentProductIDs, orderItem.StudentProductID.String)
	}

	return
}

func (r *OrderItemRepo) GetAllByOrderID(ctx context.Context, db database.QueryExecer, orderID string) ([]*entities.OrderItem, error) {
	table := entities.OrderItem{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE order_id = $1`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)
	rows, err := db.Query(ctx, stmt, orderID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []*entities.OrderItem
	for rows.Next() {
		orderItem := new(entities.OrderItem)
		_, fieldValues := orderItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		result = append(result, orderItem)
	}
	return result, nil
}

func (r *OrderItemRepo) CountOrderItemsByOrderID(ctx context.Context, db database.QueryExecer, orderID string) (count int, err error) {
	table := entities.OrderItem{}
	stmt := fmt.Sprintf(
		`SELECT count(*) FROM "%s" WHERE order_id = $1`,
		table.TableName(),
	)
	var totalOrderItems pgtype.Int8
	err = db.QueryRow(ctx, stmt, orderID).Scan(&totalOrderItems)
	if err != nil {
		err = fmt.Errorf(constant.RowScanError, err)
		return
	}
	count = int(totalOrderItems.Int)
	return
}

func (r *OrderItemRepo) GetOrderItemsByOrderIDWithPaging(ctx context.Context, db database.QueryExecer, orderID string, offset int64, limit int64) ([]entities.OrderItem, error) {
	table := entities.OrderItem{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE order_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)

	rows, err := db.Query(ctx, stmt, orderID, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []entities.OrderItem
	for rows.Next() {
		orderItem := new(entities.OrderItem)
		_, fieldValues := orderItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf(constant.RowScanError, err)
		}
		result = append(result, *orderItem)
	}
	return result, nil
}

func (r *OrderItemRepo) GetOrderItemsByProductIDs(ctx context.Context, db database.QueryExecer, productIDs []string) ([]entities.OrderItem, error) {
	table := entities.OrderItem{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE product_id = any($1)`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)
	rows, err := db.Query(ctx, stmt, productIDs)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []entities.OrderItem
	for rows.Next() {
		orderItem := new(entities.OrderItem)
		_, fieldValues := orderItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		result = append(result, *orderItem)
	}
	return result, nil
}

func (r *OrderItemRepo) GetOrderItemsByOrderIDs(ctx context.Context, db database.QueryExecer, orderIDs []string) ([]entities.OrderItem, error) {
	table := entities.OrderItem{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM "%s" WHERE order_id = any($1)`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)
	rows, err := db.Query(ctx, stmt, orderIDs)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []entities.OrderItem
	for rows.Next() {
		orderItem := new(entities.OrderItem)
		_, fieldValues := orderItem.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		result = append(result, *orderItem)
	}
	return result, nil
}

func (r *OrderItemRepo) GetOrderItemByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductID string) (orderItem entities.OrderItem, err error) {
	table := entities.OrderItem{}
	fieldNames, _ := table.FieldMap()
	stmt := fmt.Sprintf(
		`SELECT %s FROM %s WHERE student_product_id = $1`,
		strings.Join(fieldNames, ","),
		table.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentProductID)
	if err != nil {
		return
	}
	_, fieldValues := orderItem.FieldMap()
	err = row.Scan(fieldValues...)
	if err != nil {
		err = fmt.Errorf("row.Scan: %w", err)
		return
	}

	return
}
