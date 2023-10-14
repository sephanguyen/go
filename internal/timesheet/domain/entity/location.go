package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Location struct {
	LocationID pgtype.Text
	Name       pgtype.Text
	UpdatedAt  pgtype.Timestamptz
	CreatedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

func (e *Location) FieldMap() ([]string, []interface{}) {
	return []string{
			"location_id",
			"name",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&e.LocationID,
			&e.Name,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
		}
}

func (e *Location) TableName() string {
	return "locations"
}

type Locations []*Location

func (u *Locations) Add() database.Entity {
	e := &Location{}
	*u = append(*u, e)

	return e
}
