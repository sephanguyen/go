package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	domain "github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"
	location_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_academic_year_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/academic_year/infrastructure/repo"
	mock_configuration_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/configuration/infrastructure/repo"
	mock_location_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAcademicYearService_ImportAcademicCalendar(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	masterDB := &mock_database.Ext{}
	bobDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	academicWeekRepo := new(mock_academic_year_repo.MockAcademicWeekRepo)
	academicYearRepo := new(mock_academic_year_repo.MockAcademicYearRepo)
	academicClosedDayRepo := new(mock_academic_year_repo.MockAcademicClosedDayRepo)
	locationRepo := new(mock_location_repo.MockLocationRepo)
	locationTypeRepo := new(mock_location_repo.MockLocationTypeRepo)
	configRepo := new(mock_configuration_repo.MockConfigRepo)

	academicYearService := NewAcademicYearService(masterDB, bobDB, academicYearRepo, academicWeekRepo, academicClosedDayRepo, locationRepo, locationTypeRepo, configRepo)

	now := time.Now()

	academicYear := &domain.AcademicYear{
		AcademicYearID: "academic_year_id",
		Name:           "2023",
		StartDate:      now,
		EndDate:        now.Add(24 * 7 * time.Hour),
	}

	reqLocationID := "location-1"

	locationsUnderReqLocation := []*location_domain.Location{
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

	testCases := []struct {
		name             string
		req              *mpb.ImportAcademicCalendarRequest
		expectedResp     *mpb.ImportAcademicCalendarResponse
		expectedErr      error
		setup            func(ctx context.Context)
		expectedErrModel *errdetails.BadRequest
	}{
		{
			name: "success",
			req: &mpb.ImportAcademicCalendarRequest{
				Payload: []byte(`order,academic_week,start_date,end_date,period,academic_closed_day
				1,Week1,2023-04-01,2023-04-07,Term 1,2023-04-01;2023-04-04
				2,Week2,2023-04-11,2023-04-16,Term 1,2023-04-12`),
				LocationId:         reqLocationID,
				AcademicYearId:     "temp_academic_year_id",
				AcademicClosedDays: []string{"2023-04-08", "2023-04-10"},
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, mock.Anything).Return(academicYear, nil)
				masterDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				academicWeekRepo.On("Insert", ctx, tx, mock.Anything).Return(nil)
				academicClosedDayRepo.On("Insert", ctx, tx, mock.Anything).Return(nil)
			},
			expectedResp: &mpb.ImportAcademicCalendarResponse{Errors: []*mpb.ImportAcademicCalendarResponse_ImportAcademicCalendarError{}},
		},
		{
			name: "error when have invalid header",
			req: &mpb.ImportAcademicCalendarRequest{
				Payload: []byte(`order,academic_week,start_date,end_date,invalid_header,academic_closed_day
				1,Week1,2023-04-01,2023-04-07,Term 1,2023-04-03;2023-04-05
				2,Week2,2023-04-11,fake,Term 1,`),
				LocationId:         "location-1",
				AcademicYearId:     "temp_academic_year_id",
				AcademicClosedDays: []string{"2023-04-08", "2023-04-09", "2023-04-10"},
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, mock.Anything).Return(academicYear, nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "csv has invalid format, column number 5 should be period, got invalid_header"),
		},
		{
			name: "error when miss header",
			req: &mpb.ImportAcademicCalendarRequest{
				Payload: []byte(`order,academic_week,start_date,end_date,academic_closed_day
				1,Week1,2023-04-01,2023-04-07,Term 1,2023-04-03;2023-04-04
				2,Week2,2023-04-11,fake,Term 1,`),
				LocationId:         "location-1",
				AcademicYearId:     "temp_academic_year_id",
				AcademicClosedDays: []string{"2023-04-08", "2023-04-09", "2023-04-10"},
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, mock.Anything).Return(academicYear, nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
		},
		{
			name: "error when have redundant header",
			req: &mpb.ImportAcademicCalendarRequest{
				Payload: []byte(`order,academic_week,start_date,end_date,period,academic_closed_day,redundant_header
				1,Week1,2023-03-31,2023-04-07,Summer,2023-04-03;2023-04-04
				2,Week2,2023-04-11,fake,Summer,`),
				LocationId:         "location-1",
				AcademicYearId:     "temp_academic_year_id",
				AcademicClosedDays: []string{"2023-04-08", "2023-04-09", "2023-04-10"},
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, mock.Anything).Return(academicYear, nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
		},
		{
			name: "closed day is out of week",
			req: &mpb.ImportAcademicCalendarRequest{
				Payload: []byte(`order,academic_week,start_date,end_date,period,academic_closed_day
				1,Week1,2023-04-01,2023-04-07,Term 1,2023-04-03;2023-04-04
				2,Week2,2023-04-11,2023-04-14,Term 1,2023-04-16`),
				LocationId:         "location-1",
				AcademicYearId:     "temp_academic_year_id",
				AcademicClosedDays: []string{"2023-04-08", "2023-04-09", "2023-04-10"},
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, mock.Anything).Return(academicYear, nil)
			},
			expectedResp: &mpb.ImportAcademicCalendarResponse{Errors: []*mpb.ImportAcademicCalendarResponse_ImportAcademicCalendarError{
				{
					RowNumber: 3,
					Error:     "Closed day of week must be day of week",
				},
			}},
		},
		{
			name: "order is invalid number",
			req: &mpb.ImportAcademicCalendarRequest{
				Payload: []byte(`order,academic_week,start_date,end_date,period,academic_closed_day
				one,Week1,2023-03-30,2023-04-08,Term 1,2023-04-02;2023-04-03;2023-04-04
				2,Week2,2023-04-11,2023-04-14,Term 1,2023-04-12`),
				LocationId:         reqLocationID,
				AcademicYearId:     "academic-year-id-01",
				AcademicClosedDays: []string{"2023-04-09", "2023-04-10"},
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, mock.Anything).Return(academicYear, nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "order is not a valid integer",
					},
				},
			},
		},
		{
			name: "error with invalid end date",
			req: &mpb.ImportAcademicCalendarRequest{
				Payload: []byte(`order,academic_week,start_date,end_date,period,academic_closed_day
				1,Week1,2023-04-01,2023-04-07,Term 1,2023-04-03;2023-04-04
				2,Week2,2023-04-11,fake,Term 1,`),
				LocationId:         "location-1",
				AcademicYearId:     "temp_academic_year_id",
				AcademicClosedDays: []string{"2023-04-08", "2023-04-09", "2023-04-10"},
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, mock.Anything).Return(academicYear, nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "data is not valid, please asd check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 3",
						Description: "Invalid format end_date",
					},
				},
			},
		},
		{
			name: "error with missing start date and invalid end date",
			req: &mpb.ImportAcademicCalendarRequest{
				Payload: []byte(`order,academic_week,start_date,end_date,period,academic_closed_day
				1,Week1,,2023-04-07,Term 1,2023-04-03;2023-04-04
				2,Week2,2023-04-11,fake,Term 1,abc`),
				LocationId:         "location-1",
				AcademicYearId:     "temp_academic_year_id",
				AcademicClosedDays: []string{"2023-04-08", "2023-04-09", "2023-04-10"},
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, mock.Anything).Return(academicYear, nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "Invalid format start_date",
					},
					{
						Field:       "Row Number: 3",
						Description: "Invalid format end_date",
					},
				},
			},
		},
		{
			name: "error with invalid closed day",
			req: &mpb.ImportAcademicCalendarRequest{
				Payload: []byte(`order,academic_week,start_date,end_date,period,academic_closed_day
				1,Week1,2023-04-01,2023-04-07,Term 1,
				2,Week2,2023-04-11,2023-04-14,Term 1,fake`),
				LocationId:         "location-1",
				AcademicYearId:     "temp_academic_year_id",
				AcademicClosedDays: []string{"2023-04-08", "2023-04-09", "2023-04-10"},
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
				academicYearRepo.On("GetAcademicYearByID", ctx, masterDB, mock.Anything).Return(academicYear, nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 3",
						Description: "Invalid format academic_closed_day",
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			resp, err := academicYearService.ImportAcademicCalendar(ctx, tc.req)
			if tc.expectedErr != nil {
				if tc.expectedErrModel != nil {
					utils.AssertBadRequestErrorModel(t, tc.expectedErrModel, err)
				} else {
					assert.Equal(t, tc.expectedErr, err)
				}
			} else {
				assert.NotNil(t, resp)
				expectedResp := tc.expectedResp
				for i, err := range resp.Errors {
					assert.Equal(t, err.RowNumber, expectedResp.Errors[i].RowNumber)
					assert.Equal(t, err.Error, expectedResp.Errors[i].Error)
				}

			}
			mock.AssertExpectationsForObjects(t, masterDB, academicWeekRepo)
			mock.AssertExpectationsForObjects(t, masterDB, academicYearRepo)
			mock.AssertExpectationsForObjects(t, masterDB, academicClosedDayRepo)
		})
	}
}

