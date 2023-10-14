package entities

import (
	"github.com/jackc/pgtype"
)

type LocationType struct {
	LocationTypeID       pgtype.Text
	Name                 pgtype.Text
	DisplayName          pgtype.Text
	ParentName           pgtype.Text
	ParentLocationTypeID pgtype.Text
	IsArchived           pgtype.Bool
	UpdatedAt            pgtype.Timestamptz
	CreatedAt            pgtype.Timestamptz
	DeletedAt            pgtype.Timestamptz
}

func (l *LocationType) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"location_type_id", "name", "display_name", "parent_name", "parent_location_type_id", "is_archived", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&l.LocationTypeID, &l.Name, &l.DisplayName, &l.ParentName, &l.ParentLocationTypeID, &l.IsArchived, &l.UpdatedAt, &l.CreatedAt, &l.DeletedAt}
	return
}

func (*LocationType) TableName() string {
	return "location_types"
}
