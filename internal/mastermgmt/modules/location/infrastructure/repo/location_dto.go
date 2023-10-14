package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
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
	ResourcePath            pgtype.Text
	UpdatedAt               pgtype.Timestamptz
	CreatedAt               pgtype.Timestamptz
	DeletedAt               pgtype.Timestamptz
}

func (l *Location) FieldMapWithoutRP() (fields []string, values []interface{}) {
	fields = []string{"location_id", "name", "location_type", "parent_location_id", "partner_internal_id", "partner_internal_parent_id", "is_archived", "updated_at", "created_at", "deleted_at", "access_path"}
	values = []interface{}{&l.LocationID, &l.Name, &l.LocationType, &l.ParentLocationID, &l.PartnerInternalID, &l.PartnerInternalParentID, &l.IsArchived, &l.UpdatedAt, &l.CreatedAt, &l.DeletedAt, &l.AccessPath}
	return
}

func (l *Location) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"location_id", "name", "location_type", "parent_location_id", "partner_internal_id", "partner_internal_parent_id", "is_archived", "updated_at", "created_at", "deleted_at", "access_path", "resource_path"}
	values = []interface{}{&l.LocationID, &l.Name, &l.LocationType, &l.ParentLocationID, &l.PartnerInternalID, &l.PartnerInternalParentID, &l.IsArchived, &l.UpdatedAt, &l.CreatedAt, &l.DeletedAt, &l.AccessPath, &l.ResourcePath}
	return
}

func (*Location) TableName() string {
	return "locations"
}

func (l *Location) ToLocationEntity() *domain.Location {
	location := &domain.Location{
		LocationID:              l.LocationID.String,
		Name:                    l.Name.String,
		ParentLocationID:        l.ParentLocationID.String,
		PartnerInternalID:       l.PartnerInternalID.String,
		PartnerInternalParentID: l.PartnerInternalParentID.String,
		LocationType:            l.LocationType.String,
		IsArchived:              l.IsArchived.Bool,
		AccessPath:              l.AccessPath.String,
		ResourcePath:            l.ResourcePath.String,
		CreatedAt:               l.CreatedAt.Time,
		UpdatedAt:               l.UpdatedAt.Time,
	}
	if l.DeletedAt.Status == pgtype.Present {
		location.DeletedAt = &l.DeletedAt.Time
	}
	return location
}

func NewLocationFromEntity(l *domain.Location) (*Location, error) {
	dto := &Location{}
	database.AllNullEntity(dto)
	parentID := l.ParentLocationID
	if err := multierr.Combine(
		dto.LocationID.Set(l.LocationID),
		dto.Name.Set(l.Name),
		dto.ParentLocationID.Set(parentID),
		dto.PartnerInternalID.Set(CheckNilString(l.PartnerInternalID)),
		dto.PartnerInternalParentID.Set(CheckNilString(l.PartnerInternalParentID)),
		dto.ParentLocationID.Set(CheckNilString(l.ParentLocationID)),
		dto.LocationType.Set(l.LocationType),
		dto.IsArchived.Set(l.IsArchived),
		dto.AccessPath.Set(l.AccessPath),
		dto.ResourcePath.Set(l.ResourcePath),
		dto.CreatedAt.Set(l.CreatedAt),
		dto.UpdatedAt.Set(l.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from location entity to location dto: %w", err)
	}

	return dto, nil
}

func CheckNilString(a string) interface{} {
	if len(a) == 0 {
		return nil
	}
	return a
}

func (l *Location) String() string {
	if l == nil {
		return "nil"
	}
	return fmt.Sprintf("[Name: %s, PartnerID: %s, ParentPartnerID: %s, ID: %s, LocationType: %s, ParentID: %s]",
		l.Name.String, l.PartnerInternalID.String, l.PartnerInternalParentID.String, l.LocationID.String, l.LocationType.String, l.ParentLocationID.String)
}
