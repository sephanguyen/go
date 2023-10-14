package controller

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	location_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_location_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mock_working_hours_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/working_hours/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWorkingHoursService_ImportWorkingHours(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	masterDB := &mock_database.Ext{}
	bobDB := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	workingHoursRepo := new(mock_working_hours_repo.MockWorkingHoursRepo)
	locationRepo := new(mock_location_repo.MockLocationRepo)

	WorkingHoursService := NewWorkingHoursService(masterDB, bobDB, workingHoursRepo, locationRepo)

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

	mockLocationIDs := sliceutils.Map(locationsUnderReqLocation, func(l *location_domain.Location) string {
		return l.LocationID
	})

	testCases := []struct {
		name             string
		req              *mpb.ImportWorkingHoursRequest
		expectedResp     *mpb.ImportWorkingHoursResponse
		expectedErr      error
		setup            func(ctx context.Context)
		expectedErrModel *errdetails.BadRequest
	}{
		{
			name: "success",
			req: &mpb.ImportWorkingHoursRequest{
				Payload: []byte(`day,opening_time,closing_time
				Monday,09:00,18:00
				Tuesday,09:00,18:00`),
				LocationId: reqLocationID,
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
				masterDB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				workingHoursRepo.On("Upsert", ctx, tx, mock.Anything, mockLocationIDs).Return(nil)
			},
			expectedResp: &mpb.ImportWorkingHoursResponse{Errors: []*mpb.ImportWorkingHoursResponse_ImportWorkingHoursError{}},
		},
		{
			name: "error when have invalid header",
			req: &mpb.ImportWorkingHoursRequest{
				Payload: []byte(`day,opening_time,invalid_header
				Monday,09:00,18:00
				Tuesday,09:00,18:00`),
				LocationId: "location-1",
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "csv has invalid format, column number 3 should be closing_time, got invalid_header"),
		},
		{
			name: "error when miss header",
			req: &mpb.ImportWorkingHoursRequest{
				Payload: []byte(`day,opening_time
				Monday,09:00,18:00
				Tuesday,09:00,18:00`),
				LocationId: "location-1",
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
		},
		{
			name: "error when have redundant header",
			req: &mpb.ImportWorkingHoursRequest{
				Payload: []byte(`day,opening_time,closing_time,redundant_header
				Monday,09:00,18:00
				Tuesday,09:00,18:00`),
				LocationId: reqLocationID,
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetChildLocations", ctx, bobDB, reqLocationID).Return(locationsUnderReqLocation, nil)
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.InvalidArgument, "record on line 2: wrong number of fields"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			resp, err := WorkingHoursService.ImportWorkingHours(ctx, tc.req)
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
			mock.AssertExpectationsForObjects(t, masterDB, workingHoursRepo)
		})
	}
}
