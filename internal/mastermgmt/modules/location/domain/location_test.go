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

func TestLocation_Build(t *testing.T) {
	t.Parallel()
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	repo := new(mock_repositories.MockLocationRepo)
	typeRepo := new(mock_repositories.MockLocationTypeRepo)

	tcs := []struct {
		name         string
		location     *domain.Location
		locationType string
		setup        func(ctx context.Context)
		isValid      bool
	}{
		{
			name: "full fields",
			location: &domain.Location{
				LocationID:              "location_id_1",
				PartnerInternalID:       "partner_1",
				Name:                    "name 1",
				LocationType:            "type_1",
				CreatedAt:               now,
				UpdatedAt:               now,
				Persisted:               true,
				PartnerInternalParentID: "parent_2",
				ParentLocationID:        "location_id_2",
				Repo:                    repo,
				TypeRepo:                typeRepo,
				AccessPath:              "root",
			},
			locationType: "org",
			setup: func(ctx context.Context) {
				repo.On("GetLocationByPartnerInternalID", ctx, db, "partner_1").Return(&domain.Location{
					LocationID:              "location_id_1",
					PartnerInternalID:       "partner_1",
					Name:                    "name 1",
					LocationType:            "type_1",
					ParentLocationID:        "location_id_2",
					PartnerInternalParentID: "parent_2",
					CreatedAt:               now,
					UpdatedAt:               now,
				}, nil).Once()
				repo.On("GetLocationByPartnerInternalID", ctx, db, "parent_2").Return(&domain.Location{
					LocationID:              "location_id_2",
					PartnerInternalID:       "partner_2",
					Name:                    "name 2",
					LocationType:            "type_2",
					ParentLocationID:        "parent_3",
					PartnerInternalParentID: "",
					CreatedAt:               now,
					UpdatedAt:               now,
				}, nil).Once()
				typeRepo.On("GetLocationTypeByID", ctx, db, "type_2").Return(&domain.LocationType{
					LocationTypeID: "type_2",
					Name:           "type_2",
				}, nil).Once()
				typeRepo.On("GetLocationTypeByNameAndParent", ctx, db, "type_1", "type_2").Return(&domain.LocationType{
					LocationTypeID: "type_2",
					Name:           "type_2",
				}, nil).Once()
				typeRepo.On("GetLocationTypeByName", ctx, db, "type_1", false).Return(&domain.LocationType{
					LocationTypeID: "type_1",
					Name:           "type_1",
				}, nil).Once()
			},
			isValid: true,
		},
		{
			name: "missing location id",
			location: &domain.Location{
				LocationID:        "",
				Name:              "location name 1",
				PartnerInternalID: "",
				LocationType:      "brand",
				CreatedAt:         now,
				UpdatedAt:         now,
				Persisted:         true,
				ParentLocationID:  "",
				Repo:              repo,
				TypeRepo:          typeRepo,
			},
			setup: func(ctx context.Context) {
			},
			isValid: false,
		},
		{
			name: "missing location type",
			location: &domain.Location{
				LocationID:        "location-1",
				PartnerInternalID: "partner-1",
				Name:              "location name 1",
				LocationType:      "",
				CreatedAt:         now,
				UpdatedAt:         now,
				Persisted:         true,
				ParentLocationID:  "",
			},
			setup: func(ctx context.Context) {
			},
			isValid: false,
		},
		{
			name: "missing name",
			location: &domain.Location{
				LocationID:        "location-1",
				Name:              "",
				PartnerInternalID: "partner-1",
				LocationType:      "center",
				CreatedAt:         now,
				UpdatedAt:         now,
				Persisted:         true,
				ParentLocationID:  "",
			},
			setup: func(ctx context.Context) {
			},
			isValid: false,
		},
		{
			name: "same PartnerInternalID and PartnerInternalParentID",
			location: &domain.Location{
				LocationID:              "location-2",
				Name:                    "aaa",
				PartnerInternalID:       "location-2",
				LocationType:            "center",
				CreatedAt:               now,
				UpdatedAt:               now,
				PartnerInternalParentID: "location-2",
				Persisted:               true,
				ParentLocationID:        "location-2",
			},
			setup: func(ctx context.Context) {
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			builder := domain.NewLocation().
				WithLocationTypeRepo(typeRepo).
				WithLocationRepo(repo).
				WithName(tc.location.Name).
				WithPartnerInternalID(tc.location.PartnerInternalID).
				WithLocationType(tc.location.LocationType).
				WithIsArchived(false).
				WithPartnerInternalParentID(tc.location.PartnerInternalParentID).
				WithModificationTime(now, now)

			actual, err := builder.Build(ctx, db, "root", []*domain.Location{})
			if !tc.isValid {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.EqualValues(t, tc.location, actual)
			}
			mock.AssertExpectationsForObjects(
				t,
				db,
				repo,
				typeRepo,
			)
		})
	}
}
