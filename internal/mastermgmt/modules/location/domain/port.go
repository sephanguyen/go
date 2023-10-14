package domain

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type LocationRepo interface {
	DeleteByPartnerInternalIDs(ctx context.Context, db database.Ext, id pgtype.TextArray) error
	GetLocationByID(ctx context.Context, db database.Ext, id string) (*Location, error)
	GetLocationByPartnerInternalID(ctx context.Context, db database.Ext, id string) (*Location, error)
	GetLocationsByPartnerInternalIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray) ([]*Location, error)
	GetLocationsByLocationIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray, allowDeleted bool) ([]*Location, error)
	GetLocationByLocationTypeName(ctx context.Context, db database.Ext, name string) ([]*Location, error)
	GetLocationByLocationTypeID(ctx context.Context, db database.Ext, id string) ([]*Location, error)
	UpdateAccessPath(ctx context.Context, db database.Ext, ids []string) error
	UpsertLocations(ctx context.Context, db database.Ext, locations []*Location) error
	GetLocationByLocationTypeIDs(ctx context.Context, db database.Ext, ids []string) ([]*Location, error)
	GetChildLocations(ctx context.Context, db database.Ext, id string) ([]*Location, error)
}

type LocationTypeRepo interface {
	GetLocationTypeByID(ctx context.Context, db database.Ext, id string) (*LocationType, error)
	GetLocationTypeByName(ctx context.Context, db database.Ext, name string, allowEmpty bool) (*LocationType, error)
	GetLocationTypeByNames(ctx context.Context, db database.Ext, names pgtype.TextArray) ([]*LocationType, error)
	GetLocationTypeByIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray, allowDeleted bool) ([]*LocationType, error)
	GetLocationTypeByParentName(ctx context.Context, db database.Ext, parentName string) (*LocationType, error)
	GetLocationTypeByNameAndParent(ctx context.Context, db database.Ext, name, parentName string) (*LocationType, error)
	DeleteByPartnerNames(ctx context.Context, db database.Ext, names pgtype.TextArray) error
	UpsertLocationTypes(ctx context.Context, db database.Ext, locationTypes map[int]*LocationType) (errors []*UpsertErrors)
	GetLocationTypesByLevel(ctx context.Context, db database.Ext, level string) ([]*LocationType, error)
}
