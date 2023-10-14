package domain_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_academic_year_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/academic_year/infrastructure/repo"
	mock_location_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAcademicClosedDay_Build(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	repo := new(mock_academic_year_repo.MockAcademicClosedDayRepo)
	locationRepo := new(mock_location_repo.MockLocationRepo)

	tcs := []struct {
		name              string
		academicClosedDay *domain.AcademicClosedDay
		setup             func(ctx context.Context)
		expectedErr       error
	}{
		{
			name: "full fields",
			academicClosedDay: &domain.AcademicClosedDay{
				Date:                now,
				AcademicClosedDayID: "academic_closed_day_id",
				AcademicWeekID:      "academic_week_id",
				AcademicYearID:      "academic_year_id",
				CreatedAt:           now,
				UpdatedAt:           now,
				Repo:                repo,
				LocationID:          "location_id",
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: nil,
		},
		{
			name: "missing date",
			academicClosedDay: &domain.AcademicClosedDay{
				// Date:           now,
				AcademicClosedDayID: "academic_closed_day_id",
				AcademicWeekID:      "academic_week_id",
				AcademicYearID:      "academic_year_id",
				CreatedAt:           now,
				UpdatedAt:           now,
				Repo:                repo,
				LocationID:          "location_id",
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid closed day: %w", fmt.Errorf("AcademicClosedDay.Date cannot be empty")),
		},
		{
			name: "missing created at",
			academicClosedDay: &domain.AcademicClosedDay{
				Date:                now,
				AcademicClosedDayID: "academic_closed_day_id",
				AcademicWeekID:      "academic_week_id",
				AcademicYearID:      "academic_year_id",
				// CreatedAt:      now,
				UpdatedAt:    now,
				Repo:         repo,
				LocationID:   "location_id",
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid closed day: %w", fmt.Errorf("AcademicClosedDay.CreatedAt cannot be empty")),
		},
		{
			name: "missing updated at",
			academicClosedDay: &domain.AcademicClosedDay{
				Date:                now,
				AcademicClosedDayID: "academic_closed_day_id",
				AcademicWeekID:      "academic_week_id",
				AcademicYearID:      "academic_year_id",
				CreatedAt:           now,
				// UpdatedAt:      now,
				Repo:         repo,
				LocationID:   "location_id",
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid closed day: %w", fmt.Errorf("AcademicClosedDay.UpdatedAt cannot be empty")),
		},
		{
			name: "missing AcademicWeekID",
			academicClosedDay: &domain.AcademicClosedDay{
				Date:                now,
				AcademicClosedDayID: "academic_closed_day_id",
				// AcademicWeekID: "academic_week_id",
				AcademicYearID: "academic_year_id",
				CreatedAt:      now,
				UpdatedAt:      now,
				Repo:           repo,
				LocationID:     "location_id",
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: nil,
		},
		{
			name: "missing AcademicYearID",
			academicClosedDay: &domain.AcademicClosedDay{
				Date:                now,
				AcademicClosedDayID: "academic_closed_day_id",
				AcademicWeekID:      "academic_week_id",
				// AcademicYearID: "academic_year_id",
				CreatedAt:    now,
				UpdatedAt:    now,
				Repo:         repo,
				LocationID:   "location_id",
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid closed day: %w", fmt.Errorf("AcademicClosedDay.AcademicYearID cannot be empty")),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			builder := domain.NewAcademicClosedDayBuilder().
				WithAcademicClosedDayID(tc.academicClosedDay.AcademicClosedDayID).
				WithAcademicClosedDayRepo(repo).
				WithDate(tc.academicClosedDay.Date).
				WithAcademicWeekID(tc.academicClosedDay.AcademicWeekID).
				WithAcademicYearID(tc.academicClosedDay.AcademicYearID).
				WithLocationID(tc.academicClosedDay.LocationID).
				WithModificationTime(tc.academicClosedDay.CreatedAt, tc.academicClosedDay.UpdatedAt)

			actual, err := builder.Build(false)
			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.academicClosedDay, actual)
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
