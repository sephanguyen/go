package repositories

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type Scheduler struct {
	SchedulerID pgtype.Text
	StartDate   pgtype.Timestamptz
	EndDate     pgtype.Timestamptz
	Frequency   pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
}

func (sch *Scheduler) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"scheduler_id",
		"start_date",
		"end_date",
		"freq",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&sch.SchedulerID,
		&sch.StartDate,
		&sch.EndDate,
		&sch.Frequency,
		&sch.CreatedAt,
		&sch.UpdatedAt,
		&sch.DeletedAt,
	}
	return
}

func (*Scheduler) TableName() string {
	return "scheduler"
}

func (sch *Scheduler) PreUpdate() error {
	return multierr.Combine(
		sch.UpdatedAt.Set(time.Now()),
		sch.DeletedAt.Set(nil),
	)
}

func (sch *Scheduler) PreInsert() error {
	return multierr.Combine(
		sch.CreatedAt.Set(time.Now()),
		sch.UpdatedAt.Set(time.Now()),
	)
}

func NewScheduler(values map[string]interface{}) (*Scheduler, error) {
	schedulerDTO := &Scheduler{}
	database.AllNullEntity(schedulerDTO)
	var err error
	if schedulerID, ok := values["scheduler_id"]; ok {
		err = multierr.Append(err, schedulerDTO.SchedulerID.Set(schedulerID))
	}
	if startDate, ok := values["start_date"]; ok {
		err = multierr.Append(err, schedulerDTO.StartDate.Set(startDate))
	}
	if endDate, ok := values["end_date"]; ok {
		err = multierr.Append(err, schedulerDTO.EndDate.Set(endDate))
	}
	if freq, ok := values["frequency"]; ok {
		err = multierr.Append(err, schedulerDTO.Frequency.Set(freq))
	}
	if createdAt, ok := values["created_at"]; ok {
		err = multierr.Append(err, schedulerDTO.CreatedAt.Set(createdAt))
	}
	if updatedAt, ok := values["updated_at"]; ok {
		err = multierr.Append(err, schedulerDTO.UpdatedAt.Set(updatedAt))
	}
	return schedulerDTO, err
}
