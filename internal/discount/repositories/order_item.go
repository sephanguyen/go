package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4"
)

type OrderItemRepo struct{}

func (r *OrderItemRepo) GetLatestByStudentProductID(ctx context.Context, db database.QueryExecer, studentProductID string) (entities.OrderItem, error) {
	orderItem := &entities.OrderItem{}
	orderItemFieldNames, orderItemFieldValues := orderItem.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			student_product_id = $1`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(orderItemFieldNames, ","),
		orderItem.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentProductID)
	err := row.Scan(orderItemFieldValues...)
	if err != nil {
		return entities.OrderItem{}, fmt.Errorf("row.Scan OrderItemRepo.GetLatestByStudentProductID: %w", err)
	}
	return *orderItem, nil
}

func (r *OrderItemRepo) GetStudentProductIDsByOrderID(ctx context.Context, db database.QueryExecer, orderID string) (studentProductIDs []string, err error) {
	orderItem := &entities.OrderItem{}
	stmt := `SELECT student_product_id
		FROM 
			%s
		WHERE 
			order_id = $1`

	stmt = fmt.Sprintf(
		stmt,
		orderItem.TableName(),
	)
	rows, err := db.Query(ctx, stmt, orderID)
	studentProductIDs = []string{}

	if err != nil {
		if err == pgx.ErrNoRows {
			return studentProductIDs, nil
		}
		return
	}

	defer rows.Close()

	for rows.Next() {
		var studentProductID string

		err := rows.Scan(
			&studentProductID,
		)
		if err != nil {
			return studentProductIDs, fmt.Errorf("row.Scan: %w", err)
		}

		studentProductIDs = append(studentProductIDs, studentProductID)
	}
	return studentProductIDs, nil
}
