package entities

import "github.com/jackc/pgtype"

type LeavingReason struct {
	LeavingReasonID   pgtype.Text
	Name              pgtype.Text
	LeavingReasonType pgtype.Text
	Remark            pgtype.Text
	IsArchived        pgtype.Bool
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	ResourcePath      pgtype.Text
}

func (e *LeavingReason) TableName() string {
	return "leaving_reason"
}

func (e *LeavingReason) FieldMap() ([]string, []interface{}) {
	return []string{
			"leaving_reason_id",
			"name",
			"leaving_reason_type",
			"remark",
			"is_archived",
			"created_at",
			"updated_at",
			"resource_path",
		}, []interface{}{
			&e.LeavingReasonID,
			&e.Name,
			&e.LeavingReasonType,
			&e.Remark,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.ResourcePath,
		}
}
