package entity

import (
	"github.com/jackc/pgtype"
)

type UsrEmail struct {
	UsrID        pgtype.Text
	Email        pgtype.Text
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
	ImportID     pgtype.Int8
}

func (e *UsrEmail) TableName() string {
	return "usr_email"
}

func (e *UsrEmail) FieldMap() ([]string, []interface{}) {
	return []string{
			"usr_id",
			"email",
			"create_at",
			"updated_at",
			"delete_at",
			"resource_path",
			"import_id",
		}, []interface{}{
			&e.UsrID,
			&e.Email,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
			&e.ImportID,
		}
}
