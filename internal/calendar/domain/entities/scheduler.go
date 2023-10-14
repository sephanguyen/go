package entities

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type Scheduler struct {
	SchedulerID   string
	StartDate     time.Time
	EndDate       time.Time
	Frequency     constants.Frequency
	SchedulerRepo infrastructure.SchedulerPort
}

func NewScheduler(startDate, endDate time.Time, freq constants.Frequency, repo infrastructure.SchedulerPort) *Scheduler {
	sch := &Scheduler{
		SchedulerID:   idutil.ULIDNow(),
		StartDate:     startDate,
		EndDate:       endDate,
		Frequency:     freq,
		SchedulerRepo: repo,
	}
	return sch
}

func (sch *Scheduler) Validate() error {
	if len(sch.Frequency) == 0 {
		return fmt.Errorf("Scheduler.Frequency cannot be empty")
	}

	if sch.StartDate.IsZero() {
		return fmt.Errorf("start date could not be empty")
	}

	if sch.EndDate.IsZero() {
		return fmt.Errorf("end date could not be empty")
	}

	if sch.EndDate.Before(sch.StartDate) {
		return fmt.Errorf("end date could not before start date")
	}

	return nil
}

func (sch *Scheduler) Create(ctx context.Context, db database.QueryExecer) (string, error) {
	if err := sch.Validate(); err != nil {
		return "", err
	}
	schedulerID, err := sch.SchedulerRepo.Create(
		ctx,
		db,
		&dto.CreateSchedulerParams{
			SchedulerID: sch.SchedulerID,
			StartDate:   sch.StartDate,
			EndDate:     sch.EndDate,
			Frequency:   string(sch.Frequency),
		})
	return schedulerID, err
}

func (sch *Scheduler) Update(ctx context.Context, db database.QueryExecer) error {
	if len(sch.SchedulerID) == 0 {
		return fmt.Errorf("scheduler_id could not be empty")
	}
	if sch.EndDate.IsZero() {
		return fmt.Errorf("end_date could not be empty")
	}
	err := sch.SchedulerRepo.Update(
		ctx,
		db,
		&dto.UpdateSchedulerParams{
			SchedulerID: sch.SchedulerID,
			EndDate:     sch.EndDate,
		}, []string{"end_date"})
	return err
}

func (sch *Scheduler) Get(ctx context.Context, db database.QueryExecer) (*dto.Scheduler, error) {
	return sch.SchedulerRepo.GetByID(ctx, db, sch.SchedulerID)
}