func TestAcademicCalendarQueryHandler_RetrieveLocationsForAcademic(t *testing.T) {
	const locationTypeLevelKey = "mastermgmt.academic_calendar.location_type_level"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()

	masterDB := &mock_database.Ext{}
	bobDB := &mock_database.Ext{}

	weekRepo := new(mock_academic_year_repo.MockAcademicWeekRepo)
	yearRepo := new(mock_academic_year_repo.MockAcademicYearRepo)
	closedDayRepo := new(mock_academic_year_repo.MockAcademicClosedDayRepo)
	locationRepo := new(mock_location_repo.MockLocationRepo)
	locationTypeRepo := new(mock_location_repo.MockLocationTypeRepo)
	configRepo := new(mock_configuration_repo.MockConfigRepo)

	academicYearId := "academic-year-id"

	academicYearService := NewAcademicYearService(masterDB, bobDB, yearRepo, weekRepo, closedDayRepo, locationRepo, locationTypeRepo, configRepo)

	mockAcademicYear := &domain.AcademicYear{
		AcademicYearID: academicYearId,
		Name:           "2024",
		StartDate:      now.Add(24 * time.Hour),
		EndDate:        now.Add(24 * 8 * time.Hour),
	}

	tcs := []struct {
		name         string
		expectedResp interface{}
		setup        func(ctx context.Context)
		expectedErr  error
	}{
		{
			name:         "verify year err",
			expectedResp: []*location_domain.Location{},
			setup: func(ctx context.Context) {
				yearRepo.On("GetAcademicYearByID", ctx, masterDB, mock.Anything).Once().Return(nil, fmt.Errorf("verify year err"))
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Sprintf("failed to get academic year by id: %v", fmt.Errorf("verify year err"))),
		},
		{
			name:         "GetLocationsForAcademic fail",
			expectedResp: []*location_domain.Location{},
			setup: func(ctx context.Context) {
				yearRepo.On("GetAcademicYearByID", ctx, masterDB, mock.Anything).Once().Return(mockAcademicYear, nil)
				configRepo.On("GetByKey", ctx, masterDB, locationTypeLevelKey).Once().Return(nil, fmt.Errorf("query config err"))
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("GetByKey: %s %v", locationTypeLevelKey, fmt.Errorf("query config err"))),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)

			resp, err := academicYearService.RetrieveLocationsForAcademic(ctx, &mpb.RetrieveLocationsForAcademicRequest{
				AcademicYearId: academicYearId,
			})

			if tc.expectedErr != nil {
				assert.ErrorIs(t, tc.expectedErr, err)
				assert.Empty(t, resp)
			} else {
				assert.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}

func TestAcademicCalendarQueryHandler_RetrieveLocationsByLocationTypeLevelConfig(t *testing.T) {
	const locationTypeLevelKey = "mastermgmt.academic_calendar.location_type_level"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	now := time.Now()

	masterDB := &mock_database.Ext{}
	bobDB := &mock_database.Ext{}

	weekRepo := new(mock_academic_year_repo.MockAcademicWeekRepo)
	yearRepo := new(mock_academic_year_repo.MockAcademicYearRepo)
	closedDayRepo := new(mock_academic_year_repo.MockAcademicClosedDayRepo)
	locationRepo := new(mock_location_repo.MockLocationRepo)
	locationTypeRepo := new(mock_location_repo.MockLocationTypeRepo)
	configRepo := new(mock_configuration_repo.MockConfigRepo)

	academicYearId := "academic-year-id"

	academicYearService := NewAcademicYearService(masterDB, bobDB, yearRepo, weekRepo, closedDayRepo, locationRepo, locationTypeRepo, configRepo)

	mockAcademicYear := &domain.AcademicYear{
		AcademicYearID: academicYearId,
		Name:           "2024",
		StartDate:      now.Add(24 * time.Hour),
		EndDate:        now.Add(24 * 8 * time.Hour),
	}

	tcs := []struct {
		name         string
		expectedResp interface{}
		setup        func(ctx context.Context)
		expectedErr  error
	}{
		{
			name:         "GetLocationsByLocationTypeLevelConfig fail",
			expectedResp: []*location_domain.Location{},
			setup: func(ctx context.Context) {
				yearRepo.On("GetLocationsByLocationTypeLevelConfig", ctx, masterDB, mock.Anything).Once().Return(mockAcademicYear, nil)
				configRepo.On("GetByKey", ctx, masterDB, locationTypeLevelKey).Once().Return(nil, fmt.Errorf("query config err"))
			},
			expectedErr: status.Error(codes.Internal, fmt.Sprintf("GetByKey: %s %v", locationTypeLevelKey, fmt.Errorf("query config err"))),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)

			resp, err := academicYearService.RetrieveLocationsByLocationTypeLevelConfig(ctx, &mpb.RetrieveLocationsByLocationTypeLevelConfigRequest{})

			if tc.expectedErr != nil {
				assert.ErrorIs(t, tc.expectedErr, err)
				assert.Empty(t, resp)
			} else {
				assert.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}
