package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	location_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/domain"
)

type TimeSlotRepo interface {
	Upsert(ctx context.Context, db database.QueryExecer, weeks []*domain.TimeSlot, locationIDs []string) error
}

type LocationRepo interface {
	GetChildLocations(ctx context.Context, db database.Ext, id string) ([]*location_domain.Location, error)
}
