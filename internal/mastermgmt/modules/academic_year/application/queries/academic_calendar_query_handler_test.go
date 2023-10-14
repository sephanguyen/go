package queries

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	domain "github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/infrastructure/repo"
	configuration_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/domain"
	location_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_academic_year_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/academic_year/infrastructure/repo"
	mock_configuration_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/configuration/infrastructure/repo"
	mock_location_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	name         string
	expectedResp interface{}
	setup        func(ctx context.Context)
	expectedErr  error
}

func TestAcademicCalendarQueryHandler_GetLocationsForAcademic(t *testing.T) {
	const locationTypeLevelKey = "mastermgmt.academic_calendar.location_type_level"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()

	masterDB := &mock_database.Ext{}
	bobDB := &mock_database.Ext{}

	academicWeekRepo := new(mock_academic_year_repo.MockAcademicWeekRepo)
	locationRepo := new(mock_location_repo.MockLocationRepo)
	locationTypeRepo := new(mock_location_repo.MockLocationTypeRepo)
	configRepo := new(mock_configuration_repo.MockConfigRepo)

	tempAcademicYearId := "temp-academic-year-id"

	levelConfigResp := &configuration_domain.InternalConfiguration{
		ID:              "location-type-level-config-key",
		ConfigKey:       locationTypeLevelKey,
		ConfigValue:     "2",
		ConfigValueType: "number",
	}

	locationTypesResp := []*location_domain.LocationType{
		{
			LocationTypeID: "location-type-id-01",
			Name:           "",
			DisplayName:    "location type display",
			Level:          2,
			CreatedAt:      now,
			UpdatedAt:      now,
			Persisted:      true,
		},
	}

	locationsResp := []*location_domain.Location{
		{
			LocationID: "location-id-01",
			Name:       "Center A1",
		},
	}

	locTypeIDsResp := sliceutils.Map(locationTypesResp, func(l *location_domain.LocationType) string {
		return l.LocationTypeID
	})

	tcs := []TestCase{
		{
			name:         "get location ids by location type config level success",
			expectedResp: locationsResp,
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, masterDB, locationTypeLevelKey).Once().Return(levelConfigResp, nil)
				locationTypeRepo.On("GetLocationTypesByLevel", ctx, bobDB, "2").Once().Return(locationTypesResp, nil)
				locationRepo.On("GetLocationByLocationTypeIDs", ctx, bobDB, locTypeIDsResp).Once().Return(locationsResp, nil)
				academicWeekRepo.On("GetLocationsByAcademicWeekID", ctx, masterDB, tempAcademicYearId).Once().Return([]string{}, nil)
			},
			expectedErr: nil,
		},
		{
			name:         "get location ids by location type config level success with location have been imported before",
			expectedResp: []*location_domain.Location{},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, masterDB, locationTypeLevelKey).Once().Return(levelConfigResp, nil)
				locationTypeRepo.On("GetLocationTypesByLevel", ctx, bobDB, "2").Once().Return(locationTypesResp, nil)
				locationRepo.On("GetLocationByLocationTypeIDs", ctx, bobDB, locTypeIDsResp).Once().Return([]*location_domain.Location{
					{
						LocationID: "location-id-01",
						Name:       "Center A1",
					},
				}, nil)
				academicWeekRepo.On("GetLocationsByAcademicWeekID", ctx, masterDB, tempAcademicYearId).Once().Return([]string{"location-id-01"}, nil)
			},
			expectedErr: nil,
		},
		{
			name:         "call GetLocationTypesByLevel fail",
			expectedResp: []*location_domain.Location{},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, masterDB, locationTypeLevelKey).Once().Return(levelConfigResp, nil)
				locationTypeRepo.On("GetLocationTypesByLevel", ctx, bobDB, "2").Once().Return(nil, fmt.Errorf("query err"))
			},
			expectedErr: fmt.Errorf("GetLocationsByLocationTypeLevelConfig: %w", fmt.Errorf("GetLocationTypesByLevel: %w", fmt.Errorf("query err"))),
		},
		{
			name:         "call GetLocationByLocationTypeIDs fail",
			expectedResp: []*location_domain.Location{},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, masterDB, locationTypeLevelKey).Once().Return(levelConfigResp, nil)
				locationTypeRepo.On("GetLocationTypesByLevel", ctx, bobDB, "2").Once().Return(locationTypesResp, nil)
				locationRepo.On("GetLocationByLocationTypeIDs", ctx, bobDB, locTypeIDsResp).Once().Return(nil, fmt.Errorf("query err"))
			},
			expectedErr: fmt.Errorf("GetLocationsByLocationTypeLevelConfig: %w", fmt.Errorf("GetLocationByLocationTypeIDs: %w", fmt.Errorf("query err"))),
		},
		{
			name:         "call GetLocationsByAcademicWeekID fail",
			expectedResp: []*location_domain.Location{},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, masterDB, locationTypeLevelKey).Once().Return(levelConfigResp, nil)
				locationTypeRepo.On("GetLocationTypesByLevel", ctx, bobDB, "2").Once().Return(locationTypesResp, nil)
				locationRepo.On("GetLocationByLocationTypeIDs", ctx, bobDB, locTypeIDsResp).Once().Return(locationsResp, nil)
				academicWeekRepo.On("GetLocationsByAcademicWeekID", ctx, masterDB, tempAcademicYearId).Once().Return(nil, fmt.Errorf("query err"))
			},
			expectedErr: fmt.Errorf("GetLocationsByAcademicWeekID: %w", fmt.Errorf("query err")),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			academicCalendarQueryHandler := &AcademicCalendarQueryHandler{
				MasterDB:         masterDB,
				BobDB:            bobDB,
				AcademicWeekRepo: academicWeekRepo,
				LocationRepo:     locationRepo,
				LocationTypeRepo: locationTypeRepo,
				ConfigRepo:       configRepo,
			}
			resp, err := academicCalendarQueryHandler.GetLocationsForAcademic(ctx, tempAcademicYearId)

			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
				assert.Empty(t, resp)
			} else {
				assert.Equal(t, tc.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, masterDB, academicWeekRepo, locationRepo, locationTypeRepo, configRepo)
		})
	}
}

