package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLocationService_GetLocationNameByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		locationRepo *mockRepositories.MockLocationRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get by id for update",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				locationRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Location{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when get location without locationID",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "This location with id %s does not exist in the system", constant.LocationID),
			Setup: func(ctx context.Context) {
				locationRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Location{}, nil)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				locationRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Location{
					LocationID: pgtype.Text{Status: pgtype.Present, String: constant.LocationID},
					Name:       pgtype.Text{Status: pgtype.Present, String: constant.LocationName},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			locationRepo = new(mockRepositories.MockLocationRepo)
			testCase.Setup(testCase.Ctx)
			s := &LocationService{
				locationRepo: locationRepo,
			}
			locationName, err := s.GetLocationNameByID(testCase.Ctx, db, constant.LocationID)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, constant.LocationName, locationName)
			}

			mock.AssertExpectationsForObjects(t, db, locationRepo)
		})
	}
}

func TestLocationService_GetLocationsByIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		locationRepo *mockRepositories.MockLocationRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Failed case: Error when getting locations by ids",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				locationRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.Location{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Success case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				locationRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.Location{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			locationRepo = new(mockRepositories.MockLocationRepo)
			testCase.Setup(testCase.Ctx)
			s := &LocationService{
				locationRepo: locationRepo,
			}
			_, err := s.GetLocationsByIDs(testCase.Ctx, db, []string{"1", "2"})

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, locationRepo)
		})
	}
}

func TestLocationService_GetLowestGrantedLocationsForCreatingOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		locationRepo *mockRepositories.MockLocationRepo
	)
	expectedResp := []*pb.LocationInfo{
		{
			LocationId:   "location_id_1",
			LocationName: "location_name_1",
		},
		{
			LocationId:   "location_id_2",
			LocationName: "location_name_2",
		},
	}
	testcases := []utils.TestCase{
		{
			Name:        "Failed case: Error when user id is empty",
			Ctx:         interceptors.ContextWithUserID(ctx, ""),
			Req:         &pb.GetLocationsForCreatingOrderRequest{},
			ExpectedErr: fmt.Errorf("cannot get userID from context"),
			Setup:       func(ctx context.Context) {},
		},
		{
			Name: "Failed case: Error when get lowest granted locations by user_id and permissions",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.GetLocationsForCreatingOrderRequest{
				Name:  "",
				Limit: 30,
			},
			ExpectedErr: status.Errorf(codes.Internal, "Error when get lowest granted locations by user_id and permissions: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				locationRepo.On("GetLowestGrantedLocationIDsByUserIDAndPermissions", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Failed case: Error when get locations by ids",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         &pb.GetLocationsForCreatingOrderRequest{},
			ExpectedErr: status.Errorf(codes.Internal, "Error when get locations by ids: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				locationRepo.On("GetLowestGrantedLocationIDsByUserIDAndPermissions", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{"location_id_1", "location_id_2"}, nil)
				locationRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.Location{}, constant.ErrDefault)
			},
		},
		{
			Name:         "Happy case",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:          &pb.GetLocationsForCreatingOrderRequest{},
			ExpectedErr:  nil,
			ExpectedResp: expectedResp,
			Setup: func(ctx context.Context) {
				locationRepo.On("GetLowestGrantedLocationIDsByUserIDAndPermissions", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{"location_id_1", "location_id_2"}, nil)
				locationRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.Location{
					{
						LocationID: pgtype.Text{
							String: "location_id_1",
							Status: pgtype.Present,
						},
						Name: pgtype.Text{
							String: "location_name_1",
							Status: pgtype.Present,
						},
					},
					{
						LocationID: pgtype.Text{
							String: "location_id_2",
							Status: pgtype.Present,
						},
						Name: pgtype.Text{
							String: "location_name_2",
							Status: pgtype.Present,
						},
					},
				}, nil)
			},
		},
		{
			Name:         "Happy case (locations is empty)",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:          &pb.GetLocationsForCreatingOrderRequest{},
			ExpectedErr:  nil,
			ExpectedResp: []*pb.LocationInfo{},
			Setup: func(ctx context.Context) {
				locationRepo.On("GetLowestGrantedLocationIDsByUserIDAndPermissions", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
				locationRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.Location{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			locationRepo = new(mockRepositories.MockLocationRepo)
			testCase.Setup(testCase.Ctx)
			s := &LocationService{
				locationRepo: locationRepo,
			}
			req := testCase.Req.(*pb.GetLocationsForCreatingOrderRequest)
			resp, err := s.GetLowestGrantedLocationsForCreatingOrder(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, testCase.ExpectedResp, resp)
			}

			mock.AssertExpectationsForObjects(t, db, locationRepo)
		})
	}
}
