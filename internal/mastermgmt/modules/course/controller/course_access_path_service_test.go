package controller

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure/repo"
	locationDomain "github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	mockDatabase "github.com/manabie-com/backend/mock/golibs/database"
	mockCourseRepo "github.com/manabie-com/backend/mock/mastermgmt/modules/course/infrastructure/repo"
	mockLocationRepo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCourseAccessPathService_ImportCourseAccessPaths(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("Import successfully", func(t *testing.T) {
		t.Parallel()

		//arrange
		db := &mockDatabase.Ext{}
		tx := &mockDatabase.Tx{}
		capRepo := new(mockCourseRepo.MockCourseAccessPathRepo)
		courseRepo := new(mockCourseRepo.MockCourseRepo)
		locationRepo := new(mockLocationRepo.MockLocationRepo)
		service := NewCourseAccessPathService(db, capRepo, locationRepo, courseRepo)
		req := &mpb.ImportCourseAccessPathsRequest{Payload: []byte(`course_access_path_id,course_id,location_id
			,course_id_1,location_id_1
			,course_id_2,location_id_2
			,course_id_3,location_id_3`)}
		locationRepo.On("GetLocationsByLocationIDs",
			mock.Anything,
			db,
			database.TextArray([]string{"location_id_1", "location_id_2", "location_id_3"}),
			false).
			Return([]*locationDomain.Location{
				{
					LocationID: "location_id_1",
				},
				{
					LocationID: "location_id_3",
				},
				{
					LocationID: "location_id_2",
				},
			}, nil).
			Once()
		courseRepo.On("GetByIDs",
			mock.Anything,
			db,
			[]string{"course_id_1", "course_id_2", "course_id_3"}).
			Return([]*domain.Course{
				{
					CourseID: "course_id_1",
				},
				{
					CourseID: "course_id_3",
				},
				{
					CourseID: "course_id_2",
				},
			}, nil).
			Once()
		capRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Once()
		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil).Once()
		tx.On("Commit", mock.Anything).Return(nil).Once()

		// act
		res, err := service.ImportCourseAccessPaths(ctx, req)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, &mpb.ImportCourseAccessPathsResponse{}, res)

		mock.AssertExpectationsForObjects(t, locationRepo)
		mock.AssertExpectationsForObjects(t, courseRepo)
		mock.AssertExpectationsForObjects(t, db)
		mock.AssertExpectationsForObjects(t, capRepo)
	})

	t.Run("Import successfully more than chunk size's items", func(t *testing.T) {
		t.Parallel()

		//arrange
		db := &mockDatabase.Ext{}
		tx := &mockDatabase.Tx{}
		capRepo := new(mockCourseRepo.MockCourseAccessPathRepo)
		courseRepo := new(mockCourseRepo.MockCourseRepo)
		locationRepo := new(mockLocationRepo.MockLocationRepo)
		caps := fakeManyAccessPaths(250)
		capChunks := sliceutils.Chunk(caps, commands.ChunkSize)
		csvContent := ""
		for _, v := range caps {
			csvContent += fmt.Sprintf(",%s,%s\n", v.CourseID, v.LocationID)
		}
		service := NewCourseAccessPathService(db, capRepo, locationRepo, courseRepo)
		req := &mpb.ImportCourseAccessPathsRequest{Payload: []byte(fmt.Sprintf(`course_access_path_id,course_id,location_id
			%s`, csvContent))}
		for _, chunk := range capChunks {
			locIDs := sliceutils.Map(chunk, func(c domain.CourseAccessPath) string {
				return c.LocationID
			})
			courseIDs := sliceutils.Map(chunk, func(c domain.CourseAccessPath) string {
				return c.CourseID
			})
			courses := sliceutils.Map(chunk, func(c domain.CourseAccessPath) *domain.Course {
				return &domain.Course{
					CourseID: c.CourseID,
				}
			})
			locations := sliceutils.Map(chunk, func(c domain.CourseAccessPath) *locationDomain.Location {
				return &locationDomain.Location{
					LocationID: c.LocationID,
				}
			})
			locationRepo.On("GetLocationsByLocationIDs",
				mock.Anything,
				mock.Anything,
				database.TextArray(locIDs),
				false).
				Return(locations, nil).
				Once()
			courseRepo.On("GetByIDs",
				mock.Anything,
				mock.Anything,
				courseIDs).
				Return(courses, nil).
				Once()

			capRepo.On("Upsert", mock.Anything, mock.Anything, mock.Anything).
				Return(nil)
		}
		db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)

		// act
		res, err := service.ImportCourseAccessPaths(ctx, req)

		// assert
		assert.Nil(t, err)
		assert.Equal(t, &mpb.ImportCourseAccessPathsResponse{}, res)

		mock.AssertExpectationsForObjects(t, locationRepo)
		mock.AssertExpectationsForObjects(t, courseRepo)
		mock.AssertExpectationsForObjects(t, db)
		mock.AssertExpectationsForObjects(t, capRepo)
	})

	t.Run("Import failed with missing locations and courses", func(t *testing.T) {
		t.Parallel()
		//arrange
		db := &mockDatabase.Ext{}
		capRepo := new(mockCourseRepo.MockCourseAccessPathRepo)
		courseRepo := new(mockCourseRepo.MockCourseRepo)
		locationRepo := new(mockLocationRepo.MockLocationRepo)
		service := NewCourseAccessPathService(db, capRepo, locationRepo, courseRepo)
		req := &mpb.ImportCourseAccessPathsRequest{Payload: []byte(`course_access_path_id,course_id,location_id
			,course_id_1,location_id_1
			,course_id_2,location_id_2
			,course_id_3,location_id_3`)}
		locationRepo.On("GetLocationsByLocationIDs",
			mock.Anything,
			db,
			database.TextArray([]string{"location_id_1", "location_id_2", "location_id_3"}),
			false).
			Return([]*locationDomain.Location{
				{
					LocationID: "location_id_3",
				},
			}, nil).
			Once()
		courseRepo.On("GetByIDs",
			mock.Anything,
			db,
			[]string{"course_id_1", "course_id_2", "course_id_3"}).
			Return([]*domain.Course{
				{
					CourseID: "course_id_1",
				}, {
					CourseID: "course_id_2",
				},
			}, nil).
			Once()
		expectedBr := &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "Row Number: 2",
					Description: fmt.Sprintf("location id %s is not exist", "location_id_1"),
				},
				{
					Field:       "Row Number: 3",
					Description: fmt.Sprintf("location id %s is not exist", "location_id_2"),
				},
				{
					Field:       "Row Number: 4",
					Description: fmt.Sprintf("course id %s is not exist", "course_id_3"),
				},
			},
		}
		// act
		res, err := service.ImportCourseAccessPaths(ctx, req)

		// assert
		assert.Nil(t, res)
		utils.AssertBadRequestErrorModel(t, expectedBr, err)

		mock.AssertExpectationsForObjects(t, locationRepo)
		mock.AssertExpectationsForObjects(t, courseRepo)
	})
}

