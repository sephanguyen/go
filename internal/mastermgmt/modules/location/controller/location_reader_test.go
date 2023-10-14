package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLocationReaderService_RetrieveLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	db := new(mock_database.Ext)
	locationRepo := new(mock_repositories.MockLocationRepo)
	locationTypeRepo := new(mock_repositories.MockLocationTypeRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	service := &LocationReaderServices{
		DB:               db,
		LocationRepo:     locationRepo,
		LocationTypeRepo: locationTypeRepo,
		GetLocationQueryHandler: queries.GetLocationQueryHandler{
			UnleashClientIns: mockUnleashClient,
			DB:               db,
			LocationRepo:     locationRepo,
			LocationTypeRepo: locationTypeRepo,
		},
	}
	locationTypes := []*domain.LocationType{
		{LocationTypeID: "location-type1", ParentLocationTypeID: ""},
		{LocationTypeID: "location-type2", ParentLocationTypeID: "location-type1"},
	}

	t.Run("success with enable new structure", func(t *testing.T) {
		locationTypeRepo.On("RetrieveLocationTypes", mock.Anything, db).Return(locationTypes, nil).Once()
		locations := []*domain.Location{
			{LocationID: "location-1", LocationType: "location-type1", AccessPath: "location-1"},
			{LocationID: "location-2", LocationType: "location-type2", ParentLocationID: "location-1", AccessPath: "location-1/location-2"},
		}
		locationRepo.On("RetrieveLocations", mock.Anything, db, domain.FilterLocation{
			IncludeIsArchived: true,
		}).Return(locations, nil).Once()
		gotLocations, err := service.RetrieveLocations(ctx, &mpb.RetrieveLocationsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, gotLocations)
		assert.Len(t, gotLocations.Locations, len(locations))
		for _, l := range gotLocations.Locations {
			assert.NotEmpty(t, l.LocationId)
			assert.NotEmpty(t, l.LocationType)
			assert.Empty(t, l.Name)
			assert.NotEmpty(t, l.AccessPath)
		}
		locationRepo.AssertExpectations(t)
	})

	t.Run("error with enable new structure", func(t *testing.T) {
		locationTypeRepo.On("RetrieveLocationTypes", mock.Anything, db).Return(locationTypes, nil).Once()
		locationRepo.On("RetrieveLocations", mock.Anything, db, domain.FilterLocation{
			IncludeIsArchived: true,
		}).Return(nil, errors.New("Internal Error")).Once()
		gotLocations, err := service.RetrieveLocations(ctx, &mpb.RetrieveLocationsRequest{})
		assert.Error(t, err)
		assert.Nil(t, gotLocations)
		locationRepo.AssertExpectations(t)
	})

}

func TestLocationReaderService_RetrieveLocationTypes(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	db := new(mock_database.Ext)
	locationRepo := new(mock_repositories.MockLocationRepo)
	locationTypeRepo := new(mock_repositories.MockLocationTypeRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	service := &LocationReaderServices{
		DB:               db,
		LocationRepo:     locationRepo,
		LocationTypeRepo: locationTypeRepo,
		GetLocationQueryHandler: queries.GetLocationQueryHandler{
			UnleashClientIns: mockUnleashClient,
		},
	}
	t.Run("success", func(t *testing.T) {
		locationTypes := []*domain.LocationType{
			{LocationTypeID: "location-type1", ParentLocationTypeID: "location-type2"},
			{LocationTypeID: "location-type2", ParentLocationTypeID: "location-type1"},
		}
		locationTypeRepo.On("RetrieveLocationTypes", mock.Anything, db).Return(locationTypes, nil).Once()
		gotLocationTypes, err := service.RetrieveLocationTypes(ctx, &mpb.RetrieveLocationTypesRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, gotLocationTypes)
		assert.Len(t, gotLocationTypes.LocationTypes, len(locationTypes))
		for _, lt := range gotLocationTypes.LocationTypes {
			assert.NotEmpty(t, lt.LocationTypeId)
			assert.NotEmpty(t, lt.ParentLocationTypeId)
			assert.Empty(t, lt.DisplayName)
		}
		locationTypeRepo.AssertExpectations(t)
	})
	t.Run("error", func(t *testing.T) {
		locationTypeRepo.On("RetrieveLocationTypes", mock.Anything, db).Return(nil, errors.New("Internal Error")).Once()
		gotLocationTypes, err := service.RetrieveLocationTypes(ctx, &mpb.RetrieveLocationTypesRequest{})
		assert.Error(t, err)
		assert.Nil(t, gotLocationTypes)
		locationTypeRepo.AssertExpectations(t)
	})
}

func TestLocationReaderService_RetrieveLowestLevelLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	db := new(mock_database.Ext)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	locationRepo := new(mock_repositories.MockLocationRepo)
	service := &LocationReaderServices{
		DB:           db,
		LocationRepo: locationRepo,
		GetLocationQueryHandler: queries.GetLocationQueryHandler{
			UnleashClientIns: mockUnleashClient,
		},
	}
	locationIDs := []string{"location-1", "location-2"}
	testCases := []struct {
		name          string
		request       *mpb.RetrieveLowestLevelLocationsRequest
		setup         func(ctx context.Context, req *mpb.RetrieveLowestLevelLocationsRequest)
		expectedError error
	}{
		{
			name: "success",
			request: &mpb.RetrieveLowestLevelLocationsRequest{
				Name:        "Vinh",
				Limit:       5,
				LocationIds: locationIDs,
			},
			setup: func(ctx context.Context, req *mpb.RetrieveLowestLevelLocationsRequest) {
				params := &repo.GetLowestLevelLocationsParams{
					Name:        req.Name,
					Limit:       req.Limit,
					Offset:      req.Offset,
					LocationIDs: req.LocationIds,
				}
				locationRepo.On("GetLowestLevelLocationsV2", ctx, db, params).Return([]*domain.Location{
					{LocationID: "1", Name: "Vinh center 1"},
					{LocationID: "2", Name: "Vinh center 2"},
				}, nil).Once()
			},
		},
		{
			name: "fail",
			request: &mpb.RetrieveLowestLevelLocationsRequest{
				Name:        "Hanoi",
				Limit:       5,
				LocationIds: locationIDs,
			},
			setup: func(ctx context.Context, req *mpb.RetrieveLowestLevelLocationsRequest) {
				params := &repo.GetLowestLevelLocationsParams{
					Name:        req.Name,
					Limit:       req.Limit,
					Offset:      req.Offset,
					LocationIDs: req.LocationIds,
				}
				locationRepo.On("GetLowestLevelLocationsV2", ctx, db, params).Return(nil, errors.New("Internal Error x")).Once()
			},
			expectedError: errors.New("Internal Error x"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx, tc.request)
			gotLocations, err := service.RetrieveLowestLevelLocations(ctx, tc.request)
			if tc.expectedError != nil {
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gotLocations)
				for _, l := range gotLocations.Locations {
					assert.NotEmpty(t, l.LocationId)
					assert.NotEmpty(t, l.Name)
				}
				locationRepo.AssertExpectations(t)
			}
		})
	}
}

