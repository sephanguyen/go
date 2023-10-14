package entities

import "github.com/jackc/pgtype"

type LocationType struct {
	LocationTypeID       pgtype.Text
	Name                 pgtype.Text
	DisplayName          pgtype.Text
	ParentName           pgtype.Text
	ParentLocationTypeID pgtype.Text
	IsArchived           pgtype.Bool
	ResourcePath         pgtype.Text
	UpdatedAt            pgtype.Timestamptz
	CreatedAt            pgtype.Timestamptz
	DeletedAt            pgtype.Timestamptz
	Level                pgtype.Int4
}

func (e *LocationType) FieldMap() (fields []string, values []interface{}) {
	return []string{
			"location_type_id",
			"name",
			"display_name",
			"parent_name",
			"parent_location_type_id",
			"is_archived",
			"resource_path",
			"updated_at",
			"created_at",
			"deleted_at",
			"level",
		}, []interface{}{
			&e.LocationTypeID,
			&e.Name,
			&e.DisplayName,
			&e.ParentName,
			&e.ParentLocationTypeID,
			&e.IsArchived,
			&e.ResourcePath,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
			&e.Level,
		}
}

func (*LocationType) TableName() string {
	return "location_types"
}