func TestAcademicCalendarQueryHandler_GetLocationsByLocationTypeLevelConfig(t *testing.T) {
	const locationTypeLevelKey = "mastermgmt.academic_calendar.location_type_level"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()

	masterDB := &mock_database.Ext{}
	bobDB := &mock_database.Ext{}

	academicWeekRepo := new(mock_academic_year_repo.MockAcademicWeekRepo)
	locationRepo := new(mock_location_repo.MockLocationRepo)
	locationTypeRepo := new(mock_location_repo.MockLocationTypeRepo)
	configRepo := new(mock_configuration_repo.MockConfigRepo)

	levelConfigResp := &configuration_domain.InternalConfiguration{
		ID:              "location-type-level-config-key",
		ConfigKey:       locationTypeLevelKey,
		ConfigValue:     "2",
		ConfigValueType: "number",
	}

	locationTypesResp := []*location_domain.LocationType{
		{
			LocationTypeID: "location-type-id-01",
			Name:           "",
			DisplayName:    "location type display",
			Level:          2,
			CreatedAt:      now,
			UpdatedAt:      now,
			Persisted:      true,
		},
	}

	locationsResp := []*location_domain.Location{
		{
			LocationID: "location-id-01",
			Name:       "Center A1",
		},
	}

	locTypeIDsResp := sliceutils.Map(locationTypesResp, func(l *location_domain.LocationType) string {
		return l.LocationTypeID
	})

	tcs := []TestCase{
		{
			name:         "get location ids by location type config level success",
			expectedResp: locationsResp,
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, masterDB, locationTypeLevelKey).Once().Return(levelConfigResp, nil)
				locationTypeRepo.On("GetLocationTypesByLevel", ctx, bobDB, "2").Once().Return(locationTypesResp, nil)
				locationRepo.On("GetLocationByLocationTypeIDs", ctx, bobDB, locTypeIDsResp).Once().Return(locationsResp, nil)
			},
			expectedErr: nil,
		},
		{
			name:         "call GetLocationTypesByLevel fail",
			expectedResp: []*location_domain.Location{},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, masterDB, locationTypeLevelKey).Once().Return(levelConfigResp, nil)
				locationTypeRepo.On("GetLocationTypesByLevel", ctx, bobDB, "2").Once().Return(nil, fmt.Errorf("query err"))
			},
			expectedErr: fmt.Errorf("GetLocationTypesByLevel: %w", fmt.Errorf("query err")),
		},
		{
			name:         "call GetLocationByLocationTypeIDs fail",
			expectedResp: []*location_domain.Location{},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, masterDB, locationTypeLevelKey).Once().Return(levelConfigResp, nil)
				locationTypeRepo.On("GetLocationTypesByLevel", ctx, bobDB, "2").Once().Return(locationTypesResp, nil)
				locationRepo.On("GetLocationByLocationTypeIDs", ctx, bobDB, locTypeIDsResp).Once().Return(nil, fmt.Errorf("query err"))
			},
			expectedErr: fmt.Errorf("GetLocationByLocationTypeIDs: %w", fmt.Errorf("query err")),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			academicCalendarQueryHandler := &AcademicCalendarQueryHandler{
				MasterDB:         masterDB,
				BobDB:            bobDB,
				AcademicWeekRepo: academicWeekRepo,
				LocationRepo:     locationRepo,
				LocationTypeRepo: locationTypeRepo,
				ConfigRepo:       configRepo,
			}
			resp, err := academicCalendarQueryHandler.GetLocationsByLocationTypeLevelConfig(ctx)

			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
				assert.Empty(t, resp)
			} else {
				assert.Equal(t, tc.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, masterDB, academicWeekRepo, locationRepo, locationTypeRepo, configRepo)
		})
	}
}