func TestLocationReaderService_ExportLocation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	locationRepo := new(mock_repositories.MockLocationRepo)

	locations := []*repo.Location{
		{
			LocationID:              database.Text("ID 1"),
			Name:                    database.Text("Location 1"),
			LocationType:            database.Text("Type 1"),
			PartnerInternalID:       database.Text("Partner 1"),
			PartnerInternalParentID: database.Text("Parent 1"),
		},
		{
			LocationID:              database.Text("ID 2"),
			Name:                    database.Text("Location 2"),
			LocationType:            database.Text("Type 2"),
			PartnerInternalParentID: database.Text("Parent 2"),
			PartnerInternalID:       database.Text("Partner 2"),
		},
		{
			LocationID:              database.Text("ID 3"),
			Name:                    database.Text("Location 3"),
			LocationType:            database.Text("Type 3"),
			PartnerInternalParentID: database.Text("Parent 3"),
			PartnerInternalID:       database.Text("Partner 3"),
		},
	}

	locationStr := `"location_id","partner_internal_id","name","location_type","partner_internal_parent_id"` + "\n" +
		`"ID 1","Partner 1","Location 1","Type 1","Parent 1"` + "\n" +
		`"ID 2","Partner 2","Location 2","Type 2","Parent 2"` + "\n" +
		`"ID 3","Partner 3","Location 3","Type 3","Parent 3"` + "\n"

	s := &LocationReaderServices{
		ExportLocationQueryHandler: queries.ExportLocationQueryHandler{
			DB:           db,
			LocationRepo: locationRepo,
		},
	}

	t.Run("export all data in db with correct column", func(t *testing.T) {
		// arrange
		locationRepo.On("GetAllLocations", ctx, db).Once().Return(locations, nil)

		byteData := []byte(locationStr)

		// act
		resp, err := s.ExportLocations(ctx, &mpb.ExportLocationsRequest{})

		// assert
		assert.Nil(t, err)
		assert.Equal(t, resp.Data, byteData)
	})

	t.Run("return internal error when retrieve data failed", func(t *testing.T) {
		// arrange
		locationRepo.On("GetAllLocations", ctx, db).Once().Return(nil, errors.New("sample error"))

		// act
		resp, err := s.ExportLocations(ctx, &mpb.ExportLocationsRequest{})

		// assert
		assert.Nil(t, resp.Data)
		assert.Equal(t, err, status.Error(codes.Internal, "sample error"))
	})
}

