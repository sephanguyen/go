package entities

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	mock_repositories "github.com/manabie-com/backend/mock/calendar/repositories"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestScheduler_Validate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name      string
		scheduler *Scheduler
		hasError  bool
	}{
		{
			name: "scheduler is valid",
			scheduler: &Scheduler{
				SchedulerID: "scheduler-id",
				StartDate:   time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
				Frequency:   constants.FrequencyWeekly,
			},
		},
		{
			name: "start date greater than end date",
			scheduler: &Scheduler{
				SchedulerID: "scheduler-id",
				StartDate:   time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				Frequency:   constants.FrequencyWeekly,
			},
			hasError: true,
		},
		{
			name: "empty frequency",
			scheduler: &Scheduler{
				SchedulerID: "scheduler-id",
				StartDate:   time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.scheduler.Validate()
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				require.False(t, tc.hasError)
			}
		})
	}
}

func TestScheduler_Create(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	schedulerRepo := &mock_repositories.MockSchedulerRepo{}

	testCases := []struct {
		name      string
		scheduler *Scheduler
		setup     func(context.Context)
		hasError  bool
	}{
		{
			name: "create scheduler successfully",
			scheduler: &Scheduler{
				SchedulerID: "scheduler-id",
				StartDate:   time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				EndDate:     time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
				Frequency:   constants.FrequencyWeekly,
			},
			setup: func(ctx context.Context) {
				schedulerRepo.On("Create", mock.Anything, mockDB.DB, mock.MatchedBy(func(sch *dto.CreateSchedulerParams) bool {
					if !sch.StartDate.Equal(time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC)) {
						return false
					}
					if !sch.EndDate.Equal(time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC)) {
						return false
					}
					return true
				})).Once().Return("scheduler-id", nil)
			},
		},
		{
			name: "failed to create scheduler since scheduler is invalid",
			scheduler: &Scheduler{
				SchedulerID: "scheduler-id",
				StartDate:   time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				Frequency:   constants.FrequencyWeekly,
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			tc.scheduler.SchedulerRepo = schedulerRepo
			schedulerId, err := tc.scheduler.Create(ctx, mockDB.DB)
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				require.False(t, tc.hasError)
				require.Equal(t, tc.scheduler.SchedulerID, schedulerId)
			}
		})
	}
}

func TestScheduler_Update(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	schedulerRepo := &mock_repositories.MockSchedulerRepo{}

	testCases := []struct {
		name      string
		scheduler *Scheduler
		setup     func(context.Context)
		hasError  bool
	}{
		{
			name: "update scheduler successfully",
			scheduler: &Scheduler{
				SchedulerID: "scheduler-id",
				EndDate:     time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
			},
			setup: func(ctx context.Context) {
				schedulerRepo.On("Update", mock.Anything, mockDB.DB, mock.MatchedBy(func(sch *dto.UpdateSchedulerParams) bool {
					if sch.SchedulerID != "scheduler-id" {
						return false
					}
					if !sch.EndDate.Equal(time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC)) {
						return false
					}
					return true
				}), mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "failed to update scheduler since empty end date",
			scheduler: &Scheduler{
				SchedulerID: "scheduler-id",
			},
			setup:    func(ctx context.Context) {},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			tc.scheduler.SchedulerRepo = schedulerRepo
			err := tc.scheduler.Update(ctx, mockDB.DB)
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				require.False(t, tc.hasError)
			}
		})
	}
}
