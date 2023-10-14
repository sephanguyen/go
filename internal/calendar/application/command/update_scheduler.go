package command

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/entities"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type UpdateSchedulerCommand struct {
	SchedulerRepo infrastructure.SchedulerPort
}

type UpdateSchedulerRequest struct {
	SchedulerID string
	EndDate     time.Time
}

func (usc *UpdateSchedulerCommand) UpdateScheduler(ctx context.Context, db database.QueryExecer, req *UpdateSchedulerRequest) error {
	scheduler := &entities.Scheduler{
		SchedulerID:   req.SchedulerID,
		EndDate:       req.EndDate,
		SchedulerRepo: usc.SchedulerRepo,
	}
	err := scheduler.Update(ctx, db)
	if err != nil {
		return err
	}
	return nil
}
