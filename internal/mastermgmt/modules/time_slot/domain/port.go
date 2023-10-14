package domain

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	location_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
)

type TimeSlotRepo interface {
	Upsert(ctx context.Context, db database.QueryExecer, weeks []*TimeSlot, locationIDs []string) error
}

type LocationRepo interface {
	GetChildLocations(ctx context.Context, db database.Ext, id string) ([]*location_domain.Location, error)
}
