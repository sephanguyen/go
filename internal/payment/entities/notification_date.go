package entities

import "github.com/jackc/pgtype"

type NotificationDate struct {
	NotificationDateID pgtype.Text
	OrderType          pgtype.Text
	NotificationDate   pgtype.Int4
	IsArchived         pgtype.Bool
	CreatedAt          pgtype.Timestamptz
	UpdatedAt          pgtype.Timestamptz
	ResourcePath       pgtype.Text
}

func (e *NotificationDate) TableName() string {
	return "notification_date"
}

func (e *NotificationDate) FieldMap() ([]string, []interface{}) {
	return []string{
			"notification_date_id",
			"order_type",
			"notification_date",
			"is_archived",
			"created_at",
			"updated_at",
			"resource_path",
		}, []interface{}{
			&e.NotificationDateID,
			&e.OrderType,
			&e.NotificationDate,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.ResourcePath,
		}
}
