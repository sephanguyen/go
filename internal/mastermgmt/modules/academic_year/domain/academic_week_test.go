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

func TestAcademicWeek_Build(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	repo := new(mock_academic_year_repo.MockAcademicWeekRepo)
	locationRepo := new(mock_location_repo.MockLocationRepo)

	tcs := []struct {
		name         string
		academicWeek *domain.AcademicWeek
		setup        func(ctx context.Context)
		expectedErr  error
	}{
		{
			name: "full fields",
			academicWeek: &domain.AcademicWeek{
				AcademicWeekID: "academic_week_id",
				AcademicYearID: "academic_year_id",
				Name:           "Week 1",
				StartDate:      now,
				EndDate:        now.Add(24 * 7 * time.Hour),
				Period:         "Term 1",
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
			name: "missing start date",
			academicWeek: &domain.AcademicWeek{
				AcademicWeekID: "academic_week_id",
				AcademicYearID: "academic_year_id",
				Name:           "Week 1",
				// StartDate:      now,
				EndDate:      now.Add(24 * 7 * time.Hour),
				Period:       "Term 1",
				LocationID:   "location_id",
				CreatedAt:    now,
				UpdatedAt:    now,
				Repo:         repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid academic week: %w", fmt.Errorf("AcademicWeek.StartDate cannot be empty")),
		},
		{
			name: "missing end date",
			academicWeek: &domain.AcademicWeek{
				AcademicWeekID: "academic_week_id",
				AcademicYearID: "academic_year_id",
				Name:           "Week 1",
				StartDate:      now,
				// EndDate:      now.Add(24 * 7 * time.Hour),
				Period:       "Term 1",
				LocationID:   "location_id",
				CreatedAt:    now,
				UpdatedAt:    now,
				Repo:         repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid academic week: %w", fmt.Errorf("AcademicWeek.EndDate cannot be empty")),
		},
		{
			name: "missing academic year id",
			academicWeek: &domain.AcademicWeek{
				AcademicWeekID: "academic_week_id",
				// AcademicYearID: "academic_year_id",
				Name:         "Week 1",
				StartDate:    now,
				EndDate:      now.Add(24 * 7 * time.Hour),
				Period:       "Term 1",
				LocationID:   "location_id",
				CreatedAt:    now,
				UpdatedAt:    now,
				Repo:         repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid academic week: %w", fmt.Errorf("AcademicWeek.AcademicYearID cannot be empty")),
		},
		{
			name: "missing name",
			academicWeek: &domain.AcademicWeek{
				AcademicWeekID: "academic_week_id",
				AcademicYearID: "academic_year_id",
				// Name:           "Week 1",
				StartDate:    now,
				EndDate:      now.Add(24 * 7 * time.Hour),
				Period:       "Term 1",
				LocationID:   "location_id",
				CreatedAt:    now,
				UpdatedAt:    now,
				Repo:         repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid academic week: %w", fmt.Errorf("AcademicWeek.Name cannot be empty")),
		},
		{
			name: "missing period",
			academicWeek: &domain.AcademicWeek{
				AcademicWeekID: "academic_week_id",
				AcademicYearID: "academic_year_id",
				Name:           "Week 1",
				StartDate:      now,
				EndDate:        now.Add(24 * 7 * time.Hour),
				// Period:       "Term 1",
				LocationID:   "location_id",
				CreatedAt:    now,
				UpdatedAt:    now,
				Repo:         repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid academic week: %w", fmt.Errorf("AcademicWeek.Period cannot be empty")),
		},
		{
			name: "end date is before start date",
			academicWeek: &domain.AcademicWeek{
				AcademicWeekID: "academic_week_id",
				AcademicYearID: "academic_year_id",
				Name:           "Week 1",
				StartDate:      now.Add(24 * 7 * time.Hour),
				EndDate:        now,
				Period:         "Term 1",
				LocationID:     "location_id",
				CreatedAt:      now,
				UpdatedAt:      now,
				Repo:           repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid academic week: %w", fmt.Errorf("AcademicWeek.EndDate cannot before AcademicWeek.StartDate")),
		},
		{
			name: "missing created at",
			academicWeek: &domain.AcademicWeek{
				AcademicWeekID: "academic_week_id",
				AcademicYearID: "academic_year_id",
				Name:           "Week 1",
				StartDate:      now,
				EndDate:        now.Add(24 * 7 * time.Hour),
				Period:         "Term 1",
				LocationID:   "location_id",
				// CreatedAt:    now,
				UpdatedAt:    now,
				Repo:         repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid academic week: %w", fmt.Errorf("AcademicWeek.CreatedAt cannot be empty")),
		},
		{
			name: "missing updated at",
			academicWeek: &domain.AcademicWeek{
				AcademicWeekID: "academic_week_id",
				AcademicYearID: "academic_year_id",
				Name:           "Week 1",
				StartDate:      now,
				EndDate:        now.Add(24 * 7 * time.Hour),
				Period:         "Term 1",
				LocationID:   "location_id",
				CreatedAt:    now,
				// UpdatedAt:    now,
				Repo:         repo,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: fmt.Errorf("invalid academic week: %w", fmt.Errorf("AcademicWeek.UpdatedAt cannot be empty")),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			builder := domain.NewAcademicWeekBuilder().
				WithAcademicWeekID(tc.academicWeek.AcademicWeekID).
				WithAcademicWeekRepo(repo).
				WithAcademicYearID(tc.academicWeek.AcademicYearID).
				WithName(tc.academicWeek.Name).
				WithStartDate(tc.academicWeek.StartDate).
				WithEndDate(tc.academicWeek.EndDate).
				WithPeriod(tc.academicWeek.Period).
				WithLocationID(tc.academicWeek.LocationID).
				WithModificationTime(tc.academicWeek.CreatedAt, tc.academicWeek.UpdatedAt)

			actual, err := builder.Build()
			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.academicWeek, actual)
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
