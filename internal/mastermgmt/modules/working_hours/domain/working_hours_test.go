package domain_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_location_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mock_working_hours_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/working_hours/infrastructure/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWorkingHours_Build(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	repo := new(mock_working_hours_repo.MockWorkingHoursRepo)
	locationRepo := new(mock_location_repo.MockLocationRepo)

	tcs := []struct {
		name         string
		workingHours *domain.WorkingHours
		setup        func(ctx context.Context)
		expectedErr  error
	}{
		{
			name: "full fields",
			workingHours: &domain.WorkingHours{
				WorkingHoursID: "working_hour_id",
				Day:            "Monday",
				OpeningTime:    "08:00",
				ClosingTime:    "17:00",
				LocationID:     "location_id",
				CreatedAt:      now,
				UpdatedAt:      now,
				Repo:           repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: nil,
		},
		{
			name: "opening time is not in format HH:mm",
			workingHours: &domain.WorkingHours{
				WorkingHoursID: "working_hour_id",
				Day:            "Monday",
				OpeningTime:    "34:99",
				ClosingTime:    "17:00",
				LocationID:     "location_id",
				CreatedAt:      now,
				UpdatedAt:      now,
				Repo:           repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid working hours: %w", fmt.Errorf("WorkingHours.OpeningTime is not valid time format")),
		},
		{
			name: "closing time is not in format HH:mm",
			workingHours: &domain.WorkingHours{
				WorkingHoursID: "working_hour_id",
				Day:            "Monday",
				OpeningTime:    "08:00",
				ClosingTime:    "017:000",
				LocationID:     "location_id",
				CreatedAt:      now,
				UpdatedAt:      now,
				Repo:           repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid working hours: %w", fmt.Errorf("WorkingHours.ClosingTime is not valid time format")),
		},
		{
			name: "missing Day",
			workingHours: &domain.WorkingHours{
				WorkingHoursID: "working_hour_id",
				// Day:            "Monday",
				OpeningTime: "08:00",
				ClosingTime: "17:00",
				LocationID:  "location_id",
				CreatedAt:   now,
				UpdatedAt:   now,
				Repo:        repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid working hours: %w", fmt.Errorf("WorkingHours.Day cannot be empty")),
		},
		{
			name: "missing opening time",
			workingHours: &domain.WorkingHours{
				WorkingHoursID: "working_hour_id",
				Day:            "Monday",
				// OpeningTime:    "08:00",
				ClosingTime: "17:00",
				LocationID:  "location_id",
				CreatedAt:   now,
				UpdatedAt:   now,
				Repo:        repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid working hours: %w", fmt.Errorf("WorkingHours.OpeningTime cannot be empty")),
		},
		{
			name: "missing closing time",
			workingHours: &domain.WorkingHours{
				WorkingHoursID: "working_hour_id",
				Day:            "Monday",
				OpeningTime:    "08:00",
				// ClosingTime:    "17:00",
				LocationID: "location_id",
				CreatedAt:  now,
				UpdatedAt:  now,
				Repo:       repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid working hours: %w", fmt.Errorf("WorkingHours.ClosingTime cannot be empty")),
		},
		{
			name: "missing created at",
			workingHours: &domain.WorkingHours{
				WorkingHoursID: "working_hour_id",
				Day:            "Monday",
				OpeningTime:    "08:00",
				ClosingTime:    "17:00",
				LocationID:     "location_id",
				// CreatedAt:    now,
				UpdatedAt: now,
				Repo:      repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid working hours: %w", fmt.Errorf("WorkingHours.CreatedAt cannot be empty")),
		},
		{
			name: "missing updated at",
			workingHours: &domain.WorkingHours{
				WorkingHoursID: "working_hour_id",
				Day:            "Monday",
				OpeningTime:    "08:00",
				ClosingTime:    "17:00",
				LocationID:     "location_id",
				CreatedAt:      now,
				// UpdatedAt:    now,
				Repo: repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid working hours: %w", fmt.Errorf("WorkingHours.UpdatedAt cannot be empty")),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			builder := domain.NewWorkingHoursBuilder().
				WithWorkingHoursRepo(repo).
				WithWorkingHoursID(tc.workingHours.WorkingHoursID).
				WithDay(tc.workingHours.Day).
				WithOpeningTime(tc.workingHours.OpeningTime).
				WithClosingTime(tc.workingHours.ClosingTime).
				WithLocationID(tc.workingHours.LocationID).
				WithModificationTime(tc.workingHours.CreatedAt, tc.workingHours.UpdatedAt)

			actual, err := builder.BuildWithoutPKCheck()
			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.workingHours, actual)
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
