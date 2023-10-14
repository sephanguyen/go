package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"

	"github.com/jackc/pgtype"
)

type LocationRepo interface {
	GetAllLocations(ctx context.Context, db database.Ext) ([]*repo.Location, error)
	GetAllRawLocations(ctx context.Context, db database.Ext) ([]*domain.Location, error)
	GetLocationByID(ctx context.Context, db database.Ext, id string) (*domain.Location, error)
	RetrieveLocations(ctx context.Context, db database.Ext, queries domain.FilterLocation) ([]*domain.Location, error)
	UpsertLocations(ctx context.Context, db database.Ext, locations []*domain.Location) error
	GetLocationByPartnerInternalID(ctx context.Context, db database.Ext, id string) (*domain.Location, error)
	DeleteByPartnerInternalIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray) error
	GetLocationsByPartnerInternalIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray) ([]*domain.Location, error)
	GetLocationsByLocationIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray, allowDeleted bool) ([]*domain.Location, error)
	GetLocationByLocationTypeName(ctx context.Context, db database.Ext, name string) ([]*domain.Location, error)
	GetLocationByLocationTypeID(ctx context.Context, db database.Ext, id string) ([]*domain.Location, error)
	UpdateAccessPath(ctx context.Context, db database.Ext, ids []string) error
	RetrieveLowestLevelLocations(ctx context.Context, db database.Ext, params *repo.GetLowestLevelLocationsParams) ([]*domain.Location, error)
	GetLowestLevelLocationsV2(ctx context.Context, db database.Ext, params *repo.GetLowestLevelLocationsParams) ([]*domain.Location, error)
	GetLocationByLocationTypeIDs(ctx context.Context, db database.Ext, ids []string) ([]*domain.Location, error)
	GetChildLocations(ctx context.Context, db database.Ext, id string) ([]*domain.Location, error)
	GetRootLocation(ctx context.Context, db database.Ext) (string, error)
}

type LocationTypeRepo interface {
	GetLocationTypeByID(ctx context.Context, db database.Ext, id string) (*domain.LocationType, error)
	GetLocationTypeByName(ctx context.Context, db database.Ext, name string, allowEmpty bool) (*domain.LocationType, error)
	UpsertLocationTypes(ctx context.Context, db database.Ext, locationTypes map[int]*domain.LocationType) (errors []*domain.UpsertErrors)
	Import(ctx context.Context, db database.Ext, locationTypes []*domain.LocationType) error
	GetLocationTypeByNames(ctx context.Context, db database.Ext, names pgtype.TextArray) ([]*domain.LocationType, error)
	DeleteByPartnerNames(ctx context.Context, db database.Ext, names pgtype.TextArray) error
	GetLocationTypeByIDs(ctx context.Context, db database.Ext, ids pgtype.TextArray, allowDeleted bool) ([]*domain.LocationType, error)
	RetrieveLocationTypes(ctx context.Context, db database.Ext) ([]*domain.LocationType, error)
	RetrieveLocationTypesV2(ctx context.Context, db database.Ext) ([]*domain.LocationType, error)
	GetAllLocationTypes(ctx context.Context, db database.Ext) ([]*repo.LocationType, error)
	GetLocationTypeByParentName(ctx context.Context, db database.Ext, parentName string) (*domain.LocationType, error)
	GetLocationTypeByNameAndParent(ctx context.Context, db database.Ext, name, parentName string) (*domain.LocationType, error)
	UpdateLevels(ctx context.Context, db database.Ext) error
	GetLocationTypesByLevel(ctx context.Context, db database.Ext, level string) ([]*domain.LocationType, error)
}

type ImportLogRepo interface {
	Create(ctx context.Context, db database.QueryExecer, e *domain.ImportLog) error
}
