package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Staff struct {
	StaffID   pgtype.Text
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (e *Staff) FieldMap() ([]string, []interface{}) {
	return []string{
			"staff_id",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&e.StaffID,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
		}
}

func (e *Staff) TableName() string {
	return "staff"
}

type Staffs []*Staff

func (u *Staffs) Add() database.Entity {
	e := &Staff{}
	*u = append(*u, e)

	return e
}
