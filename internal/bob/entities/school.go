package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type City struct {
	ID           pgtype.Int4 `sql:"city_id,pk"`
	Name         pgtype.Text
	Country      pgtype.Text
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DisplayOrder pgtype.Int2
}

func (c *City) FieldMap() ([]string, []interface{}) {
	return []string{
			"city_id", "name", "country", "created_at", "updated_at", "display_order",
		}, []interface{}{
			&c.ID, &c.Name, &c.Country, &c.CreatedAt, &c.UpdatedAt, &c.DisplayOrder,
		}
}

func (c *City) TableName() string { return "cities" }

type Citites []*City

func (u *Citites) Add() database.Entity {
	e := &City{}
	*u = append(*u, e)

	return e
}

type District struct {
	ID      pgtype.Int4 `sql:"district_id,pk"`
	Name    pgtype.Text
	Country pgtype.Text

	CityID pgtype.Int4 `sql:"city_id"`
	City   *City

	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

func (d *District) FieldMap() ([]string, []interface{}) {
	return []string{
			"district_id", "name", "country", "city_id", "created_at", "updated_at",
		}, []interface{}{
			&d.ID, &d.Name, &d.Country, &d.CityID, &d.CreatedAt, &d.UpdatedAt,
		}
}

func (d *District) TableName() string { return "districts" }

type Districts []*District

func (u *Districts) Add() database.Entity {
	e := &District{}
	*u = append(*u, e)

	return e
}

type School struct {
	ID          pgtype.Int4 `sql:"school_id,pk"`
	Name        pgtype.Text
	Country     pgtype.Text
	PhoneNumber pgtype.Text

	CityID pgtype.Int4 `sql:"city_id"`
	City   *City       `pg:"-"`

	DistrictID pgtype.Int4 `sql:"district_id"`
	District   *District   `pg:"-"`

	Point          pgtype.Point
	IsSystemSchool pgtype.Bool `sql:",notnull"`
	IsMerge        pgtype.Bool `sql:",notnull"`
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
}

func (s *School) FieldMap() ([]string, []interface{}) {
	return []string{
			"school_id", "name", "country", "city_id", "district_id", "point", "is_system_school", "is_merge", "phone_number", "created_at", "updated_at",
		}, []interface{}{
			&s.ID, &s.Name, &s.Country, &s.CityID, &s.DistrictID, &s.Point, &s.IsSystemSchool, &s.IsMerge, &s.PhoneNumber, &s.CreatedAt, &s.UpdatedAt,
		}
}

func (s *School) TableName() string { return "schools" }

type Schools []*School

func (u *Schools) Add() database.Entity {
	e := &School{}
	*u = append(*u, e)

	return e
}
