package dto

import "time"

type CreateSchedulerParams struct {
	SchedulerID string
	StartDate   time.Time
	EndDate     time.Time
	Frequency   string
}

type CreateSchedulerParamWithIdentity struct {
	ID                   string
	CreateSchedulerParam CreateSchedulerParams
}

type UpdateSchedulerParams struct {
	SchedulerID string
	EndDate     time.Time
}

type Scheduler struct {
	SchedulerID string
	StartDate   time.Time
	EndDate     time.Time
	Frequency   string
}
