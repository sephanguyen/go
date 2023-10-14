package repositories

import (
	"github.com/manabie-com/backend/internal/calendar/domain/dto"

	"github.com/jackc/pgtype"
)

type Location struct {
	LocationID              pgtype.Text
	Name                    pgtype.Text
	LocationType            pgtype.Text
	ParentLocationID        pgtype.Text
	PartnerInternalID       pgtype.Text
	PartnerInternalParentID pgtype.Text
	IsArchived              pgtype.Bool
	AccessPath              pgtype.Text
	CreatedAt               pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
	DeletedAt               pgtype.Timestamptz
}

func (l *Location) FieldMap() ([]string, []interface{}) {
	return []string{
			"location_id",
			"name",
			"location_type",
			"parent_location_id",
			"partner_internal_id",
			"partner_internal_parent_id",
			"is_archived",
			"access_path",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&l.LocationID,
			&l.Name,
			&l.LocationType,
			&l.ParentLocationID,
			&l.PartnerInternalID,
			&l.PartnerInternalParentID,
			&l.IsArchived,
			&l.AccessPath,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.DeletedAt,
		}
}

func (l *Location) TableName() string {
	return "locations"
}

func (l *Location) ConvertToDTO() *dto.Location {
	return &dto.Location{
		LocationID:              l.LocationID.String,
		Name:                    l.Name.String,
		LocationType:            l.LocationType.String,
		ParentLocationID:        l.ParentLocationID.String,
		PartnerInternalID:       l.PartnerInternalID.String,
		PartnerInternalParentID: l.PartnerInternalID.String,
		IsArchived:              l.IsArchived.Bool,
		AccessPath:              l.AccessPath.String,
		CreatedAt:               l.CreatedAt.Time,
		UpdatedAt:               l.UpdatedAt.Time,
		DeletedAt:               l.DeletedAt.Time,
	}
}