func TestAcademicCalendarQueryHandler_ExportAcademicCalendar(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	masterDB := &mock_database.Ext{}
	bobDB := &mock_database.Ext{}

	academicYearRepo := new(mock_academic_year_repo.MockAcademicYearRepo)
	academicWeekRepo := new(mock_academic_year_repo.MockAcademicWeekRepo)
	locationRepo := new(mock_location_repo.MockLocationRepo)
	locationTypeRepo := new(mock_location_repo.MockLocationTypeRepo)
	configRepo := new(mock_configuration_repo.MockConfigRepo)
	academicClosedDayRepo := new(mock_academic_year_repo.MockAcademicClosedDayRepo)

	academicYearID := "academic_year_id"
	reqLocationID := "location-1"

	academicYear := &domain.AcademicYear{
		AcademicYearID: academicYearID,
		Name:           "2023",
		StartDate:      time.Date(2023, 04, 01, 00, 0, 0, 0, time.Local),
		EndDate:        time.Date(2024, 03, 31, 00, 0, 0, 0, time.Local),
	}

	locations := []*location_domain.Location{
		{
			LocationID: reqLocationID,
			Name:       "Location 1",
		},
		{
			LocationID: "location-2",
			Name:       "Location 2",
		},
		{
			LocationID: "location-3",
			Name:       "Location 3",
		},
	}

	weeks := []*repo.AcademicWeek{
		{
			AcademicWeekID: database.Text("academic_week_01"),
			WeekOrder:      database.Int2(1),
			Name:           database.Text("Week 1"),
			StartDate:      pgtype.Date{Time: time.Date(2023, 04, 01, 00, 0, 0, 0, time.Local)},
			EndDate:        pgtype.Date{Time: time.Date(2023, 04, 8, 0, 0, 0, 0, time.Local)},
			Period:         database.Text("Term 1"),
			AcademicYearID: database.Text(academicYearID),
			LocationID:     database.Text(locations[0].LocationID),
		},
		{
			AcademicWeekID: database.Text("academic_week_02"),
			WeekOrder:      database.Int2(1),
			Name:           database.Text("Week 1"),
			StartDate:      pgtype.Date{Time: time.Date(2023, 04, 01, 00, 0, 0, 0, time.Local)},
			EndDate:        pgtype.Date{Time: time.Date(2023, 04, 8, 0, 0, 0, 0, time.Local)},
			Period:         database.Text("Term 1"),
			AcademicYearID: database.Text(academicYearID),
			LocationID:     database.Text(locations[1].LocationID),
		},
		{
			AcademicWeekID: database.Text("academic_week_03"),
			WeekOrder:      database.Int2(1),
			Name:           database.Text("Week 1"),
			StartDate:      pgtype.Date{Time: time.Date(2023, 04, 01, 00, 0, 0, 0, time.Local)},
			EndDate:        pgtype.Date{Time: time.Date(2023, 04, 8, 0, 0, 0, 0, time.Local)},
			Period:         database.Text("Term 1"),
			AcademicYearID: database.Text(academicYearID),
			LocationID:     database.Text(locations[2].LocationID),
		},
	}

	closedDays := []*repo.AcademicClosedDay{
		{
			AcademicClosedDayID: database.Text("academic_closed_day_01"),
			Date:                pgtype.Date{Time: time.Date(2023, 04, 05, 00, 0, 0, 0, time.Local)},
			AcademicWeekID:      database.Text("academic_week_01"),
			AcademicYearID:      database.Text(academicYearID),
			LocationID:          database.Text(locations[0].LocationID),
		},
		{
			AcademicClosedDayID: database.Text("academic_closed_day_02"),
			Date:                pgtype.Date{Time: time.Date(2023, 04, 05, 00, 0, 0, 0, time.Local)},
			AcademicWeekID:      database.Text("academic_week_02"),
			AcademicYearID:      database.Text(academicYearID),
			LocationID:          database.Text(locations[1].LocationID),
		},
		{
			AcademicClosedDayID: database.Text("academic_closed_day_03"),
			Date:                pgtype.Date{Time: time.Date(2023, 04, 05, 00, 0, 0, 0, time.Local)},
			AcademicWeekID:      database.Text("academic_week_03"),
			AcademicYearID:      database.Text(academicYearID),
			LocationID:          database.Text(locations[2].LocationID),
		},
	}

	tcs := []TestCase{
		{
			name: "GetChildLocations failed",
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Once().Return(nil, fmt.Errorf("error"))
			},
			expectedErr: fmt.Errorf("a.LocationRepo.GetChildLocations failed: %w", fmt.Errorf("error")),
		},
		{
			name: "GetAcademicYearByID failed",
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Once().Return(locations, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, academicYearID).Once().Return(nil, fmt.Errorf("error"))
			},
			expectedErr: fmt.Errorf("a.AcademicYearRepo.GetAcademicYearByID failed: %w", fmt.Errorf("error")),
		},
		{
			name: "GetAcademicWeeksByYearAndLocationIDs failed",
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Once().Return(locations, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, academicYearID).Return(academicYear, nil)
				academicWeekRepo.On("GetAcademicWeeksByYearAndLocationIDs", ctx, masterDB, academicYearID, mock.Anything).
					Once().Return(nil, fmt.Errorf("error"))
			},
			expectedErr: fmt.Errorf("a.AcademicWeekRepo.GetAcademicWeeksByYearAndLocationIDs failed: %w", fmt.Errorf("error")),
		},
		{
			name:         "GetAcademicClosedDayByWeeks failed",
			expectedResp: nil,
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Once().Return(locations, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, academicYearID).Return(academicYear, nil)
				academicWeekRepo.On("GetAcademicWeeksByYearAndLocationIDs", ctx, masterDB, academicYearID, mock.Anything).
					Once().Return(weeks, nil)
				academicClosedDayRepo.On("GetAcademicClosedDayByWeeks", ctx, masterDB, mock.Anything).
					Once().Return(nil, fmt.Errorf("error"))
			},
			expectedErr: fmt.Errorf("a.AcademicWeekRepo.GetAcademicClosedDayByWeeks failed: %w", fmt.Errorf("error")),
		},
		{
			name: "success",
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Once().Return(locations, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, academicYearID).Return(academicYear, nil)
				academicWeekRepo.On("GetAcademicWeeksByYearAndLocationIDs", ctx, masterDB, academicYearID, mock.Anything).
					Once().Return(weeks, nil)
				academicClosedDayRepo.On("GetAcademicClosedDayByWeeks", ctx, masterDB, mock.Anything).
					Once().Return(closedDays, nil)
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			academicCalendarQueryHandler := &AcademicCalendarQueryHandler{
				MasterDB:              masterDB,
				BobDB:                 bobDB,
				AcademicYearRepo:      academicYearRepo,
				AcademicWeekRepo:      academicWeekRepo,
				LocationRepo:          locationRepo,
				LocationTypeRepo:      locationTypeRepo,
				ConfigRepo:            configRepo,
				AcademicClosedDayRepo: academicClosedDayRepo,
			}
			resp, err := academicCalendarQueryHandler.ExportAcademicCalendar(ctx, academicYearID, reqLocationID)

			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
				assert.Empty(t, resp)
			} else {
				assert.NotEmpty(t, resp)
			}
			mock.AssertExpectationsForObjects(t, masterDB, academicWeekRepo, locationRepo, locationTypeRepo, configRepo)
		})
	}
}
