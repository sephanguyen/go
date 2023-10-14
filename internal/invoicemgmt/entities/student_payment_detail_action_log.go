package entities

import "github.com/jackc/pgtype"

type StudentPaymentDetailActionLog struct {
	StudentPaymentDetailActionID pgtype.Text
	StudentPaymentDetailID       pgtype.Text
	UserID                       pgtype.Text
	Action                       pgtype.Text
	ActionDetail                 pgtype.JSONB
	ResourcePath                 pgtype.Text
	CreatedAt                    pgtype.Timestamptz
	UpdatedAt                    pgtype.Timestamptz
	DeletedAt                    pgtype.Timestamptz
}

func (e *StudentPaymentDetailActionLog) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_payment_detail_action_id",
			"student_payment_detail_id",
			"user_id",
			"action",
			"action_detail",
			"resource_path",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&e.StudentPaymentDetailActionID,
			&e.StudentPaymentDetailID,
			&e.UserID,
			&e.Action,
			&e.ActionDetail,
			&e.ResourcePath,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
		}
}

func (*StudentPaymentDetailActionLog) TableName() string {
	return "student_payment_detail_action_log"
}