func TestCourseAccessPathService_ImportCourseAccessPaths_CSV_Validation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db := &mockDatabase.Ext{}
	capRepo := new(mockCourseRepo.MockCourseAccessPathRepo)
	courseRepo := new(mockCourseRepo.MockCourseRepo)
	locationRepo := new(mockLocationRepo.MockLocationRepo)
	service := NewCourseAccessPathService(db, capRepo, locationRepo, courseRepo)
	testCases := []TestCase{
		{
			name:        "no data in csv file",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
			req:         &mpb.ImportCourseAccessPathsRequest{},
		},
		{
			name:        "invalid file - number of column != 3",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "wrong number of columns, expected 3, got 2"),
			req: &mpb.ImportCourseAccessPathsRequest{
				Payload: []byte(`name,display_name
							1,Course 1`),
			},
		},
		{
			name:        "invalid file - first column name (toLowerCase) != course_access_path_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 1 should be course_access_path_id, got course_access_path_idz"),
			req: &mpb.ImportCourseAccessPathsRequest{
				Payload: []byte(`course_access_path_idz,course_id,location_id
							1,Course 1,1`),
			},
		},
		{
			name:        "invalid file - second column name (toLowerCase) != course_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 2 should be course_id, got course_idz"),
			req: &mpb.ImportCourseAccessPathsRequest{
				Payload: []byte(`course_access_path_id,course_idz,location_id
							1,Course 1,1`),
			},
		},
		{
			name:        "invalid file - third column name (toLowerCase) != location_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 3 should be location_id, got location_idz"),
			req: &mpb.ImportCourseAccessPathsRequest{
				Payload: []byte(`course_access_path_id,course_id,location_idz
							1,Course 1,1`),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			resp, err := service.ImportCourseAccessPaths(testCase.ctx, testCase.req.(*mpb.ImportCourseAccessPathsRequest))
			if testCase.expectedErr != nil {
				assert.Nil(t, resp)
				if testCase.expectedErrModel != nil {
					utils.AssertBadRequestErrorModel(t, testCase.expectedErrModel, err)
				} else {
					assert.Equal(t, testCase.expectedErr, err)
				}
			} else {
				assert.Equal(t, nil, err)
				assert.NotNil(t, resp)
				mock.AssertExpectationsForObjects(t, courseRepo, capRepo, locationRepo)
			}
		})
	}
}

