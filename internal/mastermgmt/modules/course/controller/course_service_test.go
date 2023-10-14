package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/application/queries"
	course_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure/repo"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_course_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/course/infrastructure/repo"
	mock_location "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name             string
	ctx              context.Context
	req              interface{}
	expectedResp     interface{}
	expectedErr      error
	setup            func(ctx context.Context)
	expectedErrModel *errdetails.BadRequest
}

func TestMasterDataCourseService_UpsertCourses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	courseAccessPathRepo := new(mock_course_repo.MockCourseAccessPathRepo)
	studentSubscriptionRepo := new(mock_course_repo.MockStudentSubscriptionRepo)
	locationRepo := new(mock_location.MockLocationRepo)
	masterMgmtCourseRepo := new(mock_course_repo.MockCourseRepo)
	masterMgmtCourseTypeRepo := new(mock_course_repo.MockCourseTypeRepo)
	loc1 := "loc-1"
	loc2 := "loc-2"
	name1 := "name-1"
	name2 := "name-2"
	course1 := "course-1"
	ctype1 := "course-type-1"
	cname := "course-name"
	s := &MasterDataCourseService{
		DB:             db,
		CourseTypeRepo: masterMgmtCourseTypeRepo,
		LocationRepo:   locationRepo,
		StudentSubscriptionCommandHandler: queries.StudentSubscriptionQueryHandler{
			DB:                      db,
			StudentSubscriptionRepo: studentSubscriptionRepo,
		},
		CourseCommandHandler: commands.CourseCommandHandler{
			DB:                   db,
			CourseRepo:           masterMgmtCourseRepo,
			CourseAccessPathRepo: courseAccessPathRepo,
		},
	}
	var locationIDs pgtype.TextArray
	result := make(map[string][]string)
	locationIDs.Set([]string{loc1, loc2})
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.UpsertCoursesRequest{Courses: []*bpb.UpsertCoursesRequest_Course{
				{
					Id:          course1,
					Name:        cname,
					LocationIds: []string{loc1, loc2},
					SchoolId:    constant.ManabieSchool,
					CourseType:  ctype1,
					SubjectIds:  []string{"subject_1", "subject_2"},
				},
			}},
			expectedErr: nil,
			expectedResp: &bpb.UpsertCoursesResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				studentSubscriptionRepo.On("GetLocationActiveStudentSubscriptions", ctx, mock.Anything, []string{course1}).Once().Return(
					result, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, database.TextArray([]string{loc1, loc2}), true).Once().Return(
					[]*domain.Location{
						{
							LocationID: loc1,
							Name:       name1,
						},
						{
							LocationID: loc2,
							Name:       name2,
						},
					}, nil,
				)
				masterMgmtCourseTypeRepo.On("GetByIDs", ctx, db, []string{ctype1}).Once().Return(
					[]*course_domain.CourseType{
						{
							CourseTypeID: ctype1,
							Name:         name1,
						},
					}, nil,
				)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)

				masterMgmtCourseRepo.On("LinkSubjects", ctx, tx, mock.Anything).Once().Return(nil)
				masterMgmtCourseRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, []string{loc1, loc2}).Return(
					[]string{loc1, loc2}, nil).Once()
				courseAccessPathRepo.On("Delete", ctx, db, []string{course1}).Return(nil).Once()
				courseAccessPathRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "upsert course without course type",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.UpsertCoursesRequest{Courses: []*bpb.UpsertCoursesRequest_Course{
				{
					Id:          course1,
					Name:        cname,
					LocationIds: []string{loc1, loc2},
					SchoolId:    constant.ManabieSchool,
					CourseType:  "",
				},
			}},
			expectedErr: nil,
			expectedResp: &bpb.UpsertCoursesResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				studentSubscriptionRepo.On("GetLocationActiveStudentSubscriptions", ctx, mock.Anything, []string{course1}).Once().Return(
					result, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, database.TextArray([]string{loc1, loc2}), true).Once().Return(
					[]*domain.Location{
						{
							LocationID: loc1,
							Name:       name1,
						},
						{
							LocationID: loc2,
							Name:       name2,
						},
					}, nil,
				)

				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)

				masterMgmtCourseRepo.On("LinkSubjects", ctx, tx, mock.Anything).Once().Return(nil)
				masterMgmtCourseRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, []string{loc1, loc2}).Return(
					[]string{loc1, loc2}, nil).Once()
				courseAccessPathRepo.On("Delete", ctx, db, []string{course1}).Return(nil).Once()
				courseAccessPathRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "update with removed location have active subscriptions",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.UpsertCoursesRequest{Courses: []*bpb.UpsertCoursesRequest_Course{
				{
					Id:          course1,
					Name:        cname,
					LocationIds: []string{loc1},
					SchoolId:    constant.ManabieSchool,
					CourseType:  ctype1,
				},
			}},
			expectedErr: status.Error(codes.AlreadyExists, "ra.manabie-error.already_exists"),
			setup: func(ctx context.Context) {
				studentSubscriptionRepo.On("GetLocationActiveStudentSubscriptions", ctx, mock.Anything, []string{course1}).Once().Return(
					map[string][]string{
						course1: {loc2},
					}, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, database.TextArray([]string{loc1}), true).Once().Return(
					[]*domain.Location{
						{
							LocationID: loc1,
							Name:       name1,
						},
					}, nil,
				)
				masterMgmtCourseTypeRepo.On("GetByIDs", ctx, db, []string{ctype1}).Once().Return(
					[]*course_domain.CourseType{
						{
							CourseTypeID: ctype1,
							Name:         name1,
						},
					}, nil,
				)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Commit", mock.Anything).Return(nil)

				masterMgmtCourseRepo.On("LinkSubjects", ctx, tx, mock.Anything).Once().Return(nil)
				masterMgmtCourseRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, []string{loc1, loc2}).Return(
					[]string{loc1, loc2}, nil).Once()
				courseAccessPathRepo.On("Delete", ctx, db, []string{course1}).Return(nil).Once()
				courseAccessPathRepo.On("Upsert", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "update with removed location without active subscriptions",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.UpsertCoursesRequest{Courses: []*bpb.UpsertCoursesRequest_Course{
				{
					Id:          course1,
					Name:        cname,
					LocationIds: []string{loc1},
					SchoolId:    constant.ManabieSchool,
					Grade:       "Grade 12",
					CourseType:  ctype1,
				},
			}},
			expectedErr: nil,
			expectedResp: &bpb.UpsertCoursesResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {
				studentSubscriptionRepo.On("GetLocationActiveStudentSubscriptions", ctx, mock.Anything, []string{course1}).Once().Return(
					result, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, database.TextArray([]string{loc1}), true).Once().Return(
					[]*domain.Location{
						{
							LocationID: loc1,
							Name:       name1,
						},
					}, nil,
				)
				masterMgmtCourseTypeRepo.On("GetByIDs", ctx, db, []string{ctype1}).Once().Return(
					[]*course_domain.CourseType{
						{
							CourseTypeID: ctype1,
							Name:         name1,
						},
					}, nil,
				)
			},
		},
		{
			name: "missing course name",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.UpsertCoursesRequest{Courses: []*bpb.UpsertCoursesRequest_Course{
				{
					Id:          course1,
					LocationIds: []string{loc1, loc2},
					SchoolId:    constant.ManabieSchool,
					Grade:       "Grade 12",
				},
			}},
			expectedErr: status.Error(codes.InvalidArgument, "course name cannot be empty"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "course repo upsert failed",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &bpb.UpsertCoursesRequest{Courses: []*bpb.UpsertCoursesRequest_Course{
				{
					Id:          course1,
					Name:        cname,
					LocationIds: []string{loc1},
					SchoolId:    constant.ManabieSchool,
					CourseType:  ctype1,
				},
			}},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("CourseCommandHandler.UpsertCourses: upsert course failed").Error()),
			setup: func(ctx context.Context) {
				studentSubscriptionRepo.On("GetLocationActiveStudentSubscriptions", ctx, mock.Anything, []string{course1}).Once().Return(
					result, nil)
				locationRepo.On("GetLocationsByLocationIDs", ctx, db, database.TextArray([]string{loc1}), true).Once().Return(
					[]*domain.Location{
						{
							LocationID: loc1,
							Name:       name1,
						},
					}, nil,
				)
				masterMgmtCourseTypeRepo.On("GetByIDs", ctx, db, []string{ctype1}).Once().Return(
					[]*course_domain.CourseType{
						{
							CourseTypeID: ctype1,
							Name:         name1,
						},
					}, nil,
				)
				db.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
				tx.On("Rollback", ctx).Return(nil).Once()
				masterMgmtCourseRepo.On("Upsert", ctx, tx, mock.Anything).Once().Return(errors.New("upsert course failed"))
			},
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
				},
			}
			testCase.ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			testCase.setup(testCase.ctx)
			req := testCase.req.(*bpb.UpsertCoursesRequest)
			resp, err := s.UpsertCourses(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestMasterDataCourseService_GetCoursesByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	now := time.Now().UTC()
	courseRepo := &mock_course_repo.MockCourseRepo{}
	existing := make([]*course_domain.Course, 10)
	notFoundID := idutil.ULIDNow()
	for i := 0; i < 10; i++ {
		id := idutil.ULIDNow()
		existing[i] = &course_domain.Course{
			CourseID:     id,
			Name:         id + "-name",
			CourseTypeID: id + "-type_id",
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	}
	s := &MasterDataCourseService{
		DB: db,
		CourseQueryHandler: queries.CourseQueryHandler{
			DB:         db,
			CourseRepo: courseRepo,
		},
	}

	tc := []TestCase{
		{
			name: "courses found",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetCoursesByIDsRequest{
				CourseIds: []string{existing[0].CourseID},
			},
			expectedErr: nil,
			expectedResp: &mpb.GetCoursesByIDsResponse{
				Courses: []*mpb.Course{
					{
						Id:           existing[0].CourseID,
						Name:         existing[0].Name,
						CourseTypeId: existing[0].CourseTypeID,
					},
				},
			},
			setup: func(ctx context.Context) {
				courseRepo.On("GetByIDs", ctx, db, []string{existing[0].CourseID}).
					Return([]*course_domain.Course{
						existing[0],
					}, nil).Once()
			},
		},
		{
			name: "courses not found",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetCoursesByIDsRequest{
				CourseIds: []string{notFoundID},
			},
			expectedErr: nil,
			expectedResp: &mpb.GetCoursesByIDsResponse{
				Courses: []*mpb.Course{},
			},
			setup: func(ctx context.Context) {
				courseRepo.On("GetByIDs", ctx, db, []string{notFoundID}).
					Return([]*course_domain.Course{}, nil).Once()
			},
		},
		{
			name: "IDs not passed",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetCoursesByIDsRequest{
				CourseIds: []string{},
			},
			expectedErr: nil,
			expectedResp: &mpb.GetCoursesByIDsResponse{
				Courses: []*mpb.Course{},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "internal err",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &mpb.GetCoursesByIDsRequest{CourseIds: []string{existing[3].CourseID}},
			expectedErr:  status.Error(codes.Internal, "internal err"),
			expectedResp: &mpb.GetCoursesByIDsResponse{},
			setup: func(ctx context.Context) {
				courseRepo.On("GetByIDs", ctx, db, []string{existing[3].CourseID}).
					Return([]*course_domain.Course{}, errors.New("internal err")).Once()
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*mpb.GetCoursesByIDsRequest)
			resp, err := s.GetCoursesByIDs(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, db, courseRepo)
		})
	}
}

func TestMasterDataCourseService_ExportCourses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	courseRepo := new(mock_course_repo.MockCourseRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	s := &MasterDataCourseService{
		DB: db,
		CourseQueryHandler: queries.CourseQueryHandler{
			DB:         db,
			CourseRepo: courseRepo,
		},
		UnleashClientIns: mockUnleashClient,
		Env:              "",
	}

	courses := []*repo.Course{
		{
			ID:             database.Text("ID 1"),
			Name:           database.Text("Course 1"),
			CourseTypeID:   database.Text("CID 1"),
			Remarks:        database.Text("Remarks 1"),
			PartnerID:      database.Text("Partner 1"),
			TeachingMethod: database.Text("Group"),
		},
		{
			ID:             database.Text("ID 2"),
			Name:           database.Text("Course 2"),
			CourseTypeID:   database.Text("CID 2"),
			Remarks:        database.Text("Remarks 2"),
			PartnerID:      database.Text("Partner 2"),
			TeachingMethod: database.Text("Individual"),
		},
		{
			ID:             database.Text("ID 3"),
			Name:           database.Text("Course 3"),
			CourseTypeID:   database.Text("CID 3"),
			Remarks:        database.Text("Remarks 3"),
			PartnerID:      database.Text("Partner 3"),
			TeachingMethod: database.Text(""),
		},
		{
			ID:             database.Text("ID 4"),
			Name:           database.Text("Course 4"),
			CourseTypeID:   database.Text("CID 4"),
			Remarks:        database.Text("Remarks 4"),
			PartnerID:      database.Text("Partner 4"),
			TeachingMethod: database.Text(""),
		},
	}

	courseStr := `"course_id","course_name","course_type_id","course_partner_id","teaching_method","remarks"` + "\n" +
		`"ID 1","Course 1","CID 1","Partner 1","Group","Remarks 1"` + "\n" +
		`"ID 2","Course 2","CID 2","Partner 2","Individual","Remarks 2"` + "\n" +
		`"ID 3","Course 3","CID 3","Partner 3","","Remarks 3"` + "\n" +
		`"ID 4","Course 4","CID 4","Partner 4","","Remarks 4"` + "\n"

	courseStrWithoutTeachingMethod := `"course_id","course_name","course_type_id","course_partner_id","remarks"` + "\n" +
		`"ID 1","Course 1","CID 1","Partner 1","Remarks 1"` + "\n" +
		`"ID 2","Course 2","CID 2","Partner 2","Remarks 2"` + "\n" +
		`"ID 3","Course 3","CID 3","Partner 3","Remarks 3"` + "\n" +
		`"ID 4","Course 4","CID 4","Partner 4","Remarks 4"` + "\n"

	t.Run("export all data in db with correct column without teaching method", func(t *testing.T) {
		mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
		s := &MasterDataCourseService{
			DB: db,
			CourseQueryHandler: queries.CourseQueryHandler{
				DB:         db,
				CourseRepo: courseRepo,
			},
			UnleashClientIns: mockUnleashClient,
			Env:              "",
		}

		// arrange
		mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(false, nil)
		courseRepo.On("GetAll", ctx, db).Once().Return(courses, nil)

		byteData := []byte(courseStrWithoutTeachingMethod)

		// act
		resp, err := s.ExportCourses(ctx, &mpb.ExportCoursesRequest{})

		// assert
		assert.Nil(t, err)
		assert.Equal(t, resp.Data, byteData)
	})

	t.Run("export all data in db with correct column", func(t *testing.T) {
		// arrange
		mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(true, nil)
		courseRepo.On("GetAll", ctx, db).Once().Return(courses, nil)

		byteData := []byte(courseStr)

		// act
		resp, err := s.ExportCourses(ctx, &mpb.ExportCoursesRequest{})

		// assert
		assert.Nil(t, err)
		assert.Equal(t, resp.Data, byteData)
	})

	t.Run("return internal error when retrieve data failed", func(t *testing.T) {
		// arrange
		mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(true, nil)
		courseRepo.On("GetAll", ctx, db).Once().Return(nil, errors.New("sample error"))

		// act
		resp, err := s.ExportCourses(ctx, &mpb.ExportCoursesRequest{})

		// assert
		assert.Nil(t, resp.Data)
		assert.Equal(t, err, status.Error(codes.Internal, "sample error"))
	})
}

func TestMasterDataCourseService_ImportCourses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	courseRepo := new(mock_course_repo.MockCourseRepo)
	courseTypeRepo := new(mock_course_repo.MockCourseTypeRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	s := &MasterDataCourseService{
		DB: db,
		CourseQueryHandler: queries.CourseQueryHandler{
			DB:         db,
			CourseRepo: courseRepo,
		},
		CourseTypeRepo:   courseTypeRepo,
		UnleashClientIns: mockUnleashClient,
		Env:              "",
	}
	mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
		Return(false, nil)
	testCases := []TestCase{
		{
			name:        "no data in csv file",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
			req:         &mpb.ImportCoursesRequest{},
		},
		{
			name:        "invalid file - number of column != 5",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "wrong number of columns, expected 5, got 2"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`course_id,course_name
				1,Course 1`),
			},
		},
		{
			name:        "invalid file - first column name (toLowerCase) != course_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 1 should be course_id, got Number"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`Number,course_name,course_type_id,course_partner_id,remarks
				1,Course 1,1,pid,m`),
			},
		},
		{
			name:        "invalid file - second column name (toLowerCase) != course_name",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 2 should be course_name, got name"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`course_id,name,course_type_id,course_partner_id,remarks
				1,Course 1,typeid,pid,m`),
			},
		},
		{
			name:        "invalid file - third column name (toLowerCase) != course_type_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 3 should be course_type_id, got type_id"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`course_id,course_name,type_id,course_partner_id,remarks
				1,Course 1,typeid,pid,m`),
			},
		},
		{
			name:        "invalid file - fifth column name (toLowerCase) != course_partner_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 4 should be course_partner_id, got course_partner_idx"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`course_id,course_name,course_type_id,course_partner_idx,marks
				1,Course 1,typeid,pid,m`),
			},
		},
		{
			name:        "invalid file - sixth column name (toLowerCase) != remarks",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 5 should be remarks, got marks"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`course_id,course_name,course_type_id,course_partner_id,marks
				1,Course 1,typeid,pid,m`),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			resp, err := s.ImportCourses(testCase.ctx, testCase.req.(*mpb.ImportCoursesRequest))
			if testCase.expectedErr != nil {
				assert.Nil(t, resp)
				if testCase.expectedErrModel != nil {
					utils.AssertBadRequestErrorModel(t, testCase.expectedErrModel, err)
				} else {
					assert.Equal(t, err, testCase.expectedErr)
				}
			} else {
				assert.Equal(t, nil, err)
				assert.NotNil(t, resp)
				mock.AssertExpectationsForObjects(t, courseTypeRepo)
				mock.AssertExpectationsForObjects(t, courseRepo)
			}
		})
	}
}

func TestMasterDataCourseService_ImportCoursesV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	courseRepo := new(mock_course_repo.MockCourseRepo)
	courseTypeRepo := new(mock_course_repo.MockCourseTypeRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	s := &MasterDataCourseService{
		DB: db,
		CourseQueryHandler: queries.CourseQueryHandler{
			DB:         db,
			CourseRepo: courseRepo,
		},
		CourseTypeRepo:   courseTypeRepo,
		UnleashClientIns: mockUnleashClient,
		Env:              "",
	}
	mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
		Return(true, nil)
	testCases := []TestCase{
		{
			name:        "no data in csv file",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
			req:         &mpb.ImportCoursesRequest{},
		},
		{
			name:        "invalid file - number of column != 6",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "wrong number of columns, expected 6, got 2"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`course_id,course_name
				1,Course 1`),
			},
		},
		{
			name:        "invalid file - first column name (toLowerCase) != course_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 1 should be course_id, got Number"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`Number,course_name,course_type_id,course_partner_id,teaching_method,remarks
				1,Course 1,1,pid,m,m`),
			},
		},
		{
			name:        "invalid file - second column name (toLowerCase) != course_name",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 2 should be course_name, got name"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`course_id,name,course_type_id,course_partner_id,teaching_method,remarks
				1,Course 1,typeid,pid,m,m`),
			},
		},
		{
			name:        "invalid file - third column name (toLowerCase) != course_type_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 3 should be course_type_id, got type_id"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`course_id,course_name,type_id,course_partner_id,teaching_method,remarks
				1,Course 1,typeid,pid,m,m`),
			},
		},
		{
			name:        "invalid file - fifth column name (toLowerCase) != course_partner_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 4 should be course_partner_id, got course_partner_idx"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`course_id,course_name,course_type_id,course_partner_idx,teaching_method,marks
				1,Course 1,typeid,pid,m,m`),
			},
		},
		{
			name:        "invalid file - sixth column name (toLowerCase) != teaching_method",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 5 should be teaching_method, got teaching"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`course_id,course_name,course_type_id,course_partner_id,teaching,remarks
				1,Course 1,typeid,pid,m,m`),
			},
		},
		{
			name:        "invalid file - seventh column name (toLowerCase) != remarks",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 6 should be remarks, got marks"),
			req: &mpb.ImportCoursesRequest{
				Payload: []byte(`course_id,course_name,course_type_id,course_partner_id,teaching_method,marks
				1,Course 1,typeid,pid,m,m`),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			resp, err := s.ImportCourses(testCase.ctx, testCase.req.(*mpb.ImportCoursesRequest))
			if testCase.expectedErr != nil {
				assert.Nil(t, resp)
				if testCase.expectedErrModel != nil {
					utils.AssertBadRequestErrorModel(t, testCase.expectedErrModel, err)
				} else {
					assert.Equal(t, err, testCase.expectedErr)
				}
			} else {
				assert.Equal(t, nil, err)
				assert.NotNil(t, resp)
				mock.AssertExpectationsForObjects(t, courseTypeRepo)
				mock.AssertExpectationsForObjects(t, courseRepo)
			}
		})
	}
}

func TestMasterDataCourseService_ImportCourses_Business(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	courseRepo := new(mock_course_repo.MockCourseRepo)
	courseTypeRepo := new(mock_course_repo.MockCourseTypeRepo)

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	s := &MasterDataCourseService{
		DB: db,
		CourseQueryHandler: queries.CourseQueryHandler{
			DB:         db,
			CourseRepo: courseRepo,
		},
		CourseCommandHandler: commands.CourseCommandHandler{
			DB:         db,
			CourseRepo: courseRepo,
		},
		CourseTypeRepo:   courseTypeRepo,
		UnleashClientIns: mockUnleashClient,
		Env:              "",
	}

	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: "1",
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)
	tx := &mock_database.Tx{}
	t.Run("valid file happy case", func(t *testing.T) {
		// Arrange
		db.On("Begin", ctx).Once().Return(tx, nil)
		mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(true, nil)

		req := &mpb.ImportCoursesRequest{
			Payload: []byte(`course_id,course_name,course_type_id,course_partner_id,teaching_method,remarks
			1,Course 1,some_typeid,pid,Group,m
			2,Course 2,some_typeid,pid2,Individual,m
			3,Course 3,some_typeid,pid3,,m`),
		}
		courseTypeRepo.On("GetByIDs", ctx, db, []string{"some_typeid", "some_typeid", "some_typeid"}).Once().Return([]*course_domain.CourseType{
			{
				Name:         "Some type ID",
				CourseTypeID: "some_typeid",
			},
		}, nil)

		courseRepo.On("GetByPartnerIDs", ctx, db, []string{"pid", "pid2", "pid3"}).Once().Return([]*course_domain.Course{
			{
				Name:         "Some type ID",
				CourseTypeID: "some_typeid",
				PartnerID:    "pid",
			},
			{
				Name:         "Some type ID 2",
				CourseTypeID: "some_typeid",
				PartnerID:    "pid2",
			},
			{
				Name:         "Some type ID 3",
				CourseTypeID: "some_typeid",
				PartnerID:    "pid3",
			},
		}, nil)
		courseRepo.On("Import", ctx, tx, mock.Anything).Once().Return(nil)
		tx.On("Commit", ctx).Once().Return(nil)
		// Act
		_, err := s.ImportCourses(ctx, req)

		// Assert
		assert.Nil(t, err)
	})

	t.Run("valid file - but course type repo caused error", func(t *testing.T) {
		// Arrange
		mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(true, nil)
		expectedErr := status.Errorf(codes.InvalidArgument, "could not validate course type id, %v", "sample err")
		req := &mpb.ImportCoursesRequest{
			Payload: []byte(`course_id,course_name,course_type_id,course_partner_id,teaching_method,remarks
			1,Course 1,some_typeid,pid,Group,m`),
		}
		courseTypeRepo.On("GetByIDs", ctx, db, []string{"some_typeid"}).Once().Return(nil, fmt.Errorf("%s", "sample err"))

		// Act
		resp, err := s.ImportCourses(ctx, req)

		// Assert
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("parsing valid file with invalid values", func(t *testing.T) {
		// Arrange
		req := &mpb.ImportCoursesRequest{
			Payload: []byte(fmt.Sprintf(`course_id,course_name,course_type_id,course_partner_id,teaching_method,remarks
			1,Course 1,typeid,bool,pid1,Group,m
			99,,typeid,1,pid2,Group,m
			22,Course xyz,typeid2,0,Group,,
			22,Course xyz,typeid2,0,Group1,,
			%s`, fmt.Sprintf("12,%s,typeid,1,pid4,Group,m", string([]byte{0xff, 0xfe, 0xfd})))),
		}
		expectedErrModel := &errdetails.BadRequest{
			FieldViolations: []*errdetails.BadRequest_FieldViolation{
				{
					Field:       "Row Number: 2",
					Description: "bool is not a valid boolean: strconv.ParseBool: parsing \"bool\": invalid syntax",
				},
				{
					Field:       "Row Number: 3",
					Description: "name can not be empty",
				},
				{
					Field:       "Row Number: 5",
					Description: "teachingMethod must be group or individual",
				},
				{
					Field:       "Row Number: 6",
					Description: "name is not a valid UTF8 string",
				},
			},
		}
		mockUnleashClient.On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
			Return(true, nil)
		courseTypeRepo.On("GetByIDs", ctx, db, []string{"typeid", "", "typeid2", ""}).Once().Return([]*course_domain.CourseType{
			{
				Name:         "sample 2",
				CourseTypeID: "typeid2",
			},
			{
				Name:         "sample 1",
				CourseTypeID: "typeid",
			},
		}, nil)
		// Act
		resp, err := s.ImportCourses(ctx, req)

		// Assert
		assert.Nil(t, resp)
		utils.AssertBadRequestErrorModel(t, expectedErrModel, err)
	})
}
