package entities

import (
	"github.com/jackc/pgtype"
)

type SchoolAdmin struct {
	User          `sql:"-"`
	SchoolAdminID pgtype.Text `sql:"school_admin_id,pk"`
	SchoolID      pgtype.Int4
	UpdatedAt     pgtype.Timestamptz
	CreatedAt     pgtype.Timestamptz
	ResourcePath  pgtype.Text
}

func (e *SchoolAdmin) FieldMap() ([]string, []interface{}) {
	return []string{
			"school_admin_id", "school_id", "updated_at", "created_at", "resource_path",
		}, []interface{}{
			&e.SchoolAdminID, &e.SchoolID, &e.UpdatedAt, &e.CreatedAt, &e.ResourcePath,
		}
}

func (e *SchoolAdmin) TableName() string {
	return "school_admins"
}
