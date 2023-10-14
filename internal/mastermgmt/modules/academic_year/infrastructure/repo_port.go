package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/infrastructure/repo"
	configuration_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/domain"
	location_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
)

type AcademicYearRepo interface {
	Insert(ctx context.Context, db database.QueryExecer, weeks []*domain.AcademicYear) error
	GetAcademicYearByID(ctx context.Context, db database.Ext, id string) (*domain.AcademicYear, error)
}

type AcademicWeekRepo interface {
	Insert(ctx context.Context, db database.QueryExecer, weeks []*domain.AcademicWeek) error
	GetLocationsByAcademicWeekID(ctx context.Context, db database.QueryExecer, academicYearID string) ([]string, error)
	GetAcademicWeeksByYearAndLocationIDs(ctx context.Context, db database.QueryExecer, academicYearID string, locationID []string) ([]*repo.AcademicWeek, error)
}

type AcademicClosedDayRepo interface {
	Insert(ctx context.Context, db database.QueryExecer, weeks []*domain.AcademicClosedDay) error
	GetAcademicClosedDayByWeeks(ctx context.Context, db database.QueryExecer, weekIDs []string) ([]*repo.AcademicClosedDay, error)
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
