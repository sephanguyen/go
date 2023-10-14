package entity

import (
	"github.com/jackc/pgtype"
)

type Staff struct {
	LegacyUser `sql:"-"`

	ID                  pgtype.Text `sql:"staff_id,pk"`
	UpdatedAt           pgtype.Timestamptz
	CreatedAt           pgtype.Timestamptz
	DeletedAt           pgtype.Timestamptz
	ResourcePath        pgtype.Text
	AutoCreateTimesheet pgtype.Bool
	WorkingStatus       pgtype.Text
	StartDate           pgtype.Date
	EndDate             pgtype.Date
}

func (s *Staff) FieldMap() ([]string, []interface{}) {
	return []string{
			"staff_id",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
			"auto_create_timesheet",
			"working_status",
			"start_date",
			"end_date",
		}, []interface{}{
			&s.ID,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.DeletedAt,
			&s.ResourcePath,
			&s.AutoCreateTimesheet,
			&s.WorkingStatus,
			&s.StartDate,
			&s.EndDate,
		}
}

func (*Staff) TableName() string {
	return "staff"
}
