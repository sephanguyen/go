package timesheet

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/timesheet/repository"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLocationService_GetListGrantedLocationOfStaff(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		locationRepo       = new(mock_repositories.MockLocationRepoImpl)
		db                 = new(mock_database.Ext)
		listNilLocationE   = []*entity.Location{}
		listNilLocationDTO = []*dto.Location{}
	)

	s := LocationServiceImpl{
		DB:           db,
		LocationRepo: locationRepo,
	}

	locationDTO1 := &dto.Location{
		LocationID: "location-1",
		Name:       "Location 1",
	}
	locationDTO2 := &dto.Location{
		LocationID: "location-2",
		Name:       "Location 2",
	}
	locationDTO3 := &dto.Location{
		LocationID: "location-3",
		Name:       "Location 3",
	}

	listLocationDto := []*dto.Location{locationDTO1, locationDTO2, locationDTO3}
	list1LocationDto := []*dto.Location{locationDTO1}
	list2LocationDto := []*dto.Location{locationDTO1, locationDTO2}
	onlyLocation3Dto := []*dto.Location{locationDTO3}

	locationEntity1 := &entity.Location{
		LocationID: database.Text("location-1"),
		Name:       database.Text("Location 1"),
	}
	locationEntity2 := &entity.Location{
		LocationID: database.Text("location-2"),
		Name:       database.Text("Location 2"),
	}

	locationChildEntity1 := &entity.Location{
		LocationID: database.Text("location-1"),
		Name:       database.Text("Location 1"),
	}
	locationChildEntity2 := &entity.Location{
		LocationID: database.Text("location-2"),
		Name:       database.Text("Location 2"),
	}
	locationChildEntity3 := &entity.Location{
		LocationID: database.Text("location-3"),
		Name:       database.Text("Location 3"),
	}
	listParentLocationE := []*entity.Location{locationEntity1}
	listLocationChildE := []*entity.Location{locationChildEntity1, locationChildEntity2, locationChildEntity3}
	list2ParentLocationE := []*entity.Location{locationEntity1, locationEntity2}
	list23LocationChildE := []*entity.Location{locationChildEntity2, locationChildEntity3}
	testCases := []TestCase{
		{
			name:         "happy case get list locations success",
			ctx:          ctx,
			staffID:      "staff-1",
			locationName: "",
			limit:        10,
			expectedErr:  nil,
			expectedResp: listLocationDto,
			setup: func(ctx context.Context) {
				locationRepo.On("GetGrantedLocationOfStaff", ctx, db, mock.Anything).
					Return(listParentLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
			},
		},
		{
			name:         "happy case get one parents and 2 child location",
			ctx:          ctx,
			staffID:      "staff-1",
			locationName: "",
			limit:        10,
			expectedErr:  nil,
			expectedResp: listLocationDto,
			setup: func(ctx context.Context) {
				locationRepo.On("GetGrantedLocationOfStaff", ctx, db, mock.Anything).
					Return(list2ParentLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listNilLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(list23LocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(list23LocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(list23LocationChildE, nil).Once()
			},
		},
		{
			name:         "happy case get only 2 parent location success",
			ctx:          ctx,
			staffID:      "staff-1",
			locationName: "",
			limit:        10,
			expectedErr:  nil,
			expectedResp: list2LocationDto,
			setup: func(ctx context.Context) {
				locationRepo.On("GetGrantedLocationOfStaff", ctx, db, mock.Anything).
					Return(list2ParentLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listNilLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listNilLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
			},
		},
		{
			name:         "happy case search not match any records",
			ctx:          ctx,
			staffID:      "staff-1",
			locationName: "test-test",
			limit:        10,
			expectedErr:  nil,
			expectedResp: listNilLocationDTO,
			setup: func(ctx context.Context) {
				locationRepo.On("GetGrantedLocationOfStaff", ctx, db, mock.Anything).
					Return(listParentLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
			},
		},
		{
			name:         "error case pgx no row when get parents records",
			ctx:          ctx,
			staffID:      "staff-1",
			locationName: "",
			limit:        10,
			expectedErr:  nil,
			expectedResp: listNilLocationDTO,
			setup: func(ctx context.Context) {
				locationRepo.On("GetGrantedLocationOfStaff", ctx, db, mock.Anything).
					Return(listNilLocationE, pgx.ErrNoRows).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
			},
		},
		{
			name:         "happy case not get any locations",
			ctx:          ctx,
			staffID:      "staff-1",
			locationName: "",
			limit:        10,
			expectedErr:  nil,
			expectedResp: listNilLocationDTO,
			setup: func(ctx context.Context) {
				locationRepo.On("GetGrantedLocationOfStaff", ctx, db, mock.Anything).
					Return(listNilLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
			},
		},
		{
			name:         "happy case get only return match search record and limit 2",
			ctx:          ctx,
			staffID:      "staff-1",
			locationName: "Location",
			limit:        2,
			expectedErr:  nil,
			expectedResp: list2LocationDto,
			setup: func(ctx context.Context) {
				locationRepo.On("GetGrantedLocationOfStaff", ctx, db, mock.Anything).
					Return(listParentLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
			},
		},
		{
			name:         "happy case get only return 2 records",
			ctx:          ctx,
			staffID:      "staff-1",
			locationName: "",
			limit:        2,
			expectedErr:  nil,
			expectedResp: list2LocationDto,
			setup: func(ctx context.Context) {
				locationRepo.On("GetGrantedLocationOfStaff", ctx, db, mock.Anything).
					Return(listParentLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
			},
		},
		{
			name:         "happy case get only return 1 records",
			ctx:          ctx,
			staffID:      "staff-1",
			locationName: "",
			limit:        1,
			expectedErr:  nil,
			expectedResp: list1LocationDto,
			setup: func(ctx context.Context) {
				locationRepo.On("GetGrantedLocationOfStaff", ctx, db, mock.Anything).
					Return(listParentLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
			},
		},
		{
			name:         "happy case get only return match search record",
			ctx:          ctx,
			staffID:      "staff-1",
			locationName: "3",
			limit:        10,
			expectedErr:  nil,
			expectedResp: onlyLocation3Dto,
			setup: func(ctx context.Context) {
				locationRepo.On("GetGrantedLocationOfStaff", ctx, db, mock.Anything).
					Return(listParentLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
			},
		},
		{
			name:         "happy case get only return match search record 2",
			ctx:          ctx,
			staffID:      "staff-1",
			locationName: "Location",
			limit:        10,
			expectedErr:  nil,
			expectedResp: listLocationDto,
			setup: func(ctx context.Context) {
				locationRepo.On("GetGrantedLocationOfStaff", ctx, db, mock.Anything).
					Return(listParentLocationE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
				locationRepo.On("GetListChildLocations", ctx, db, mock.Anything).
					Return(listLocationChildE, nil).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.GetListGrantedLocationOfStaff(testCase.ctx, testCase.staffID, testCase.locationName, testCase.limit)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
