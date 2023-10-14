package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLocationType_IsValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	repo := new(mock_repositories.MockLocationTypeRepo)
	locationRepo := new(mock_repositories.MockLocationRepo)

	tcs := []struct {
		name         string
		locationType *domain.LocationType
		setup        func(ctx context.Context)
		isValid      bool
	}{
		{
			name: "full fields",
			locationType: &domain.LocationType{
				LocationTypeID: "location-type-id-1",
				Name:           "location type name 1",
				DisplayName:    "location type display",
				CreatedAt:      now,
				UpdatedAt:      now,
				Persisted:      true,
			},
			setup: func(ctx context.Context) {
				repo.On("GetLocationTypeByName", ctx, db, "location type name 1", true).Return(&domain.LocationType{
					LocationTypeID: "location-type-id-1",
					Name:           "location type name 1",
					DisplayName:    "location type display",
					CreatedAt:      now,
					UpdatedAt:      now,
				}, nil).Once()
			},
			isValid: true,
		},
		{
			name: "missing name",
			locationType: &domain.LocationType{
				LocationTypeID: "location-type-id-1",
				Name:           "",
				DisplayName:    "location type display",
				CreatedAt:      now,
				UpdatedAt:      now,
				Persisted:      true,
			},
			setup: func(ctx context.Context) {
			},
			isValid: false,
		},
		{
			name: "missing display name",
			locationType: &domain.LocationType{
				LocationTypeID: "location-type-id-1",
				Name:           "location type name 1",
				DisplayName:    "",
				CreatedAt:      now,
				UpdatedAt:      now,
				Persisted:      true,
			},
			setup: func(ctx context.Context) {
				repo.On("GetLocationTypeByName", ctx, db, "location type name 1", true).Return(&domain.LocationType{
					LocationTypeID: "location-type-id-1",
					Name:           "location type name 1",
					DisplayName:    "location type display",
					CreatedAt:      now,
					UpdatedAt:      now,
				}, nil).Once()
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			builder, _ := domain.NewLocationType().
				WithModificationTime(tc.locationType.CreatedAt, tc.locationType.UpdatedAt).
				WithLocationTypeRepo(repo).
				WithLocationRepo(locationRepo).
				WithDisplayName(tc.locationType.DisplayName).
				WithName(ctx, db, tc.locationType.Name)
			actual, err := builder.Build(ctx, db)
			if !tc.isValid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				actual.Repo = nil
				actual.LocationRepo = nil
				assert.EqualValues(t, tc.locationType, actual)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				repo,
				locationRepo,
			)
		})
	}
}
