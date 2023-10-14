package entities

import "github.com/jackc/pgtype"

type Location struct {
	LocationID              pgtype.Text
	Name                    pgtype.Text
	LocationType            pgtype.Text
	ParentLocationID        pgtype.Text
	PartnerInternalID       pgtype.Text
	PartnerInternalParentID pgtype.Text
	UpdatedAt               pgtype.Timestamptz
	CreatedAt               pgtype.Timestamptz
	DeletedAt               pgtype.Timestamptz
	ResourcePath            pgtype.Text
}

func (e *Location) FieldMap() ([]string, []interface{}) {
	return []string{
			"location_id",
			"name",
			"location_type",
			"parent_location_id",
			"partner_internal_id",
			"partner_internal_parent_id",
			"updated_at",
			"created_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.LocationID,
			&e.Name,
			&e.LocationType,
			&e.ParentLocationID,
			&e.PartnerInternalID,
			&e.PartnerInternalParentID,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (*Location) TableName() string {
	return "locations"
}
