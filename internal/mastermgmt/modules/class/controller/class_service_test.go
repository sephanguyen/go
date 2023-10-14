package controller

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/application/queries"
	class_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	course "github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_class_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/class/infrastructure/repo"
	mock_course_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/course/infrastructure/repo"
	mock_location_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestClassService_ImportClass(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := new(mock_nats.JetStreamManagement)
	classRepo := new(mock_class_repo.MockClassRepo)
	locationRepo := new(mock_location_repo.MockLocationRepo)
	courseRepo := new(mock_course_repo.MockCourseRepo)

	classService := &ClassService{
		DB:           db,
		JSM:          jsm,
		ClassRepo:    classRepo,
		LocationRepo: locationRepo,
		CourseRepo:   courseRepo,
		ClassCommandHandler: commands.ClassCommandHandler{
			DB:        db,
			ClassRepo: classRepo,
		},
		ClassQueryHandler: queries.ClassQueryHandler{
			DB:        db,
			ClassRepo: classRepo,
		},
	}
	course := &course.Course{
		CourseID: idutil.ULIDNow(),
	}
	location := &domain.Location{
		LocationID: "location-1",
	}

	testCases := []struct {
		name         string
		req          *mpb.ImportClassRequest
		expectedResp *mpb.ImportClassResponse
		expectedErr  error
		setup        func(ctx context.Context)
	}{
		{
			name: "success",
			req: &mpb.ImportClassRequest{
				Payload: []byte(`course_id,location_id,course_name,location_name,class_name
				course-1,location-1,,,class-1
				course-2,location-2,,,class-2`),
			},
			setup: func(ctx context.Context) {
				courseRepo.On("GetByID", ctx, db, mock.Anything).Twice().Return(course, nil)
				locationRepo.On("GetLocationByID", ctx, db, mock.Anything).Twice().Return(location, nil)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				classRepo.On("Insert", ctx, tx, mock.Anything).Once().Return(nil)
				jsm.On("PublishAsyncContext", ctx, constants.SubjectMasterMgmtClassUpserted, mock.Anything).Twice().Return("", nil)
			},
			expectedResp: &mpb.ImportClassResponse{Errors: []*mpb.ImportClassResponse_ImportClassError{}},
		},
		{
			name: "import class with invalid location_id",
			req: &mpb.ImportClassRequest{
				Payload: []byte(`course_id,location_id,course_name,location_name,class_name
				course-1,location-1,,,class-1`),
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetLocationByID", ctx, db, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
			expectedResp: &mpb.ImportClassResponse{Errors: []*mpb.ImportClassResponse_ImportClassError{
				{
					RowNumber: 2,
					Error:     "invalid class: Class.LocationID invalid",
				},
			}},
		},
		{
			name: "import class with invalid course_id",
			req: &mpb.ImportClassRequest{
				Payload: []byte(`course_id,location_id,course_name,location_name,class_name
				course-1,location-1,,,class-1`),
			},
			setup: func(ctx context.Context) {
				locationRepo.On("GetLocationByID", ctx, db, mock.Anything).Once().Return(location, nil)
				courseRepo.On("GetByID", ctx, db, mock.Anything).Once().Return(course, pgx.ErrNoRows)
			},
			expectedResp: &mpb.ImportClassResponse{Errors: []*mpb.ImportClassResponse_ImportClassError{
				{
					RowNumber: 2,
					Error:     "invalid class: Class.CourseID invalid",
				},
			}},
		},
		{
			name: "import with invalid class name format",
			req: &mpb.ImportClassRequest{
				Payload: []byte(fmt.Sprintf(`course_id,location_id,course_name,location_name,class_name
				%s`, fmt.Sprintf("course-1,location-1,,,%s", string([]byte{0xff, 0xfe, 0xfd})))),
			},
			setup: func(ctx context.Context) {},
			expectedResp: &mpb.ImportClassResponse{Errors: []*mpb.ImportClassResponse_ImportClassError{
				{
					RowNumber: 2,
					Error:     "invalid class: Class.Name is not valid UTF8 format",
				},
			}},
		},
		{
			name: "empty payload",
			req: &mpb.ImportClassRequest{
				Payload: []byte{},
			},
			setup:        func(ctx context.Context) {},
			expectedResp: &mpb.ImportClassResponse{},
			expectedErr:  status.Error(codes.InvalidArgument, "no data in csv file"),
		},
		{
			name: "import only have header in csv",
			req: &mpb.ImportClassRequest{
				Payload: []byte(`course_id,location_id,course_name,location_name,class_name`),
			},
			setup:        func(ctx context.Context) {},
			expectedResp: &mpb.ImportClassResponse{},
			expectedErr:  status.Error(codes.InvalidArgument, "no data in csv file"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			resp, err := classService.ImportClass(ctx, tc.req)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NotNil(t, resp)
				expectedResp := tc.expectedResp
				for i, err := range resp.Errors {
					assert.Equal(t, err.RowNumber, expectedResp.Errors[i].RowNumber)
					assert.Equal(t, err.Error, expectedResp.Errors[i].Error)
				}
			}
			mock.AssertExpectationsForObjects(t, db, locationRepo)
			mock.AssertExpectationsForObjects(t, db, courseRepo)
			mock.AssertExpectationsForObjects(t, db, classRepo)

		})
	}
}