func TestLocationReaderService_ExportLocationTypes(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	db := new(mock_database.Ext)
	locTypeRepo := new(mock_repositories.MockLocationTypeRepo)

	locTypes := []*repo.LocationType{
		{
			LocationTypeID: database.Text("ID 1"),
			Name:           database.Text("Location type 1"),
			DisplayName:    database.Text("Display name 1"),
			ParentName:     database.Text("Parent 1"),
			Level:          database.Int4(0),
		},
		{
			LocationTypeID: database.Text("ID 2"),
			Name:           database.Text("Location type 2"),
			DisplayName:    database.Text("Display name 2"),
			ParentName:     database.Text("Parent 2"),
			Level:          database.Int4(1),
		},
		{
			LocationTypeID: database.Text("ID 3"),
			Name:           database.Text("Location type 3"),
			DisplayName:    database.Text("Display name 3"),
			ParentName:     database.Text("Parent 3"),
			Level:          database.Int4(2),
		},
	}

	s := &LocationReaderServices{
		ExportLocationQueryHandler: queries.ExportLocationQueryHandler{
			DB:               db,
			LocationTypeRepo: locTypeRepo,
		},
		GetLocationQueryHandler: queries.GetLocationQueryHandler{
			UnleashClientIns: mockUnleashClient,
		},
	}

	t.Run("return internal error when retrieve data failed", func(t *testing.T) {
		locTypeRepo.On("GetAllLocationTypes", ctx, db).Once().Return(nil, errors.New("sample error"))

		// act
		resp, err := s.ExportLocationTypes(ctx, &mpb.ExportLocationTypesRequest{})

		// assert
		assert.Nil(t, resp.Data)
		assert.Equal(t, err, status.Error(codes.Internal, "sample error"))
	})

	t.Run("export all data v2 in db with correct column", func(t *testing.T) {
		locTypeStr := `"location_type_id","name","display_name","level"` + "\n" +
			`"ID 1","Location type 1","Display name 1","0"` + "\n" +
			`"ID 2","Location type 2","Display name 2","1"` + "\n" +
			`"ID 3","Location type 3","Display name 3","2"` + "\n"
		locTypeRepo.On("GetAllLocationTypes", ctx, db).Once().Return(locTypes, nil)

		byteData := []byte(locTypeStr)

		// act
		resp, err := s.ExportLocationTypes(ctx, &mpb.ExportLocationTypesRequest{})

		// assert
		assert.Nil(t, err)

		assert.Equal(t, resp.Data, byteData)
		assert.Equal(t, string(resp.Data), string(byteData))
	})

}

