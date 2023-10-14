package entities

import "github.com/jackc/pgtype"

type OrderCreator struct {
	OrderID pgtype.Text
	UserID  pgtype.Text
	Name    pgtype.Text
}

func (e *OrderCreator) FieldMap() ([]string, []interface{}) {
	return []string{
			"order_id",
			"user_id",
			"name",
		}, []interface{}{
			&e.OrderID,
			&e.UserID,
			&e.Name,
		}
}