func TestGradeService_ExportClasses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	classRepo := new(mock_class_repo.MockClassRepo)

	classes := []*class_domain.ExportingClass{
		{
			ClassID:      "class-id-1",
			Name:         "class-name-1",
			CourseID:     "course-id-1",
			CourseName:   "course-name-1",
			LocationID:   "location-id-1",
			LocationName: "location-name-1",
		},
		{
			ClassID:      "class-id-2",
			Name:         "class-name-2",
			CourseID:     "course-id-2",
			CourseName:   "course-name-2",
			LocationID:   "location-id-2",
			LocationName: "location-name-2",
		},
		{
			ClassID:      "class-id-3",
			Name:         "class-name-3",
			CourseID:     "course-id-3",
			CourseName:   "course-name-3",
			LocationID:   "location-id-3",
			LocationName: "location-name-3",
		},
	}

	gradeStr := `"class_id","class_name","course_id","location_id"` + "\n" +
		`"class-id-1","class-name-1","course-id-1","location-id-1"` + "\n" +
		`"class-id-2","class-name-2","course-id-2","location-id-2"` + "\n" +
		`"class-id-3","class-name-3","course-id-3","location-id-3"` + "\n"

	s := &ClassService{
		ExportClassesQueryHandler: queries.ExportClassesQueryHandler{
			DB:        db,
			ClassRepo: classRepo,
		},
	}

	t.Run("export all data in db with correct column", func(t *testing.T) {
		// arrange
		classRepo.On("GetAll", ctx, db).Once().Return(classes, nil)

		byteData := []byte(gradeStr)

		// act
		resp, err := s.ExportClasses(ctx, &mpb.ExportClassesRequest{})

		// assert
		assert.Nil(t, err)
		assert.Equal(t, resp.Data, byteData)
	})

	t.Run("return internal error when retrieve data failed", func(t *testing.T) {
		// arrange
		classRepo.On("GetAll", ctx, db).Once().Return(nil, errors.New("sample error"))

		// act
		resp, err := s.ExportClasses(ctx, &mpb.ExportClassesRequest{})

		// assert
		assert.Nil(t, resp.Data)
		assert.Equal(t, err, status.Error(codes.Internal, "sample error"))
	})
}

