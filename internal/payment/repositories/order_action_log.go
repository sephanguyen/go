package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"
	pmpb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"go.uber.org/multierr"
)

type OrderActionLogRepo struct{}

func (r *OrderActionLogRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.OrderActionLog) error {
	ctx, span := interceptors.StartSpan(ctx, "OrderActionLogRepo.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"order_action_log_id", "resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert OrderActionLog: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert OrderActionLog: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (r *OrderActionLogRepo) GetOrderCreatorsByOrderIDs(ctx context.Context, db database.QueryExecer, orderIDs []string) ([]entities.OrderCreator, error) {
	table := entities.OrderActionLog{}
	stmt := fmt.Sprintf(
		`SELECT oal.order_id, oal.user_id, u.name FROM "%s" oal 
    	JOIN users u ON u.user_id = oal.user_id 
        WHERE oal.action = $1 AND  oal.order_id = any($2)`,
		table.TableName(),
	)
	rows, err := db.Query(ctx, stmt, pmpb.OrderActionStatus_ORDER_ACTION_SUBMITTED.String(), orderIDs)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var result []entities.OrderCreator
	for rows.Next() {
		orderCreator := new(entities.OrderCreator)
		_, fieldValues := orderCreator.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		result = append(result, *orderCreator)
	}
	return result, nil
}
