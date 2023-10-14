package domain

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type WorkingHoursRepo interface {
	Upsert(ctx context.Context, db database.QueryExecer, workingHours []*WorkingHours, locationIDs []string) error
	GetWorkingHoursByID(ctx context.Context, db database.Ext, id string) (*WorkingHours, error)
}
