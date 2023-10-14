package entities

import "github.com/jackc/pgtype"

type OrderActionLog struct {
	OrderActionLogID pgtype.Int4
	OrderID          pgtype.Text
	UserID           pgtype.Text
	Action           pgtype.Text
	Comment          pgtype.Text
	UpdatedAt        pgtype.Timestamptz
	CreatedAt        pgtype.Timestamptz
	ResourcePath     pgtype.Text
}

func (e *OrderActionLog) FieldMap() ([]string, []interface{}) {
	return []string{
			"order_action_log_id",
			"order_id",
			"user_id",
			"action",
			"comment",
			"created_at",
			"updated_at",
			"resource_path",
		}, []interface{}{
			&e.OrderActionLogID,
			&e.OrderID,
			&e.UserID,
			&e.Action,
			&e.Comment,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.ResourcePath,
		}
}

func (e *OrderActionLog) TableName() string {
	return "order_action_log"
}
