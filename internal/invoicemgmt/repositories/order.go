package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
)

type OrderRepo struct{}

func (r *OrderRepo) FindByOrderID(ctx context.Context, db database.QueryExecer, orderID string) (*entities.Order, error) {
	ctx, span := interceptors.StartSpan(ctx, "OrderRepo.FindByOrderID")
	defer span.End()

	order := &entities.Order{}
	fields, _ := order.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM \"%s\" WHERE order_id = $1", strings.Join(fields, ","), order.TableName())

	err := database.Select(ctx, db, query, orderID).ScanOne(order)
	if err != nil {
		return nil, err
	}
	return order, nil
}