func TestClassService_UpdateClass(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	jsm := new(mock_nats.JetStreamManagement)
	classRepo := new(mock_class_repo.MockClassRepo)

	classService := &ClassService{
		DB:        db,
		JSM:       jsm,
		ClassRepo: classRepo,
		ClassCommandHandler: commands.ClassCommandHandler{
			DB:        db,
			ClassRepo: classRepo,
		},
		ClassQueryHandler: queries.ClassQueryHandler{
			DB:        db,
			ClassRepo: classRepo,
		},
	}
	classID := "class-id"
	className := "name"

	testCases := []struct {
		name         string
		req          *mpb.UpdateClassRequest
		expectedResp *mpb.UpdateClassResponse
		expectedErr  error
		setup        func(ctx context.Context)
	}{
		{
			name: "error with empty class id",
			req: &mpb.UpdateClassRequest{
				ClassId: "",
			},
			expectedResp: &mpb.UpdateClassResponse{},
			expectedErr:  status.Error(codes.InvalidArgument, "`class_id` cannot be empty"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "error with invalid class id",
			req: &mpb.UpdateClassRequest{
				ClassId: classID,
				Name:    className,
			},
			expectedResp: &mpb.UpdateClassResponse{},
			expectedErr:  status.Error(codes.NotFound, class_domain.ErrNotFound.Error()),
			setup: func(ctx context.Context) {
				classRepo.On("UpdateClassNameByID", ctx, db, classID, className).Once().Return(class_domain.ErrNotFound)
			},
		},
		{
			name: "error internal",
			req: &mpb.UpdateClassRequest{
				ClassId: classID,
				Name:    className,
			},
			expectedResp: &mpb.UpdateClassResponse{},
			expectedErr:  status.Error(codes.Internal, pgx.ErrTxClosed.Error()),
			setup: func(ctx context.Context) {
				classRepo.On("UpdateClassNameByID", ctx, db, classID, className).Once().Return(pgx.ErrTxClosed)
			},
		},
		{
			name: "error when update",
			req: &mpb.UpdateClassRequest{
				ClassId: classID,
				Name:    className,
			},
			expectedResp: &mpb.UpdateClassResponse{},
			expectedErr:  status.Error(codes.Internal, pgx.ErrTxCommitRollback.Error()),
			setup: func(ctx context.Context) {
				classRepo.On("UpdateClassNameByID", ctx, db, classID, className).Once().Return(pgx.ErrTxCommitRollback)
			},
		},
		{
			name: "success",
			req: &mpb.UpdateClassRequest{
				ClassId: classID,
				Name:    className,
			},
			expectedResp: &mpb.UpdateClassResponse{},
			setup: func(ctx context.Context) {
				classRepo.On("UpdateClassNameByID", ctx, db, classID, className).Once().Return(nil)
				jsm.On("PublishAsyncContext", ctx, constants.SubjectMasterMgmtClassUpserted, mock.Anything).Once().Return("", nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			gotResp, err := classService.UpdateClass(ctx, tc.req)
			if err != nil {
				assert.Error(t, err)
				assert.Equal(t, err, tc.expectedErr)
				assert.EqualValues(t, gotResp, tc.expectedResp)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, classRepo)
		})
	}
}

func TestClassService_DeleteClass(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	jsm := new(mock_nats.JetStreamManagement)
	classRepo := new(mock_class_repo.MockClassRepo)

	classService := &ClassService{
		DB:        db,
		JSM:       jsm,
		ClassRepo: classRepo,
		ClassCommandHandler: commands.ClassCommandHandler{
			DB:        db,
			ClassRepo: classRepo,
		},
		ClassQueryHandler: queries.ClassQueryHandler{
			DB:        db,
			ClassRepo: classRepo,
		},
	}
	classID := "class-id"

	testCases := []struct {
		name         string
		req          *mpb.DeleteClassRequest
		expectedResp *mpb.DeleteClassResponse
		expectedErr  error
		setup        func(ctx context.Context)
	}{
		{
			name: "error with empty class id",
			req: &mpb.DeleteClassRequest{
				ClassId: "",
			},
			expectedResp: &mpb.DeleteClassResponse{},
			expectedErr:  status.Error(codes.InvalidArgument, "`class_id` cannot be empty"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "error when delete",
			req: &mpb.DeleteClassRequest{
				ClassId: classID,
			},
			expectedResp: &mpb.DeleteClassResponse{},
			expectedErr:  status.Error(codes.Internal, pgx.ErrTxCommitRollback.Error()),
			setup: func(ctx context.Context) {
				classRepo.On("Delete", ctx, db, classID).Once().Return(pgx.ErrTxCommitRollback)
			},
		},
		{
			name: "success",
			req: &mpb.DeleteClassRequest{
				ClassId: classID,
			},
			expectedResp: &mpb.DeleteClassResponse{},
			setup: func(ctx context.Context) {
				classRepo.On("Delete", ctx, db, classID).Once().Return(nil)
				jsm.On("PublishAsyncContext", ctx, constants.SubjectMasterMgmtClassUpserted, mock.Anything).Once().Return("", nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			gotResp, err := classService.DeleteClass(ctx, tc.req)
			if err != nil {
				assert.Error(t, err)
				assert.Equal(t, err, tc.expectedErr)
				assert.EqualValues(t, gotResp, tc.expectedResp)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, db, classRepo)
		})
	}
}

func TestLocationReaderService_RetrieveClassesByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	db := new(mock_database.Ext)
	classRepo := new(mock_class_repo.MockClassRepo)
	service := &ClassService{
		DB:        db,
		ClassRepo: classRepo,
		ClassCommandHandler: commands.ClassCommandHandler{
			DB:        db,
			ClassRepo: classRepo,
		},
		ClassQueryHandler: queries.ClassQueryHandler{
			DB:        db,
			ClassRepo: classRepo,
		},
	}
	classIds := []string{"class-id-1", "class-id-2"}
	t.Run("success", func(t *testing.T) {

		classes := []*class_domain.Class{
			{ClassID: classIds[0], Name: "class-1", LocationID: "location-1"},
			{ClassID: classIds[1], Name: "class-2", LocationID: "location-2"},
		}
		classRepo.On("RetrieveByIDs", mock.Anything, db, classIds).Return(classes, nil).Once()
		res, err := service.RetrieveClassesByIDs(ctx, &mpb.RetrieveClassByIDsRequest{
			ClassIds: classIds,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, classIds, len(res.GetClasses()))
		for i, lt := range res.GetClasses() {
			assert.NotEmpty(t, lt.ClassId)
			assert.NotEmpty(t, lt.Name)
			assert.Equal(t, classes[i].LocationID, lt.LocationId)
		}
		classRepo.AssertExpectations(t)
	})
	t.Run("error", func(t *testing.T) {
		classRepo.On("RetrieveByIDs", mock.Anything, db, classIds).Return(nil, errors.New("Internal Error")).Once()
		res, err := service.RetrieveClassesByIDs(ctx, &mpb.RetrieveClassByIDsRequest{
			ClassIds: classIds,
		})
		assert.Error(t, err)
		assert.Nil(t, res)
		classRepo.AssertExpectations(t)
	})
}
