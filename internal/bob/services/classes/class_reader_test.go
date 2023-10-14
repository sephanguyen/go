package classes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_services "github.com/manabie-com/backend/mock/eureka/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_lesson_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_master "github.com/manabie-com/backend/mock/mastermgmt/modules/class/infrastructure/repo"
	pb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestClassReaderService_RetrieveClassByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	classRepo := new(mock_repositories.MockClassRepo)
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	masterClassRepo := new(mock_master.MockClassRepo)
	mockDB := &mock_database.Ext{}

	c := &ClassReaderService{
		DB:               mockDB,
		ClassRepo:        classRepo,
		Env:              "local",
		UnleashClientIns: mockUnleashClient,
		MasterClassRepo:  masterClassRepo,
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	class := &entities.Class{
		ID:   database.Int4(1),
		Name: database.Text("class-1"),
	}
	classes := []*entities.Class{class}
	testCases := map[string]TestCase{
		"empty request": {
			ctx:          ctx,
			req:          &pb.RetrieveClassByIDsRequest{},
			expectedResp: &pb.RetrieveClassByIDsResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
			},
		},
		"err find class by ids": {
			ctx:          ctx,
			req:          &pb.RetrieveClassByIDsRequest{ClassIds: []string{"1", "2"}},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("s.ClassRepo.FindByIDs: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().
					Return(false, nil)
				classRepo.On("FindByIDs", ctx, mockDB, database.Int4Array([]int32{1, 2})).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"success": {
			ctx:          ctx,
			req:          &pb.RetrieveClassByIDsRequest{ClassIds: []string{"1"}},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().
					Return(false, nil)
				classRepo.On("FindByIDs", ctx, mockDB, database.Int4Array([]int32{1})).Once().Return(classes, nil)
			},
		},
		"success with enable unleash": {
			ctx:          ctx,
			req:          &pb.RetrieveClassByIDsRequest{ClassIds: []string{"1"}},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().
					Return(true, nil)
				masterClassRepo.On("RetrieveByIDs", ctx, c.DB, []string{"1"}).Once().Return([]*domain.Class{
					{
						ClassID:  "1",
						Name:     "class-1",
						SchoolID: "1",
					},
				}, nil)
			},
		},
		"error find class by ids with enable unleash": {
			ctx:          ctx,
			req:          &pb.RetrieveClassByIDsRequest{ClassIds: []string{"1"}},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("s.MasterClassRepo.FindByIDs: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().
					Return(true, nil)
				masterClassRepo.On("RetrieveByIDs", ctx, c.DB, []string{"1"}).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := c.RetrieveClassByIDs(testCase.ctx, testCase.req.(*pb.RetrieveClassByIDsRequest))
			if testCase.expectedErr == nil {
				assert.NoError(t, err, "expecting no error")
			} else {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
			}

			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestListStudentsByLesson(t *testing.T) {
	t.Parallel()

	now := time.Now()
	timeProto := timestamppb.New(now)

	t.Run("lesson doesn't have students", func(t *testing.T) {
		t.Parallel()

		lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
		lessonRepo := new(mock_lesson_repositories.MockLessonRepo)
		studentEnrollmentHistoryRepo := new(mock_repositories.MockStudentEnrolledHistoryRepo)
		lessonMemberRepo.On("ListStudentsByLessonID", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.User{}, nil)
		lessonRepo.On("GetLessonByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(&lesson_domain.Lesson{
			LocationID: "center-id-1",
		}, nil)
		svc := &ClassReaderService{
			LessonMemberRepo:         lessonMemberRepo,
			LessonRepo:               lessonRepo,
			StudentEnrollmentHistory: studentEnrollmentHistoryRepo,
		}

		resp, err := svc.ListStudentsByLesson(context.Background(), &pb.ListStudentsByLessonRequest{})
		mock.AssertExpectationsForObjects(t, lessonMemberRepo)
		assert.Nil(t, err)
		assert.Equal(t, &pb.ListStudentsByLessonResponse{}, resp)
	})

	t.Run("lesson has valid students", func(t *testing.T) {
		t.Parallel()
		req := &pb.ListStudentsByLessonRequest{
			LessonId: "lid",
			Paging: &cpb.Paging{
				Limit: 99,
				Offset: &cpb.Paging_OffsetMultipleCombined{
					OffsetMultipleCombined: &cpb.Paging_MultipleCombined{
						Combined: []*cpb.Paging_Combined{
							{
								OffsetString: "name",
							},
							{
								OffsetString: "id",
							},
						},
					},
				},
			},
		}
		startDate := time.Now()
		endDate := startDate.AddDate(0, 0, 1)
		lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
		lessonRepo := new(mock_lesson_repositories.MockLessonRepo)
		studentEnrollmentHistoryRepo := new(mock_repositories.MockStudentEnrolledHistoryRepo)
		lessonMemberRepo.On(
			"ListStudentsByLessonID",
			mock.Anything,
			mock.Anything,
			&repositories.ListStudentsByLessonArgs{
				LessonID: database.Text(req.LessonId),
				Limit:    req.Paging.Limit,
				UserName: database.Text("name"),
				UserID:   database.Text("id"),
			},
		).Once().Return(
			[]*entities.User{
				{
					ID:            database.Text("sid1"),
					LastName:      database.Text("name1"),
					LastLoginDate: pgtype.Timestamptz{Time: now, Status: pgtype.Present},
				},
				{
					ID:       database.Text("sid2"),
					LastName: database.Text("name2"),
				},
				{
					ID:       database.Text("sid3"),
					LastName: database.Text("name3"),
				},
			},
			nil,
		)
		lessonRepo.On("GetLessonByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(&lesson_domain.Lesson{
			LocationID: "center-id-1",
		}, nil)
		studentEnrollmentHistoryRepo.On("Retrieve", mock.Anything, mock.Anything, database.TextArray([]string{
			"sid1", "sid2", "sid3",
		}), database.Text("center-id-1")).Once().Return([]*entities.StudentEnrollmentStatusHistory{
			{
				StudentID:  database.Text("sid1"),
				LocationID: database.Text("center-id-1"),
				StartDate:  database.Timestamptz(startDate),
				EndDate:    database.Timestamptz(endDate),
			},
			{
				StudentID:  database.Text("sid2"),
				LocationID: database.Text("center-id-1"),
				StartDate:  database.Timestamptz(startDate),
				EndDate:    database.Timestamptz(endDate),
			},
			{
				StudentID:  database.Text("sid3"),
				LocationID: database.Text("center-id-1"),
				StartDate:  database.Timestamptz(startDate),
				EndDate:    database.Timestamptz(endDate),
			},
		}, nil)

		svc := &ClassReaderService{
			LessonMemberRepo:         lessonMemberRepo,
			LessonRepo:               lessonRepo,
			StudentEnrollmentHistory: studentEnrollmentHistoryRepo,
		}

		resp, err := svc.ListStudentsByLesson(context.Background(), req)
		mock.AssertExpectationsForObjects(t, lessonMemberRepo)
		assert.Nil(t, err)
		expectedStudents := []*cpb.BasicProfile{
			{
				UserId:        "sid1",
				Name:          "name1",
				LastLoginDate: timeProto,
			},
			{
				UserId: "sid2",
				Name:   "name2",
			},
			{
				UserId: "sid3",
				Name:   "name3",
			},
		}
		expectedNextPaging := &cpb.Paging{
			Limit: req.Paging.Limit,
			Offset: &cpb.Paging_OffsetMultipleCombined{
				OffsetMultipleCombined: &cpb.Paging_MultipleCombined{
					Combined: []*cpb.Paging_Combined{
						{
							OffsetString: "name3",
						},
						{
							OffsetString: "sid3",
						},
					},
				},
			},
		}
		enrollmentStatus := []*pb.EnrollmentStatus{
			{
				StudentId: "sid1",
				Info: []*pb.EnrollmentStatus_EnrollmentStatusInfo{
					{
						LocationId: "center-id-1",
						StartDate:  timestamppb.New(startDate),
						EndDate:    timestamppb.New(endDate),
					},
				},
			},
			{
				StudentId: "sid2",
				Info: []*pb.EnrollmentStatus_EnrollmentStatusInfo{
					{
						LocationId: "center-id-1",
						StartDate:  timestamppb.New(startDate),
						EndDate:    timestamppb.New(endDate),
					},
				},
			},
			{
				StudentId: "sid3",
				Info: []*pb.EnrollmentStatus_EnrollmentStatusInfo{
					{
						LocationId: "center-id-1",
						StartDate:  timestamppb.New(startDate),
						EndDate:    timestamppb.New(endDate),
					},
				},
			},
		}
		assert.Equal(t, resp.Students, expectedStudents)
		assert.Equal(t, resp.NextPage, expectedNextPaging)
		assert.ObjectsAreEqual(resp.EnrollmentStatus, enrollmentStatus)
	})
}

func TestRetrieveClassMembers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	classMemberRepo := new(mock_repositories.MockClassMemberRepo)
	mockDB := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	masterClassRepo := new(mock_master.MockClassRepo)
	masterClassMemberRepo := new(mock_master.MockClassMemberRepo)
	userRepo := new(mock_repositories.MockUserRepo)
	s := &ClassReaderService{
		DB:                    mockDB,
		ClassMemberRepo:       classMemberRepo,
		Env:                   "local",
		UnleashClientIns:      mockUnleashClient,
		MasterClassRepo:       masterClassRepo,
		MasterClassMemberRepo: masterClassMemberRepo,
		UserRepo:              userRepo,
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	classMembers := []*entities.ClassMember{}
	nilText := pgtype.Text{Status: pgtype.Null}
	nilInt4 := pgtype.Int4{Status: pgtype.Null}
	now := time.Now()
	pgtypeNow := pgtype.Timestamptz{Time: now}
	timeNow := timestamppb.New(now)
	filter := &repositories.FindClassMemberFilter{
		ClassIDs: database.Int4Array([]int32{1, 2}),
		Status:   database.Text(entities.ClassMemberStatusActive),
		Group:    database.Text(cpb.UserGroup_name[int32(cpb.UserGroup_USER_GROUP_STUDENT)]),
		Limit:    nilInt4,
		OffsetID: nilText,
		UserName: nilText,
	}
	testCases := map[string]TestCase{
		"empty request": {
			ctx:          ctx,
			req:          &pb.RetrieveClassMembersRequest{},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("rpc error: code = InvalidArgument desc = invalid class ids"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().
					Return(false, nil)
				classMemberRepo.On("Find", ctx, mockDB, &repositories.FindClassMemberFilter{ClassIDs: database.Int4Array([]int32{1, 2}), Status: database.Text(entities.ClassMemberStatusActive)}).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"success": {
			ctx:          ctx,
			req:          &pb.RetrieveClassMembersRequest{ClassIds: []string{"1", "2"}, Paging: nil, UserGroup: cpb.UserGroup_USER_GROUP_STUDENT},
			expectedResp: &pb.RetrieveClassMembersResponse{Paging: nil, Members: []*pb.RetrieveClassMembersResponse_Member{}},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().
					Return(false, nil)
				classMemberRepo.On("Find", ctx, mockDB, filter).Once().Return(classMembers, nil)
			},
		},
		"empty request with enable unleash": {
			ctx:          ctx,
			req:          &pb.RetrieveClassMembersRequest{ClassIds: []string{"1", "2"}},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("LessonMemberRepo.RetrieveByClassIDs: no rows in result set"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().
					Return(true, nil)
				masterClassMemberRepo.On("RetrieveByClassIDs", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"success with enable unleash": {
			ctx: ctx,
			req: &pb.RetrieveClassMembersRequest{ClassIds: []string{"1", "2"}, Paging: nil, UserGroup: cpb.UserGroup_USER_GROUP_STUDENT},
			expectedResp: &pb.RetrieveClassMembersResponse{Paging: &cpb.Paging{
				Limit: 0,
				Offset: &cpb.Paging_OffsetMultipleCombined{
					OffsetMultipleCombined: &cpb.Paging_MultipleCombined{
						Combined: []*cpb.Paging_Combined{
							{
								OffsetString: "give last",
							},
							{
								OffsetString: "user-2",
							},
						},
					},
				},
			}, Members: []*pb.RetrieveClassMembersResponse_Member{
				{
					UserId: "user-1",
					JoinAt: timeNow,
				},
				{
					UserId: "user-2",
					JoinAt: timeNow,
				},
			}},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().
					Return(true, nil)
				masterClassMemberRepo.On("RetrieveByClassIDs", ctx, mockDB, mock.Anything).Once().Return([]*domain.ClassMember{
					{
						ClassMemberID: "class-member-1",
						UserID:        "user-1",
						ClassID:       "1",
						CreatedAt:     now,
					},
					{
						ClassMemberID: "class-member-2",
						UserID:        "user-2",
						ClassID:       "2",
						CreatedAt:     now,
					},
				}, nil)
				userRepo.On("Get", ctx, mockDB, mock.Anything).Once().Return(
					&entities.User{ID: database.Text("1"), GivenName: database.Text("give"), LastName: database.Text("last"), CreatedAt: pgtypeNow}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RetrieveClassMembers(testCase.ctx, testCase.req.(*pb.RetrieveClassMembersRequest))
			if testCase.expectedErr == nil {
				assert.NoError(t, err, "expecting no error")
			} else {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
			}

			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestListClass(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	classMemberRepo := new(mock_repositories.MockClassMemberRepo)
	mockDB := &mock_database.Ext{}

	s := &ClassReaderService{
		DB:              mockDB,
		ClassMemberRepo: classMemberRepo,
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	testCases := map[string]TestCase{
		"not implemented": {
			ctx:          ctx,
			req:          &pb.ListClassRequest{},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unimplemented, "not implemented"),
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.ListClass(testCase.ctx, testCase.req.(*pb.ListClassRequest))
			if testCase.expectedErr == nil {
				assert.NoError(t, err, "expecting no error")
			} else {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
			}

			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestRetrieveClassLearningStatistics(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	classMemberRepo := new(mock_repositories.MockClassMemberRepo)
	mockDB := &mock_database.Ext{}

	s := &ClassReaderService{
		DB:              mockDB,
		ClassMemberRepo: classMemberRepo,
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	testCases := map[string]TestCase{
		"not implemented": {
			ctx:          ctx,
			req:          &pb.RetrieveClassLearningStatisticsRequest{},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unimplemented, "not implemented"),
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RetrieveClassLearningStatistics(testCase.ctx, testCase.req.(*pb.RetrieveClassLearningStatisticsRequest))
			if testCase.expectedErr == nil {
				assert.NoError(t, err, "expecting no error")
			} else {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
			}

			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestRetrieveStudentLearningStatistics(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	classMemberRepo := new(mock_repositories.MockClassMemberRepo)
	mockDB := &mock_database.Ext{}

	s := &ClassReaderService{
		DB:              mockDB,
		ClassMemberRepo: classMemberRepo,
	}

	userID := idutil.ULIDNow()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	testCases := map[string]TestCase{
		"not implemented": {
			ctx:          ctx,
			req:          &pb.RetrieveStudentLearningStatisticsRequest{},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Unimplemented, "not implemented"),
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp, err := s.RetrieveStudentLearningStatistics(testCase.ctx, testCase.req.(*pb.RetrieveStudentLearningStatisticsRequest))
			if testCase.expectedErr == nil {
				assert.NoError(t, err, "expecting no error")
			} else {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
			}

			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}

func TestClassReaderService_retrieveAllStudentIDsByCourseID(t *testing.T) {
	t.Parallel()

	courseReadSvc := new(mock_services.MockCourseReaderServiceClient)
	mockDB := &mock_database.Ext{}

	s := &ClassReaderService{
		DB:              mockDB,
		CourseReaderSvc: courseReadSvc,
	}

	ctx := context.Background()
	ctx = interceptors.NewIncomingContext(ctx)
	cctx, _ := interceptors.GetOutgoingContext(ctx)

	// Happy case
	t.Run("happy case", func(t *testing.T) {

		resMock := []*epb.ListStudentIDsByCourseResponse_StudentCourses{}
		for i := 1; i <= 2000; i++ {
			resMock = append(resMock, &epb.ListStudentIDsByCourseResponse_StudentCourses{
				StudentId: fmt.Sprintf("student_%d", i),
			})
		}

		courseReadSvc.On("ListStudentIDsByCourse", cctx, mock.Anything).Once().Return(&epb.ListStudentIDsByCourseResponse{
			StudentCourses: resMock,
			NextPage: &cpb.Paging{
				Limit: 2000,
				Offset: &cpb.Paging_OffsetString{
					OffsetString: "student_2000",
				},
			},
		}, nil)
		courseReadSvc.On("ListStudentIDsByCourse", cctx, mock.Anything).Once().Return(&epb.ListStudentIDsByCourseResponse{
			StudentCourses: []*epb.ListStudentIDsByCourseResponse_StudentCourses{
				{StudentId: "student_2001"},
			},
			NextPage: &cpb.Paging{
				Limit: 2000,
			},
		}, nil)

		out, err := s.retrieveAllStudentIDsByCourseID(ctx, "course_id")

		assert.Nil(t, err)
		assert.NotNil(t, out)
		assert.Len(t, out, 2001)
	})

	// Error case
	t.Run("error case", func(t *testing.T) {
		courseReadSvc.On("ListStudentIDsByCourse", cctx, mock.Anything).Once().Return(nil, fmt.Errorf("something error"))

		_, err := s.retrieveAllStudentIDsByCourseID(ctx, "something")
		assert.NotNil(t, err)
	})
}

func TestClassReaderService_RetrieveClassMembersWithFilters(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var (
		schoolHistoryRepo = new(mock_repositories.MockSchoolHistoryRepo)
		classMemberRepo   = new(mock_repositories.MockClassMemberRepo)
		taggedUserRepo    = new(mock_repositories.MockTaggedUserRepo)
		db                = new(mock_database.Ext)
		courseReadSvc     = new(mock_services.MockCourseReaderServiceClient)
	)
	ctx = interceptors.NewIncomingContext(ctx)
	cctx, _ := interceptors.GetOutgoingContext(ctx)
	ListStudentIDsByCourseResponse_StudentCourses := []*epb.ListStudentIDsByCourseResponse_StudentCourses{}
	ListStudentIDsByCourseResponse_StudentCourses = append(ListStudentIDsByCourseResponse_StudentCourses, &epb.ListStudentIDsByCourseResponse_StudentCourses{
		StudentId: "user-1",
	})
	ListStudentIDsByCourseResponse_StudentCourses = append(ListStudentIDsByCourseResponse_StudentCourses, &epb.ListStudentIDsByCourseResponse_StudentCourses{
		StudentId: "user-2",
	})
	ListStudentIDsByCourseResponse_StudentCourses = append(ListStudentIDsByCourseResponse_StudentCourses, &epb.ListStudentIDsByCourseResponse_StudentCourses{
		StudentId: "user-unassign-school",
	})

	ListStudentIDsByCourseResponse_StudentCoursesAll := []*epb.ListStudentIDsByCourseResponse_StudentCourses{}
	ListStudentIDsByCourseResponse_StudentCoursesAll = append(ListStudentIDsByCourseResponse_StudentCoursesAll, &epb.ListStudentIDsByCourseResponse_StudentCourses{
		StudentId: "user-1",
	})
	ListStudentIDsByCourseResponse_StudentCoursesAll = append(ListStudentIDsByCourseResponse_StudentCoursesAll, &epb.ListStudentIDsByCourseResponse_StudentCourses{
		StudentId: "user-unassign-class",
	})
	ListStudentIDsByCourseResponse_StudentCoursesAll = append(ListStudentIDsByCourseResponse_StudentCoursesAll, &epb.ListStudentIDsByCourseResponse_StudentCourses{
		StudentId: "user-unassign-school",
	})
	s := ClassReaderService{
		DB:                db,
		SchoolHistoryRepo: schoolHistoryRepo,
		ClassMemberRepo:   classMemberRepo,
		TaggedUserRepo:    taggedUserRepo,
		CourseReaderSvc:   courseReadSvc,
	}
	schoolHistoryE := []*entities.SchoolHistory{}

	schoolHistoryE = append(schoolHistoryE, &entities.SchoolHistory{
		StudentID: database.Text("user-1"),
	})
	schoolHistoryE = append(schoolHistoryE, &entities.SchoolHistory{
		StudentID: database.Text("user-unassign-class"),
	})
	schoolHistoryE = append(schoolHistoryE, &entities.SchoolHistory{
		StudentID: database.Text("user-unassign-school"),
	})

	schoolHistoryAll := []*entities.SchoolHistory{
		&entities.SchoolHistory{
			StudentID: database.Text("user-1"),
		},
		&entities.SchoolHistory{
			StudentID: database.Text("user-unassign-class"),
		},
		&entities.SchoolHistory{
			StudentID: database.Text("user-unassign-school"),
		},
	}

	schoolHistoryAssign := []*entities.SchoolHistory{}
	schoolHistoryAssign = append(schoolHistoryAssign, &entities.SchoolHistory{
		StudentID: database.Text("user-1"),
	})
	schoolHistoryAssign = append(schoolHistoryAssign, &entities.SchoolHistory{
		StudentID: database.Text("user-2"),
	})
	classMemberAssign := []*entities.ClassMemberV2{}
	classMemberAssign = append(classMemberAssign, &entities.ClassMemberV2{
		UserID: database.Text("user-2"),
	})
	classMemberAssignAndUnassign := []*entities.ClassMemberV2{}
	classMemberAssignAndUnassign = append(classMemberAssignAndUnassign, &entities.ClassMemberV2{
		UserID: database.Text("user-2"),
	})
	classMemberAssignAndUnassign = append(classMemberAssignAndUnassign, &entities.ClassMemberV2{
		UserID: database.Text("user-unassign-class"),
	})
	testCases := []TestCase{
		{
			name: "happy case get with school id",
			ctx:  ctx,
			req: &pb.RetrieveClassMembersWithFiltersRequest{
				CourseId: "course-1",
				School: &pb.RetrieveClassMembersWithFiltersRequest_SchoolId{
					SchoolId: "school-1",
				},
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			expectedErr: nil,
			expectedResp: &pb.RetrieveClassMembersWithFiltersResponse{
				UserIds: []string{"user-1", "user-unassign-class", "user-unassign-school"},
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			setup: func(ctx context.Context) {
				courseReadSvc.On("ListStudentIDsByCourse", cctx, mock.Anything).Once().Return(&epb.ListStudentIDsByCourseResponse{
					StudentCourses: []*epb.ListStudentIDsByCourseResponse_StudentCourses{
						{StudentId: "user-1"},
					},
					NextPage: &cpb.Paging{
						Limit: 2000,
					},
				}, nil)
				schoolHistoryRepo.On("FindBySchoolAndStudentIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(schoolHistoryE, nil).Once()
			},
		},
		{
			name: "happy case get with unassign school id",
			ctx:  ctx,
			req: &pb.RetrieveClassMembersWithFiltersRequest{
				CourseId: "course-1",
				School: &pb.RetrieveClassMembersWithFiltersRequest_Unassigned{
					Unassigned: true,
				},
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			expectedErr: nil,
			expectedResp: &pb.RetrieveClassMembersWithFiltersResponse{
				UserIds: []string{"user-unassign-school"},
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			setup: func(ctx context.Context) {
				courseReadSvc.On("ListStudentIDsByCourse", cctx, mock.Anything).Once().Return(&epb.ListStudentIDsByCourseResponse{
					StudentCourses: ListStudentIDsByCourseResponse_StudentCourses,
					NextPage: &cpb.Paging{
						Limit: 2000,
					},
				}, nil)
				schoolHistoryRepo.On("FindByStudentIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(schoolHistoryAssign, nil).Once()
			},
		},
		{
			name: "happy case filter empty school id",
			ctx:  ctx,
			req: &pb.RetrieveClassMembersWithFiltersRequest{
				CourseId: "course-1",
				School: &pb.RetrieveClassMembersWithFiltersRequest_SchoolId{
					SchoolId: "school-1",
				},
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			expectedErr: nil,
			expectedResp: &pb.RetrieveClassMembersWithFiltersResponse{
				UserIds: []string{},
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			setup: func(ctx context.Context) {
				courseReadSvc.On("ListStudentIDsByCourse", cctx, mock.Anything).Once().Return(&epb.ListStudentIDsByCourseResponse{
					StudentCourses: []*epb.ListStudentIDsByCourseResponse_StudentCourses{
						{StudentId: "user-1"},
					},
					NextPage: &cpb.Paging{
						Limit: 2000,
					},
				}, nil)
				schoolHistoryRepo.On("FindBySchoolAndStudentIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
			},
		},
		{
			name: "happy case get unassign class id",
			ctx:  ctx,
			req: &pb.RetrieveClassMembersWithFiltersRequest{
				CourseId: "course-1",
				School: &pb.RetrieveClassMembersWithFiltersRequest_AllSchool{
					AllSchool: true,
				},
				ClassIds: []string{"UNASSIGN_CLASS_ID"},
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			expectedErr: nil,
			expectedResp: &pb.RetrieveClassMembersWithFiltersResponse{
				UserIds: []string{"user-unassign-class"},
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			setup: func(ctx context.Context) {
				courseReadSvc.On("ListStudentIDsByCourse", cctx, mock.Anything).Once().Return(&epb.ListStudentIDsByCourseResponse{
					StudentCourses: ListStudentIDsByCourseResponse_StudentCoursesAll,
					NextPage: &cpb.Paging{
						Limit: 2000,
					},
				}, nil)
				schoolHistoryRepo.On("FindByStudentIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(schoolHistoryAll, nil).Once()
				classMemberRepo.On("FindByClassIDsAndUserIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, nil).Once()
				classMemberRepo.On("FindByUserIDsAndCourseIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return([]*entities.ClassMemberV2{&entities.ClassMemberV2{UserID: database.Text("user-1")}, &entities.ClassMemberV2{UserID: database.Text("user-unassign-school")}}, nil).Once()
			},
		},
		{
			name: "happy case get class id",
			ctx:  ctx,
			req: &pb.RetrieveClassMembersWithFiltersRequest{
				CourseId: "course-1",
				School: &pb.RetrieveClassMembersWithFiltersRequest_Unassigned{
					Unassigned: true,
				},
				ClassIds: []string{"class-1", "class-2"},
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			expectedErr: nil,
			expectedResp: &pb.RetrieveClassMembersWithFiltersResponse{
				UserIds: []string{"user-2"},
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			setup: func(ctx context.Context) {
				courseReadSvc.On("ListStudentIDsByCourse", cctx, mock.Anything).Once().Return(&epb.ListStudentIDsByCourseResponse{
					StudentCourses: ListStudentIDsByCourseResponse_StudentCourses,
					NextPage: &cpb.Paging{
						Limit: 2000,
					},
				}, nil)
				schoolHistoryRepo.On("FindByStudentIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(schoolHistoryE, nil).Once()
				classMemberRepo.On("FindByClassIDsAndUserIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(classMemberAssign, nil).Once()
			},
		},

		{
			name: "happy case get all",
			ctx:  ctx,
			req: &pb.RetrieveClassMembersWithFiltersRequest{
				CourseId: "course-1",
				School: &pb.RetrieveClassMembersWithFiltersRequest_AllSchool{
					AllSchool: true,
				},
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			expectedErr: nil,
			expectedResp: &pb.RetrieveClassMembersWithFiltersResponse{
				UserIds: []string{"user-1"},
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			setup: func(ctx context.Context) {
				courseReadSvc.On("ListStudentIDsByCourse", cctx, mock.Anything).Once().Return(&epb.ListStudentIDsByCourseResponse{
					StudentCourses: []*epb.ListStudentIDsByCourseResponse_StudentCourses{
						{StudentId: "user-1"},
					},
					NextPage: &cpb.Paging{
						Limit: 2000,
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase.setup(testCase.ctx)
		req := testCase.req.(*pb.RetrieveClassMembersWithFiltersRequest)
		resp, err := s.RetrieveClassMembersWithFilters(testCase.ctx, req)
		assert.Equal(t, testCase.expectedErr, err)
		assert.Equal(t, testCase.expectedResp, resp)
	}
}
