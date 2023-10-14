package domain

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"

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
	UpdatedAt               pgtype.Timestamptz
	CreatedAt               pgtype.Timestamptz
	DeletedAt               pgtype.Timestamptz
}

func (l *Location) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"location_id", "name", "location_type", "parent_location_id", "partner_internal_id", "partner_internal_parent_id", "is_archived", "access_path", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&l.LocationID, &l.Name, &l.LocationType, &l.ParentLocationID, &l.PartnerInternalID, &l.PartnerInternalParentID, &l.IsArchived, &l.AccessPath, &l.UpdatedAt, &l.CreatedAt, &l.DeletedAt}
	return
}

func (*Location) TableName() string {
	return "locations"
}

type LocationRepository interface {
	GetLocationByID(ctx context.Context, db database.Ext, id []string) ([]*Location, error)
}