func TestCourseAccessPathService_ExportCourseAccessPaths(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db := &mockDatabase.Ext{}
	capRepo := new(mockCourseRepo.MockCourseAccessPathRepo)
	courseRepo := new(mockCourseRepo.MockCourseRepo)
	locationRepo := new(mockLocationRepo.MockLocationRepo)
	service := NewCourseAccessPathService(db, capRepo, locationRepo, courseRepo)

	t.Run("Export successfully", func(t *testing.T) {
		// arrange

		courseAPs := []*repo.CourseAccessPath{
			{
				ID:         database.Varchar("ID 1"),
				CourseID:   database.Text("Course 1"),
				LocationID: database.Text("Location 1"),
			},
			{
				ID:         database.Varchar("ID 2"),
				CourseID:   database.Text("Course 2"),
				LocationID: database.Text("Location 2"),
			},
		}

		courseAPStr := `"course_access_path_id","course_id","location_id"` + "\n" +
			`"ID 1","Course 1","Location 1"` + "\n" +
			`"ID 2","Course 2","Location 2"` + "\n"

		capRepo.On("GetAll", ctx, db).Once().Return(courseAPs, nil)
		byteData := []byte(courseAPStr)

		// act
		resp, err := service.ExportCourseAccessPaths(ctx, &mpb.ExportCourseAccessPathsRequest{})

		// assert
		assert.Nil(t, err)
		assert.Equal(t, resp.Data, byteData)
		mock.AssertExpectationsForObjects(t, capRepo)
	})

	t.Run("return internal error when retrieve data failed", func(t *testing.T) {
		// arrange
		capRepo.On("GetAll", ctx, db).Once().
			Return(nil, errors.New("sample error"))

		// act
		resp, err := service.ExportCourseAccessPaths(ctx, &mpb.ExportCourseAccessPathsRequest{})

		// assert
		assert.Nil(t, resp)
		assert.Equal(t, err, status.Error(codes.Internal, "sample error"))
		mock.AssertExpectationsForObjects(t, capRepo)
	})

}

func fakeManyAccessPaths(limit int) []domain.CourseAccessPath {
	caps := make([]domain.CourseAccessPath, limit)
	for i := 0; i < limit; i++ {
		counter := i + 1
		caps[i] = domain.CourseAccessPath{
			ID:         fmt.Sprintf("id_%d", counter),
			LocationID: fmt.Sprintf("location_id_%d", counter),
			CourseID:   fmt.Sprintf("course_id_%d", counter),
		}
	}
	return caps
}
