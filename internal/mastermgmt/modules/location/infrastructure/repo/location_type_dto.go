package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
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
	Level                pgtype.Int4
}

func (l *LocationType) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"location_type_id", "name", "display_name", "parent_name", "parent_location_type_id", "is_archived", "updated_at", "created_at", "deleted_at", "level"}
	values = []interface{}{&l.LocationTypeID, &l.Name, &l.DisplayName, &l.ParentName, &l.ParentLocationTypeID, &l.IsArchived, &l.UpdatedAt, &l.CreatedAt, &l.DeletedAt, &l.Level}
	return
}

func (*LocationType) TableName() string {
	return "location_types"
}

func (l *LocationType) ToLocationTypeEntity() *domain.LocationType {
	locationType := &domain.LocationType{
		LocationTypeID:       l.LocationTypeID.String,
		Name:                 l.Name.String,
		DisplayName:          l.DisplayName.String,
		ParentName:           l.ParentName.String,
		ParentLocationTypeID: l.ParentLocationTypeID.String,
		Level:                int(l.Level.Int),
		IsArchived:           l.IsArchived.Bool,
		CreatedAt:            l.CreatedAt.Time,
		UpdatedAt:            l.UpdatedAt.Time,
	}
	if l.DeletedAt.Status == pgtype.Present {
		locationType.DeletedAt = &l.DeletedAt.Time
	}
	return locationType
}

func NewLocationTypeFromEntity(l *domain.LocationType) (*LocationType, error) {
	dto := &LocationType{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.LocationTypeID.Set(l.LocationTypeID),
		dto.Name.Set(l.Name),
		dto.ParentName.Set(CheckNilString(l.ParentName)),
		dto.DisplayName.Set(l.DisplayName),
		dto.Level.Set(l.Level),
		dto.ParentLocationTypeID.Set(CheckNilString(l.ParentLocationTypeID)),
		dto.IsArchived.Set(l.IsArchived),
		dto.CreatedAt.Set(l.CreatedAt),
		dto.UpdatedAt.Set(l.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from location type entity to location type dto: %w", err)
	}

	return dto, nil
}
