package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	configuration_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/domain"
	location_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/domain"
)

type WorkingHoursRepo interface {
	Upsert(ctx context.Context, db database.QueryExecer, workingHours []*domain.WorkingHours, locationIDs []string) error
	GetWorkingHoursByID(ctx context.Context, db database.Ext, id string) (*domain.WorkingHours, error)
}

type LocationRepo interface {
	GetLocationByID(ctx context.Context, db database.Ext, id string) (*location_domain.Location, error)
	GetLocationByLocationTypeIDs(ctx context.Context, db database.Ext, ids []string) ([]*location_domain.Location, error)
	GetChildLocations(ctx context.Context, db database.Ext, id string) ([]*location_domain.Location, error)
}

type LocationTypeRepo interface {
	GetLocationTypesByLevel(ctx context.Context, db database.Ext, level string) ([]*location_domain.LocationType, error)
}

type ConfigRepo interface {
	GetByKey(ctx context.Context, db database.QueryExecer, cKey string) (c *configuration_domain.InternalConfiguration, err error)
}
