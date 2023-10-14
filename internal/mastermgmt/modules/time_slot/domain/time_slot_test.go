package domain_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_location_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mock_time_slot_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/time_slot/infrastructure/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTimeSlot_Build(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	repo := new(mock_time_slot_repo.MockTimeSlotRepo)
	locationRepo := new(mock_location_repo.MockLocationRepo)

	tcs := []struct {
		name         string
		timeSlot *domain.TimeSlot
		setup        func(ctx context.Context)
		expectedErr  error
	}{
		{
			name: "full fields",
			timeSlot: &domain.TimeSlot{
				TimeSlotID:         "time_slot_01",
				TimeSlotInternalID: "1",
				StartTime:          "11:00",
				EndTime:            "13:00",
				LocationID:         "location_id",
				CreatedAt:          now,
				UpdatedAt:          now,
				Repo:               repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: nil,
		},
		{
			name: "StartTime is not in format HH:mm",
			timeSlot: &domain.TimeSlot{
				TimeSlotID:         "time_slot_01",
				TimeSlotInternalID: "1",
				StartTime:          "11:77",
				EndTime:            "13:00",
				LocationID:         "location_id",
				CreatedAt:          now,
				UpdatedAt:          now,
				Repo:               repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid time slot: %w", fmt.Errorf("TimeSlot.StartTime is not valid time format")),
		},
		{
			name: "EndTime is not in format HH:mm",
			timeSlot: &domain.TimeSlot{
				TimeSlotID:         "time_slot_01",
				TimeSlotInternalID: "1",
				StartTime:          "11:00",
				EndTime:            "25:00",
				LocationID:         "location_id",
				CreatedAt:          now,
				UpdatedAt:          now,
				Repo:               repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid time slot: %w", fmt.Errorf("TimeSlot.EndTime is not valid time format")),
		},
		{
			name: "missing StartTime",
			timeSlot: &domain.TimeSlot{
				TimeSlotID:         "time_slot_01",
				TimeSlotInternalID: "1",
				// StartTime:          "11:00",
				EndTime:            "13:00",
				LocationID:         "location_id",
				CreatedAt:          now,
				UpdatedAt:          now,
				Repo:               repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid time slot: %w", fmt.Errorf("TimeSlot.StartTime cannot be empty")),
		},
		{
			name: "missing EndTime",
			timeSlot: &domain.TimeSlot{
				TimeSlotID:         "time_slot_01",
				TimeSlotInternalID: "1",
				StartTime:          "11:00",
				// EndTime:            "13:00",
				LocationID:         "location_id",
				CreatedAt:          now,
				UpdatedAt:          now,
				Repo:               repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid time slot: %w", fmt.Errorf("TimeSlot.EndTime cannot be empty")),
		},
		{
			name: "missing created at",
			timeSlot: &domain.TimeSlot{
				TimeSlotID:         "time_slot_01",
				TimeSlotInternalID: "1",
				StartTime:          "11:00",
				EndTime:            "13:00",
				LocationID:         "location_id",
				// CreatedAt:    now,
				UpdatedAt: now,
				Repo:      repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid time slot: %w", fmt.Errorf("TimeSlot.CreatedAt cannot be empty")),
		},
		{
			name: "missing updated at",
			timeSlot: &domain.TimeSlot{
				TimeSlotID:         "time_slot_01",
				TimeSlotInternalID: "1",
				StartTime:          "11:00",
				EndTime:            "13:00",
				LocationID:         "location_id",
				CreatedAt:          now,
				// UpdatedAt:    now,
				Repo: repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid time slot: %w", fmt.Errorf("TimeSlot.UpdatedAt cannot be empty")),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			builder := domain.NewTimeSlotBuilder().
				WithTimeSlotRepo(repo).
				WithTimeSlotID(tc.timeSlot.TimeSlotID).
				WithTimeSlotInternalID(tc.timeSlot.TimeSlotInternalID).
				WithStartTime(tc.timeSlot.StartTime).
				WithEndTime(tc.timeSlot.EndTime).
				WithLocationID(tc.timeSlot.LocationID).
				WithModificationTime(tc.timeSlot.CreatedAt, tc.timeSlot.UpdatedAt)

			actual, err := builder.BuildWithoutPKCheck()
			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.timeSlot, actual)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				repo,
				repo,
			)
			mock.AssertExpectationsForObjects(
				t,
				db,
				repo,
				locationRepo,
			)
		})
	}
}