func TestGenerateUnauthorizedLocation(t *testing.T) {
	t.Parallel()
	service := &LocationReaderServices{}

	tt := time.Date(2022, 9, 28, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2022, 9, 29, 0, 0, 0, 0, time.UTC)

	locationTypeInput := []*domain.LocationType{
		{LocationTypeID: "O", ParentLocationTypeID: ""},
		{LocationTypeID: "D", ParentLocationTypeID: "O"},
		{LocationTypeID: "B", ParentLocationTypeID: "D"},
		{LocationTypeID: "P", ParentLocationTypeID: "B"},
		{LocationTypeID: "C", ParentLocationTypeID: "P"},
	}

	locationTypeWrongInput := []*domain.LocationType{
		{LocationTypeID: "O", ParentLocationTypeID: ""},
		{LocationTypeID: "D", ParentLocationTypeID: "O"},
		{LocationTypeID: "B", ParentLocationTypeID: ""},
		{LocationTypeID: "P", ParentLocationTypeID: "B"},
		{LocationTypeID: "C", ParentLocationTypeID: "P"},
	}

	testCases := []struct {
		name              string
		inputLocationType []*domain.LocationType
		inputLocations    []*domain.Location
		expectedLocations map[string]domain.Location
		expectedError     error
	}{
		{
			name:              "wrong location type",
			inputLocationType: locationTypeWrongInput,
			inputLocations: []*domain.Location{
				{LocationID: "O_1", Name: "UnAuthorized", LocationType: "O", ParentLocationID: "", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1"},
			},
			expectedLocations: map[string]domain.Location{},
			expectedError:     errors.New("wrong location_type"),
		},
		{
			name:              "One child each location",
			inputLocationType: locationTypeInput,
			inputLocations: []*domain.Location{
				//	{LocationID: "D_1", Name: "N_D_1", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O"},
				//	{LocationID: "D_3", Name: "N_D_3", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_3"},
				//	{LocationID: "B_3", Name: "N_B_3", LocationType: "B", ParentLocationID: "D_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_3/B_3"},
				{LocationID: "P_3", Name: "N_P_3", LocationType: "P", ParentLocationID: "B_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_3"},
				{LocationID: "C_3", Name: "N_C_3", LocationType: "C", ParentLocationID: "P_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_3/C_3"},
				//	{LocationID: "D_4", Name: "N_D_4", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4"},
				//	{LocationID: "B_4", Name: "N_B_4", LocationType: "B", ParentLocationID: "D_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4/B_4"},
				//	{LocationID: "P_4", Name: "N_P_4", LocationType: "P", ParentLocationID: "B_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4/B_4/P_4"},
				{LocationID: "C_4", Name: "N_C_4", LocationType: "C", ParentLocationID: "P_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_4/B_4/P_4/C_4"},
			},
			expectedLocations: map[string]domain.Location{
				"O_1": {LocationID: "O_1", Name: "UnAuthorized", LocationType: "O", ParentLocationID: "", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1"},
				"D_3": {LocationID: "D_3", Name: "UnAuthorized", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_3"},
				"B_3": {LocationID: "B_3", Name: "UnAuthorized", LocationType: "B", ParentLocationID: "D_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_3/B_3"},
				"P_3": {LocationID: "P_3", Name: "N_P_3", LocationType: "P", ParentLocationID: "B_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_3"},
				"C_3": {LocationID: "C_3", Name: "N_C_3", LocationType: "C", ParentLocationID: "P_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_3/C_3"},
				"D_4": {LocationID: "D_4", Name: "UnAuthorized", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4"},
				"B_4": {LocationID: "B_4", Name: "UnAuthorized", LocationType: "B", ParentLocationID: "D_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4/B_4"},
				"P_4": {LocationID: "P_4", Name: "UnAuthorized", LocationType: "P", ParentLocationID: "B_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4/B_4/P_4"},
				"C_4": {LocationID: "C_4", Name: "N_C_4", LocationType: "C", ParentLocationID: "P_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_4/B_4/P_4/C_4"},
			},
			expectedError: nil,
		},
		{
			name:              "One child each location",
			inputLocationType: locationTypeInput,
			inputLocations: []*domain.Location{
				//	{LocationID: "D_1", Name: "N_D_1", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O"},
				{LocationID: "D_1", Name: "N_D_1", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_1"},
				{LocationID: "B_1", Name: "N_B_1", LocationType: "B", ParentLocationID: "D_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_1/B_1"},
				{LocationID: "P_1", Name: "N_P_1", LocationType: "P", ParentLocationID: "B_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_1/B_1/P_1"},
				{LocationID: "C_1", Name: "N_C_1", LocationType: "C", ParentLocationID: "P_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_1/B_1/P_1/C_1"},
				//	{LocationID: "D_2", Name: "N_D_2", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_2"},
				{LocationID: "B_2", Name: "N_B_2", LocationType: "B", ParentLocationID: "D_2", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_2/B_2"},
				{LocationID: "P_2", Name: "N_P_2", LocationType: "P", ParentLocationID: "B_2", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_2/B_2/P_2"},
				{LocationID: "C_2", Name: "N_C_2", LocationType: "C", ParentLocationID: "P_2", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_2/B_2/P_2/C_2"},
				//	{LocationID: "D_3", Name: "N_D_3", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_3"},
				//	{LocationID: "B_3", Name: "N_B_3", LocationType: "B", ParentLocationID: "D_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_3/B_3"},
				{LocationID: "P_3", Name: "N_P_3", LocationType: "P", ParentLocationID: "B_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_3"},
				{LocationID: "C_3", Name: "N_C_3", LocationType: "C", ParentLocationID: "P_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_3/C_3"},
				//	{LocationID: "D_4", Name: "N_D_4", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4"},
				//	{LocationID: "B_4", Name: "N_B_4", LocationType: "B", ParentLocationID: "D_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4/B_4"},
				//	{LocationID: "P_4", Name: "N_P_4", LocationType: "P", ParentLocationID: "B_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4/B_4/P_4"},
				{LocationID: "C_4", Name: "N_C_4", LocationType: "C", ParentLocationID: "P_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_4/B_4/P_4/C_4"},
			},
			expectedLocations: map[string]domain.Location{
				"O_1": {LocationID: "O_1", Name: "UnAuthorized", LocationType: "O", ParentLocationID: "", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1"},
				"D_1": {LocationID: "D_1", Name: "N_D_1", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_1"},
				"B_1": {LocationID: "B_1", Name: "N_B_1", LocationType: "B", ParentLocationID: "D_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_1/B_1"},
				"P_1": {LocationID: "P_1", Name: "N_P_1", LocationType: "P", ParentLocationID: "B_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_1/B_1/P_1"},
				"C_1": {LocationID: "C_1", Name: "N_C_1", LocationType: "C", ParentLocationID: "P_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_1/B_1/P_1/C_1"},
				"D_2": {LocationID: "D_2", Name: "UnAuthorized", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_2"},
				"B_2": {LocationID: "B_2", Name: "N_B_2", LocationType: "B", ParentLocationID: "D_2", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_2/B_2"},
				"P_2": {LocationID: "P_2", Name: "N_P_2", LocationType: "P", ParentLocationID: "B_2", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_2/B_2/P_2"},
				"C_2": {LocationID: "C_2", Name: "N_C_2", LocationType: "C", ParentLocationID: "P_2", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_2/B_2/P_2/C_2"},
				"D_3": {LocationID: "D_3", Name: "UnAuthorized", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_3"},
				"B_3": {LocationID: "B_3", Name: "UnAuthorized", LocationType: "B", ParentLocationID: "D_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_3/B_3"},
				"P_3": {LocationID: "P_3", Name: "N_P_3", LocationType: "P", ParentLocationID: "B_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_3"},
				"C_3": {LocationID: "C_3", Name: "N_C_3", LocationType: "C", ParentLocationID: "P_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_3/C_3"},
				"D_4": {LocationID: "D_4", Name: "UnAuthorized", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4"},
				"B_4": {LocationID: "B_4", Name: "UnAuthorized", LocationType: "B", ParentLocationID: "D_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4/B_4"},
				"P_4": {LocationID: "P_4", Name: "UnAuthorized", LocationType: "P", ParentLocationID: "B_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4/B_4/P_4"},
				"C_4": {LocationID: "C_4", Name: "N_C_4", LocationType: "C", ParentLocationID: "P_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_4/B_4/P_4/C_4"},
			},
			expectedError: nil,
		},
		{
			name:              "One child each location",
			inputLocationType: locationTypeInput,
			inputLocations: []*domain.Location{
				// {LocationID: "O_1", Name: "N_O_1", LocationType: "O", ParentLocationID: "", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1"},
				{LocationID: "D_10", Name: "N_D_10", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_10"},
				{LocationID: "B_10", Name: "N_B_10", LocationType: "B", ParentLocationID: "D_10", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_10/B_10"},
				{LocationID: "P_10", Name: "N_P_10", LocationType: "P", ParentLocationID: "B_10", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_10/B_10/P_10"},
				{LocationID: "C_10", Name: "N_C_10", LocationType: "C", ParentLocationID: "P_10", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_10/B_10/P_10/C_10"},
				{LocationID: "D_11", Name: "N_D_11", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_11"},
				{LocationID: "B_11", Name: "N_B_11", LocationType: "B", ParentLocationID: "D_11", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_11/B_11"},
				{LocationID: "P_11", Name: "N_P_11", LocationType: "P", ParentLocationID: "B_11", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_11/B_11/P_11"},
				{LocationID: "C_11", Name: "N_C_11", LocationType: "C", ParentLocationID: "P_11", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_11/B_11/P_11/C_11"},
				// {LocationID: "D_20", Name: "N_D_20", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20"},
				{LocationID: "B_20", Name: "N_B_20", LocationType: "B", ParentLocationID: "D_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_20"},
				{LocationID: "P_20", Name: "N_P_20", LocationType: "P", ParentLocationID: "B_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_20/P_20"},
				{LocationID: "C_20", Name: "N_C_20", LocationType: "C", ParentLocationID: "P_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_20/P_20/C_20"},
				{LocationID: "B_21", Name: "N_B_21", LocationType: "B", ParentLocationID: "D_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_21"},
				{LocationID: "P_21", Name: "N_P_21", LocationType: "P", ParentLocationID: "B_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_21/P_21"},
				{LocationID: "C_21", Name: "N_C_21", LocationType: "C", ParentLocationID: "P_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_21/P_21/C_21"},
				// {LocationID: "B_22", Name: "N_B_22", LocationType: "B", ParentLocationID: "D_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_22"},
				{LocationID: "P_22", Name: "N_P_22", LocationType: "P", ParentLocationID: "B_22", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_22/P_22"},
				{LocationID: "C_22", Name: "N_C_22", LocationType: "C", ParentLocationID: "P_22", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_22/P_22/C_22"},
				// {LocationID: "D_3", Name: "N_D_3", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3"},
				// {LocationID: "B_3", Name: "N_B_3", LocationType: "B", ParentLocationID: "D_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3"},
				{LocationID: "P_3", Name: "N_P_3", LocationType: "P", ParentLocationID: "B_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_3"},
				{LocationID: "C_3", Name: "N_C_3", LocationType: "C", ParentLocationID: "P_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_3/C_3"},
				{LocationID: "P_6", Name: "N_P_6", LocationType: "P", ParentLocationID: "B_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_6"},
				{LocationID: "C_6", Name: "N_C_6", LocationType: "C", ParentLocationID: "P_6", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_6/C_6"},
				// {LocationID: "D_4", Name: "N_D_4", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_4"},
				// {LocationID: "B_4", Name: "N_B_4", LocationType: "B", ParentLocationID: "D_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_4/B_4"},
				// {LocationID: "P_4", Name: "N_P_4", LocationType: "P", ParentLocationID: "B_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_4/B_4/P_4"},
				{LocationID: "C_4", Name: "N_C_4", LocationType: "C", ParentLocationID: "P_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_4/B_4/P_4/C_4"},
				{LocationID: "C_5", Name: "N_C_5", LocationType: "C", ParentLocationID: "P_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_4/B_4/P_4/C_5"},
			},
			expectedLocations: map[string]domain.Location{
				"O_1":  {LocationID: "O_1", Name: "UnAuthorized", LocationType: "O", ParentLocationID: "", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1"},
				"D_10": {LocationID: "D_10", Name: "N_D_10", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_10"},
				"B_10": {LocationID: "B_10", Name: "N_B_10", LocationType: "B", ParentLocationID: "D_10", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_10/B_10"},
				"P_10": {LocationID: "P_10", Name: "N_P_10", LocationType: "P", ParentLocationID: "B_10", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_10/B_10/P_10"},
				"C_10": {LocationID: "C_10", Name: "N_C_10", LocationType: "C", ParentLocationID: "P_10", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_10/B_10/P_10/C_10"},
				"D_11": {LocationID: "D_11", Name: "N_D_11", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_11"},
				"B_11": {LocationID: "B_11", Name: "N_B_11", LocationType: "B", ParentLocationID: "D_11", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_11/B_11"},
				"P_11": {LocationID: "P_11", Name: "N_P_11", LocationType: "P", ParentLocationID: "B_11", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_11/B_11/P_11"},
				"C_11": {LocationID: "C_11", Name: "N_C_11", LocationType: "C", ParentLocationID: "P_11", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_11/B_11/P_11/C_11"},
				"D_20": {LocationID: "D_20", Name: "UnAuthorized", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_20"},
				"B_20": {LocationID: "B_20", Name: "N_B_20", LocationType: "B", ParentLocationID: "D_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_20"},
				"P_20": {LocationID: "P_20", Name: "N_P_20", LocationType: "P", ParentLocationID: "B_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_20/P_20"},
				"C_20": {LocationID: "C_20", Name: "N_C_20", LocationType: "C", ParentLocationID: "P_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_20/P_20/C_20"},
				"B_21": {LocationID: "B_21", Name: "N_B_21", LocationType: "B", ParentLocationID: "D_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_21"},
				"P_21": {LocationID: "P_21", Name: "N_P_21", LocationType: "P", ParentLocationID: "B_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_21/P_21"},
				"C_21": {LocationID: "C_21", Name: "N_C_21", LocationType: "C", ParentLocationID: "P_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_21/P_21/C_21"},
				"B_22": {LocationID: "B_22", Name: "UnAuthorized", LocationType: "B", ParentLocationID: "D_20", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_20/B_22"},
				"P_22": {LocationID: "P_22", Name: "N_P_22", LocationType: "P", ParentLocationID: "B_22", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_22/P_22"},
				"C_22": {LocationID: "C_22", Name: "N_C_22", LocationType: "C", ParentLocationID: "P_22", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_20/B_22/P_22/C_22"},
				"D_3":  {LocationID: "D_3", Name: "UnAuthorized", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_3"},
				"B_3":  {LocationID: "B_3", Name: "UnAuthorized", LocationType: "B", ParentLocationID: "D_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_3/B_3"},
				"P_3":  {LocationID: "P_3", Name: "N_P_3", LocationType: "P", ParentLocationID: "B_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_3"},
				"C_3":  {LocationID: "C_3", Name: "N_C_3", LocationType: "C", ParentLocationID: "P_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_3/C_3"},
				"P_6":  {LocationID: "P_6", Name: "N_P_6", LocationType: "P", ParentLocationID: "B_3", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_6"},
				"C_6":  {LocationID: "C_6", Name: "N_C_6", LocationType: "C", ParentLocationID: "P_6", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_3/B_3/P_6/C_6"},
				"D_4":  {LocationID: "D_4", Name: "UnAuthorized", LocationType: "D", ParentLocationID: "O_1", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4"},
				"B_4":  {LocationID: "B_4", Name: "UnAuthorized", LocationType: "B", ParentLocationID: "D_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4/B_4"},
				"P_4":  {LocationID: "P_4", Name: "UnAuthorized", LocationType: "P", ParentLocationID: "B_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: true, AccessPath: "O_1/D_4/B_4/P_4"},
				"C_4":  {LocationID: "C_4", Name: "N_C_4", LocationType: "C", ParentLocationID: "P_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_4/B_4/P_4/C_4"},
				"C_5":  {LocationID: "C_5", Name: "N_C_5", LocationType: "C", ParentLocationID: "P_4", CreatedAt: tt, UpdatedAt: updated, IsUnauthorized: false, AccessPath: "O_1/D_4/B_4/P_4/C_5"},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run("success", func(t *testing.T) {
			rs, err := service.GenerateUnauthorizedLocation(tc.inputLocations, tc.inputLocationType)
			assert.Equal(t, len(tc.expectedLocations), len(rs))
			assert.Equal(t, tc.expectedError, err)
			for _, l := range rs {
				l.CreatedAt = tt
				l.UpdatedAt = updated
				assert.Equal(t, tc.expectedLocations[l.LocationID], *l)
			}
		})
	}

}

func TestSortLocationType(t *testing.T) {
	t.Parallel()
	service := &LocationReaderServices{}

	locationTypeInput := []*domain.LocationType{
		{LocationTypeID: "D", ParentLocationTypeID: "O"},
		{LocationTypeID: "O", ParentLocationTypeID: ""},
		{LocationTypeID: "C", ParentLocationTypeID: "P"},
		{LocationTypeID: "P", ParentLocationTypeID: "B"},
		{LocationTypeID: "B", ParentLocationTypeID: "D"},
	}

	locationTypeOutputExpect := []*domain.LocationType{
		{LocationTypeID: "O", ParentLocationTypeID: ""},
		{LocationTypeID: "D", ParentLocationTypeID: "O"},
		{LocationTypeID: "B", ParentLocationTypeID: "D"},
		{LocationTypeID: "P", ParentLocationTypeID: "B"},
		{LocationTypeID: "C", ParentLocationTypeID: "P"},
	}

	rs, _ := service.SortLocationType(locationTypeInput)

	t.Run("success", func(t *testing.T) {
		assert.Equal(t, locationTypeOutputExpect, rs)
	})

	locationTypeWrongInput := []*domain.LocationType{
		{LocationTypeID: "D", ParentLocationTypeID: "O"},
		{LocationTypeID: "O", ParentLocationTypeID: ""},
		{LocationTypeID: "C", ParentLocationTypeID: ""},
		{LocationTypeID: "P", ParentLocationTypeID: "B"},
		{LocationTypeID: "B", ParentLocationTypeID: "D"},
	}
	rs, err := service.SortLocationType(locationTypeWrongInput)

	t.Run("success", func(t *testing.T) {
		assert.Equal(t, err.Error(), "wrong location_type")
	})

	locationTypeWrongInput_2 := []*domain.LocationType{
		{LocationTypeID: "D", ParentLocationTypeID: "O"},
		{LocationTypeID: "O", ParentLocationTypeID: ""},
		{LocationTypeID: "C", ParentLocationTypeID: "B"},
		{LocationTypeID: "P", ParentLocationTypeID: "B"},
		{LocationTypeID: "B", ParentLocationTypeID: "D"},
	}
	rs, err = service.SortLocationType(locationTypeWrongInput_2)
	t.Run("success", func(t *testing.T) {
		assert.Equal(t, err.Error(), "wrong location_type")
	})

	locationTypeWrongInput_3 := []*domain.LocationType{
		{LocationTypeID: "D", ParentLocationTypeID: "O"},
		{LocationTypeID: "O", ParentLocationTypeID: ""},
		{LocationTypeID: "C", ParentLocationTypeID: "D"},
		{LocationTypeID: "C", ParentLocationTypeID: "B"},
		{LocationTypeID: "B", ParentLocationTypeID: "D"},
	}
	rs, err = service.SortLocationType(locationTypeWrongInput_3)

	t.Run("success", func(t *testing.T) {
		assert.Equal(t, err.Error(), "wrong location_type")
	})
}

func TestGetLocationTree(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	locationRepo := new(mock_repositories.MockLocationRepo)
	locationTypeRepo := new(mock_repositories.MockLocationTypeRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	service := &LocationReaderServices{
		DB:               db,
		LocationRepo:     locationRepo,
		LocationTypeRepo: locationTypeRepo,
		GetLocationQueryHandler: queries.GetLocationQueryHandler{
			UnleashClientIns: mockUnleashClient,
			DB:               db,
			LocationRepo:     locationRepo,
			LocationTypeRepo: locationTypeRepo,
		},
	}

	today := time.Now()

	locationTypes := []*domain.LocationType{
		{LocationTypeID: "TO", Level: 0, Name: "Org"},
		{LocationTypeID: "T1", Level: 1, Name: "Brand"},
		{LocationTypeID: "T2", Level: 2, Name: "Area"},
		{LocationTypeID: "T3", Level: 3, Name: "Center"},
		{LocationTypeID: "T4", Level: 4, Name: "Place"},
	}

	t.Run("success", func(t *testing.T) {
		// arrange
		locations := []*domain.Location{
			{LocationID: "P1", Name: "Place One 1", LocationType: "T4", ParentLocationID: "C1", AccessPath: "O/B1/A1/C1/P1", CreatedAt: today, UpdatedAt: today, IsUnauthorized: false},
			{LocationID: "P2", Name: "Place Two 2", LocationType: "T4", ParentLocationID: "C2", AccessPath: "O/B1/A2/C2/P2", CreatedAt: today, UpdatedAt: today, IsUnauthorized: false},
		}

		locationRepo.On("RetrieveLocations", mock.Anything, db, domain.FilterLocation{IncludeIsArchived: true}).Return(locations, nil).Once()
		locationTypeRepo.On("GetLocationTypeByIDs", mock.Anything, db, database.TextArray([]string{"T4"}), false).Return(locationTypes, nil).Once()
		expectedTree := &domain.TreeLocation{
			LocationID:       "O",
			Name:             "UnAuthorized",
			ParentLocationID: "",
			LocationType:     "",
			IsArchived:       false,
			AccessPath:       "O",
			IsUnauthorized:   true,
			IsLowestLevel:    false,
			Children: []*domain.TreeLocation{
				{
					LocationID:       "B1",
					Name:             "UnAuthorized",
					ParentLocationID: "O",
					LocationType:     "",
					IsArchived:       false,
					AccessPath:       "O/B1",
					IsUnauthorized:   true,
					IsLowestLevel:    false,
					Children: []*domain.TreeLocation{
						{
							LocationID:       "A1",
							Name:             "UnAuthorized",
							ParentLocationID: "B1",
							LocationType:     "",
							IsArchived:       false,
							AccessPath:       "O/B1/A1",
							IsUnauthorized:   true,
							IsLowestLevel:    false,
							Children: []*domain.TreeLocation{
								{
									LocationID:       "C1",
									Name:             "UnAuthorized",
									ParentLocationID: "A1",
									LocationType:     "",
									IsArchived:       false,
									AccessPath:       "O/B1/A1/C1",
									IsUnauthorized:   true,
									IsLowestLevel:    false,
									Children: []*domain.TreeLocation{
										{
											LocationID:       "P1",
											Name:             "Place One 1",
											ParentLocationID: "C1",
											LocationType:     "T4",
											IsArchived:       false,
											AccessPath:       "O/B1/A1/C1/P1",
											IsUnauthorized:   false,
											IsLowestLevel:    true,
											CreatedAt:        today,
											UpdatedAt:        today,
										},
									},
								},
							},
						},
						{
							LocationID:       "A2",
							Name:             "UnAuthorized",
							ParentLocationID: "B1",
							LocationType:     "",
							IsArchived:       false,
							AccessPath:       "O/B1/A2",
							IsUnauthorized:   true,
							IsLowestLevel:    false,
							Children: []*domain.TreeLocation{
								{
									LocationID:       "C2",
									Name:             "UnAuthorized",
									ParentLocationID: "A2",
									LocationType:     "",
									IsArchived:       false,
									AccessPath:       "O/B1/A2/C2",
									IsUnauthorized:   true,
									IsLowestLevel:    false,
									Children: []*domain.TreeLocation{
										{
											LocationID:       "P2",
											Name:             "Place Two 2",
											ParentLocationID: "C2",
											LocationType:     "T4",
											IsArchived:       false,
											AccessPath:       "O/B1/A2/C2/P2",
											IsUnauthorized:   false,
											IsLowestLevel:    true,
											CreatedAt:        today,
											UpdatedAt:        today,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		// act
		res, err := service.GetLocationTree(ctx, &mpb.GetLocationTreeRequest{})
		locTreeJSON := res.GetTree()
		var actualTree domain.TreeLocation
		err = json.Unmarshal([]byte(locTreeJSON), &actualTree)

		// assert
		assert.Nil(t, err)
		assert.True(t, compareTreeLocations(&actualTree, expectedTree))
		locationRepo.AssertExpectations(t)
		locationTypeRepo.AssertExpectations(t)
	})
	t.Run("no access to any location", func(t *testing.T) {
		// arrange
		locations := []*domain.Location{}

		locationRepo.On("RetrieveLocations", mock.Anything, db, domain.FilterLocation{IncludeIsArchived: true}).Return(locations, nil).Once()
		expectedErr := fmt.Errorf(`GetLocationsTree: User ID [%s] does not have access to any location`, "")
		expectedErrWrapper := status.Error(codes.Internal, fmt.Sprintf("GetLocationQueryHandler.GetLocationsTree: %v", expectedErr))
		// act
		res, err := service.GetLocationTree(ctx, &mpb.GetLocationTreeRequest{})

		// assert
		assert.Equal(t, expectedErrWrapper, err)
		assert.Empty(t, res)
		locationRepo.AssertExpectations(t)
	})
}

func compareTreeLocations(l1, l2 *domain.TreeLocation) bool {
	if l1.LocationID != l2.LocationID || l1.Name != l2.Name ||
		l1.LocationType != l2.LocationType || l1.ParentLocationID != l2.ParentLocationID ||
		l1.IsArchived != l2.IsArchived || l1.AccessPath != l2.AccessPath ||
		l1.IsUnauthorized != l2.IsUnauthorized || l1.IsLowestLevel != l2.IsLowestLevel ||
		!l1.UpdatedAt.Equal(l2.UpdatedAt) || !l1.CreatedAt.Equal(l2.CreatedAt) {
		return false
	}

	if len(l1.Children) != len(l2.Children) {
		return false
	}

	l2Children := make(map[string]*domain.TreeLocation)
	for _, child := range l2.Children {
		l2Children[child.LocationID] = child
	}

	// Recursively compare the children of both trees
	for _, child1 := range l1.Children {
		child2, ok := l2Children[child1.LocationID]
		if !ok || !compareTreeLocations(child1, child2) {
			return false
		}
		delete(l2Children, child1.LocationID)
	}

	// Check if there are any remaining children in the second tree
	if len(l2Children) != 0 {
		return false
	}

	return true
}
