package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockBobUserReaderService struct {
	searchbasicFn func(ctx context.Context, in *bpb.SearchBasicProfileRequest, opts ...grpc.CallOption) (*bpb.SearchBasicProfileResponse, error)
}

type mockClassService struct {
	retrieveClassesByIDs func(ctx context.Context, in *mpb.RetrieveClassByIDsRequest, opts ...grpc.CallOption) (*mpb.RetrieveClassByIDsResponse, error)
}

func (m *mockClassService) RetrieveClassesByIDs(ctx context.Context, in *mpb.RetrieveClassByIDsRequest, opts ...grpc.CallOption) (*mpb.RetrieveClassByIDsResponse, error) {
	return m.retrieveClassesByIDs(ctx, in, opts...)
}

func TestCourseReaderService_ListClassByCourse(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	courseClassRepo := new(mock_repositories.MockCourseClassRepo)
	mockDB := &mock_database.Ext{}
	mockClassService := &mockClassService{}

	s := &CourseReaderService{
		DB:              mockDB,
		CourseClassRepo: courseClassRepo,
		ClassService:    mockClassService,
	}

	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)
	testCases := map[string]TestCase{
		"error find class course": {
			ctx: ctx,
			req: &pb.ListClassByCourseRequest{
				CourseId: "course-id",
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("s.CourseClassRepo.FindClassIDByCourseID: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				courseClassRepo.On("FindClassIDByCourseID", ctx, mockDB, database.Text("course-id")).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		"happy case": {
			ctx: ctx,
			req: &pb.ListClassByCourseRequest{
				CourseId: "course-id",
			},
			expectedResp: &pb.ListClassByCourseResponse{
				ClassIds: []string{"1", "2"},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				courseClassRepo.On("FindClassIDByCourseID", ctx, mockDB, database.Text("course-id")).Once().Return([]string{"1", "2"}, nil)
			},
		},
		"get class with location": {
			ctx: ctx,
			req: &pb.ListClassByCourseRequest{
				CourseId:    "course-id",
				LocationIds: []string{"location-1", "location-2", "location-5"},
			},
			expectedResp: &pb.ListClassByCourseResponse{
				ClassIds: []string{"1", "2"},
			},
			setup: func(ctx context.Context) {
				courseClassRepo.On("FindClassIDByCourseID", ctx, mockDB, database.Text("course-id")).Once().Return([]string{"1", "2", "3"}, nil)
				mockClassService.retrieveClassesByIDs = func(ctx context.Context, in *mpb.RetrieveClassByIDsRequest, opts ...grpc.CallOption) (*mpb.RetrieveClassByIDsResponse, error) {
					assert.ElementsMatch(t, []string{"1", "2", "3"}, in.ClassIds)
					return &mpb.RetrieveClassByIDsResponse{
						Classes: []*mpb.RetrieveClassByIDsResponse_Class{
							{
								ClassId:    "1",
								LocationId: "location-1",
							},
							{
								ClassId:    "2",
								LocationId: "location-2",
							},
							{
								ClassId:    "3",
								LocationId: "location-3",
							},
						},
					}, nil
				}
			},
		},
		"there are no any match location ids": {
			ctx: ctx,
			req: &pb.ListClassByCourseRequest{
				CourseId:    "course-id",
				LocationIds: []string{"location-1", "location-2"},
			},
			expectedResp: &pb.ListClassByCourseResponse{
				ClassIds: []string{},
			},
			setup: func(ctx context.Context) {
				courseClassRepo.On("FindClassIDByCourseID", ctx, mockDB, database.Text("course-id")).Once().Return([]string{"1", "2", "3"}, nil)
				mockClassService.retrieveClassesByIDs = func(ctx context.Context, in *mpb.RetrieveClassByIDsRequest, opts ...grpc.CallOption) (*mpb.RetrieveClassByIDsResponse, error) {
					assert.ElementsMatch(t, []string{"1", "2", "3"}, in.ClassIds)
					return &mpb.RetrieveClassByIDsResponse{
						Classes: []*mpb.RetrieveClassByIDsResponse_Class{
							{
								ClassId:    "1",
								LocationId: "location-3",
							},
							{
								ClassId:    "2",
								LocationId: "location-3",
							},
							{
								ClassId:    "3",
								LocationId: "location-3",
							},
						},
					}, nil
				}
			},
		},
		"not found any class id for course": {
			ctx: ctx,
			req: &pb.ListClassByCourseRequest{
				CourseId:    "course-id",
				LocationIds: []string{"location-1", "location-2"},
			},
			expectedResp: &pb.ListClassByCourseResponse{
				ClassIds: []string{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				courseClassRepo.On("FindClassIDByCourseID", ctx, mockDB, database.Text("course-id")).Once().Return([]string{}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.ctx = interceptors.NewIncomingContext(testCase.ctx)
			testCase.setup(testCase.ctx)
			resp, err := s.ListClassByCourse(testCase.ctx, testCase.req.(*pb.ListClassByCourseRequest))
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

func (m *mockBobUserReaderService) SearchBasicProfile(ctx context.Context, in *bpb.SearchBasicProfileRequest, opts ...grpc.CallOption) (*bpb.SearchBasicProfileResponse, error) {
	return m.searchbasicFn(ctx, in, opts...)
}

func TestCourseReaderService_ListStudentByCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	courseStudentRepo := new(mock_repositories.MockCourseStudentRepo)
	mockDB := &mock_database.Ext{}
	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	t.Run("failed query", func(t *testing.T) {
		s := &CourseReaderService{
			DB:                mockDB,
			CourseStudentRepo: courseStudentRepo,
		}
		testCase := TestCase{
			ctx: ctx,
			req: &pb.ListStudentByCourseRequest{
				CourseId: "course-id",
				Paging:   &cpb.Paging{Limit: 0, Offset: &cpb.Paging_OffsetInteger{}},
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("s.CourseStudentRepo.FindStudentByCourseID: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				courseStudentRepo.On("FindStudentByCourseID", ctx, mockDB, database.Text("course-id")).Once().Return(nil, pgx.ErrNoRows)
			},
		}
		testCase.setup(testCase.ctx)

		resp, err := s.ListStudentByCourse(testCase.ctx, testCase.req.(*pb.ListStudentByCourseRequest))
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
	t.Run("failed call bob `FindStudentByCourseID`", func(t *testing.T) {
		bobService := &mockBobUserReaderService{searchbasicFn: func(ctx context.Context, in *bpb.SearchBasicProfileRequest, opts ...grpc.CallOption) (*bpb.SearchBasicProfileResponse, error) {
			return nil, fmt.Errorf("bob failed")
		}}
		s := &CourseReaderService{
			DB:                mockDB,
			BobUserReader:     bobService,
			CourseStudentRepo: courseStudentRepo,
		}
		testCase := TestCase{
			ctx: ctx,
			req: &pb.ListStudentByCourseRequest{
				CourseId: "course-id",
				Paging:   &cpb.Paging{Limit: 0},
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("bob failed"),
			setup: func(ctx context.Context) {
				courseStudentRepo.On("FindStudentByCourseID", ctx, mockDB, database.Text("course-id")).Once().Return([]string{"1", "2"}, nil)
			},
		}
		testCase.setup(testCase.ctx)
		resp, err := s.ListStudentByCourse(testCase.ctx, testCase.req.(*pb.ListStudentByCourseRequest))
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

	t.Run("happy case`", func(t *testing.T) {
		bobService := &mockBobUserReaderService{searchbasicFn: func(ctx context.Context, in *bpb.SearchBasicProfileRequest, opts ...grpc.CallOption) (*bpb.SearchBasicProfileResponse, error) {
			return &bpb.SearchBasicProfileResponse{Profiles: []*cpb.BasicProfile{{UserId: "usr1"}}, NextPage: &cpb.Paging{}}, nil
		}}
		s := &CourseReaderService{
			DB:                mockDB,
			BobUserReader:     bobService,
			CourseStudentRepo: courseStudentRepo,
		}
		testCase := TestCase{
			ctx: ctx,
			req: &pb.ListStudentByCourseRequest{
				CourseId: "course-id",
				Paging:   &cpb.Paging{Limit: 0},
			},
			expectedResp: &pb.ListStudentByCourseResponse{Profiles: []*cpb.BasicProfile{{UserId: "usr1"}}, NextPage: &cpb.Paging{}},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				courseStudentRepo.On("FindStudentByCourseID", ctx, mockDB, database.Text("course-id")).Once().Return([]string{"1", "2"}, nil)
			},
		}
		testCase.setup(testCase.ctx)
		resp, err := s.ListStudentByCourse(testCase.ctx, testCase.req.(*pb.ListStudentByCourseRequest))
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

func TestCourseReaderService_ListCourseIDsByStudents(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	courseStudentRepo := new(mock_repositories.MockCourseStudentRepo)
	mockDB := &mock_database.Ext{}
	studentIDs := []string{"student-1", "student-2", "student-3"}
	var (
		nilCourses pgtype.TextArray
		nilOffset  pgtype.Text
		nilLimit   pgtype.Int8
	)
	if err := multierr.Combine(
		nilCourses.Set(nil),
		nilOffset.Set(nil),
		nilLimit.Set(nil),
	); err != nil {
		t.Skip("unable to set nil on fields")
	}
	courseStudent := make(map[string][]string)
	courseStudent["student-1"] = []string{"course-1.1", "course-1.2"}
	courseStudent["student-2"] = []string{"course-2.1", "course-2.2"}
	courseStudent["student-3"] = []string{"course-3.1"}

	validResp := &pb.ListCourseIDsByStudentsResponse{
		StudentCourses: []*pb.ListCourseIDsByStudentsResponse_StudentCourses{
			{
				StudentId: "student-1",
				CourseIds: []string{"course-1.1", "course-1.2"},
			},
			{
				StudentId: "student-2",
				CourseIds: []string{"course-2.1", "course-2.2"},
			},
			{
				StudentId: "student-3",
				CourseIds: []string{"course-3.1"},
			},
		},
	}

	t.Run("error when search by students SearchStudents", func(t *testing.T) {
		s := &CourseReaderService{
			DB:                mockDB,
			CourseStudentRepo: courseStudentRepo,
		}
		testCase := TestCase{
			ctx: ctx,
			req: &pb.ListCourseIDsByStudentsRequest{
				StudentIds: studentIDs,
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, fmt.Errorf("CourseStudentRepo.SearchStudents: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				courseStudentRepo.On("SearchStudents", mock.Anything, mock.Anything, &repositories.SearchStudentsFilter{StudentIDs: database.TextArray(studentIDs), CourseIDs: nilCourses, Limit: nilLimit, Offset: nilOffset}).Once().Return(courseStudent, []string{}, pgx.ErrNoRows)
			},
		}
		testCase.setup(testCase.ctx)

		resp, err := s.ListCourseIDsByStudents(ctx, testCase.req.(*pb.ListCourseIDsByStudentsRequest))
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
	t.Run("happy case", func(t *testing.T) {
		s := &CourseReaderService{
			DB:                mockDB,
			CourseStudentRepo: courseStudentRepo,
		}
		testCase := TestCase{
			ctx: ctx,
			req: &pb.ListCourseIDsByStudentsRequest{
				StudentIds: studentIDs,
			},
			expectedResp: validResp,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				courseStudentRepo.On("SearchStudents", mock.Anything, mock.Anything, &repositories.SearchStudentsFilter{StudentIDs: database.TextArray(studentIDs), CourseIDs: nilCourses, Limit: nilLimit, Offset: nilOffset}).Once().Return(courseStudent, []string{}, nil)
			},
		}
		testCase.setup(testCase.ctx)

		resp, err := s.ListCourseIDsByStudents(ctx, testCase.req.(*pb.ListCourseIDsByStudentsRequest))
		if testCase.expectedErr == nil {
			assert.NoError(t, err, "expecting no error")
		} else {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error(), "unexpected error message")
		}

		if testCase.expectedResp == nil {
			assert.Nil(t, testCase.expectedResp, resp)
		} else {
			counter := 0
			expectedResp := testCase.expectedResp.(*pb.ListCourseIDsByStudentsResponse)
			assert.Equal(t, len(expectedResp.GetStudentCourses()), len(resp.GetStudentCourses()))
			for _, cs1 := range expectedResp.GetStudentCourses() {
				for _, cs2 := range resp.GetStudentCourses() {
					if cs1.StudentId == cs2.StudentId && len(cs1.CourseIds) == len(cs2.CourseIds) {
						counter++
					}
				}
			}
			assert.Equal(t, len(expectedResp.GetStudentCourses()), counter, "some course student are not expected")
		}
	})
}

func TestCourseReaderService_ListStudentIDsByCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	courseStudentRepo := new(mock_repositories.MockCourseStudentRepo)
	mockDB := &mock_database.Ext{}
	userID := ksuid.New().String()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	t.Run("fail searchStudent", func(t *testing.T) {
		s := &CourseReaderService{
			DB:                mockDB,
			CourseStudentRepo: courseStudentRepo,
		}
		testCase := TestCase{
			ctx: ctx,
			req: &pb.ListStudentIDsByCourseRequest{
				CourseIds: []string{},
			},
			expectedResp: nil,
			expectedErr:  status.Error(codes.Internal, pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				courseStudentRepo.On("SearchStudents", ctx, mockDB, mock.Anything).Once().Return(map[string][]string{}, []string{}, pgx.ErrNoRows)
			},
		}
		testCase.setup(testCase.ctx)

		resp, err := s.ListStudentIDsByCourse(testCase.ctx, testCase.req.(*pb.ListStudentIDsByCourseRequest))
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

	t.Run("happy case", func(t *testing.T) {
		s := &CourseReaderService{
			DB:                mockDB,
			CourseStudentRepo: courseStudentRepo,
		}
		testCase := TestCase{
			ctx: ctx,
			req: &pb.ListStudentIDsByCourseRequest{
				CourseIds: []string{},
			},
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				courseStudentRepo.On("SearchStudents", ctx, mockDB, mock.Anything).Once().Return(map[string][]string{}, []string{}, nil)
			},
		}
		testCase.setup(testCase.ctx)

		resp, err := s.ListStudentIDsByCourse(testCase.ctx, testCase.req.(*pb.ListStudentIDsByCourseRequest))
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

func TestCourseReaderService_RetrieveCourseStatistic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockCourseStudentRepo := &mock_repositories.MockCourseStudyPlanRepo{}
	mockAssignmentRepo := &mock_repositories.MockAssignmentRepo{}
	mockShuffleQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}
	svc := CourseReaderService{
		CourseStudyPlanRepo: mockCourseStudentRepo,
		AssignmentRepo:      mockAssignmentRepo,
		ShuffledQuizSetRepo: mockShuffleQuizSetRepo,
	}

	cases := []TestCase{
		{
			name: "invalid request mission course id",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequest{
				CourseId:    "",
				StudyPlanId: "StudyPlanId",
				ClassId:     "ClassId",
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.InvalidArgument, "Missing course"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "invalid request mission study plan",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequest{
				CourseId:    "CourseId",
				StudyPlanId: "",
				ClassId:     "ClassId",
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.InvalidArgument, "Missing study plan"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "list course statistic error",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequest{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     "ClassId",
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, "CourseStudyPlanRepo.ListCourseStatisticItems %v", pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItems", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "calculate assignment score error",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequest{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     "ClassId",
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, "AssignmentRepo.CalculateHigestScore %v", pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItems", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItem{}, nil)
				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "calculate task assignment score error",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequest{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     "ClassId",
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, "AssignmentRepo.CalculateTaskAssignmentHighestScore %v", pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItems", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItem{}, nil)
				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{}, nil)
				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "calculate shuffle quiz set score error",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequest{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     "ClassId",
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, "ShuffledQuizSetRepo.CalculateHighestSubmissionScore %v", pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItems", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItem{}, nil)
				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{}, nil)
				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{}, nil)
				mockShuffleQuizSetRepo.On("CalculateHighestSubmissionScore", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "correct score all active item",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequest{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     "ClassId",
			},
			expectedResp: &pb.RetrieveCourseStatisticResponse{
				CourseStatisticItems: []*pb.RetrieveCourseStatisticResponse_CourseStatisticItem{
					{
						TopicId:              "topic-1",
						CompletedStudent:     2,
						TotalAssignedStudent: 3,
						AverageScore:         28,
						StudyPlanItemStatisticItems: []*pb.RetrieveCourseStatisticResponse_CourseStatisticItem_StudyPlanItemStatisticItem{
							{
								StudyPlanItemId:      "root-1",
								CompletedStudent:     2,
								TotalAssignedStudent: 2,
								AverageScore:         38,
							},
							{
								StudyPlanItemId:      "root-2",
								CompletedStudent:     2,
								TotalAssignedStudent: 3,
								AverageScore:         17,
							},
						},
					},
					{
						TopicId:              "topic-2",
						CompletedStudent:     1,
						TotalAssignedStudent: 1,
						AverageScore:         91,
						StudyPlanItemStatisticItems: []*pb.RetrieveCourseStatisticResponse_CourseStatisticItem_StudyPlanItemStatisticItem{
							{
								StudyPlanItemId:      "root-3",
								CompletedStudent:     1,
								TotalAssignedStudent: 1,
								AverageScore:         91,
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItems", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItem{
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-1",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-2",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-3",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-4",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-3",
						},
						StudentID:           "student-3",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-6",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-2",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-3",
						StudyPlanItemID:     "spi-5",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
				}, nil)

				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-1", "spi-2", "spi-5"})).
					Return([]*repositories.CalculateHighestScoreResponse{
						{
							StudyPlanItemID: database.Text("spi-1"),
							Percentage:      database.Float4(31),
						},
						{
							StudyPlanItemID: database.Text("spi-2"),
							Percentage:      database.Float4(44),
						},
					}, nil)

				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-1", "spi-2", "spi-5"})).Once().
					Return([]*repositories.CalculateHighestScoreResponse{
						{
							StudyPlanItemID: database.Text("spi-5"),
							Percentage:      database.Float4(91),
						},
					}, nil)

				mockShuffleQuizSetRepo.On("CalculateHighestSubmissionScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-3", "spi-4"})).Once().
					Return([]*repositories.CalculateHighestScoreResponse{
						{

							StudyPlanItemID: database.Text("spi-3"),
							Percentage:      database.Float4(15),
						},
						{

							StudyPlanItemID: database.Text("spi-4"),
							Percentage:      database.Float4(19),
						},
					}, nil)
			},
		},
		{
			name: "archived items are ignored",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequest{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     "ClassId",
			},
			expectedResp: &pb.RetrieveCourseStatisticResponse{
				CourseStatisticItems: []*pb.RetrieveCourseStatisticResponse_CourseStatisticItem{
					{
						TopicId:              "topic-1",
						CompletedStudent:     2,
						TotalAssignedStudent: 3,
						AverageScore:         30,
						StudyPlanItemStatisticItems: []*pb.RetrieveCourseStatisticResponse_CourseStatisticItem_StudyPlanItemStatisticItem{
							{
								StudyPlanItemId:      "root-1",
								CompletedStudent:     1,
								TotalAssignedStudent: 1,
								AverageScore:         44,
							},
							{
								StudyPlanItemId:      "root-2",
								CompletedStudent:     1,
								TotalAssignedStudent: 2,
								AverageScore:         15,
							},
						},
					},
					{
						TopicId:      "topic-2",
						AverageScore: -1,
						StudyPlanItemStatisticItems: []*pb.RetrieveCourseStatisticResponse_CourseStatisticItem_StudyPlanItemStatisticItem{
							{
								StudyPlanItemId: "root-3",
								AverageScore:    -1,
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItems", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItem{
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-1",
						Status:              "STUDY_PLAN_ITEM_STATUS_ARCHIVED",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-2",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-3",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-4",
						Status:              "STUDY_PLAN_ITEM_STATUS_ARCHIVED",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-3",
						},
						StudentID:           "student-3",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-6",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-2",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-3",
						StudyPlanItemID:     "spi-5",
						Status:              "STUDY_PLAN_ITEM_STATUS_ARCHIVED",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
				}, nil)

				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-2"})).
					Return([]*repositories.CalculateHighestScoreResponse{
						{
							StudyPlanItemID: database.Text("spi-2"),
							Percentage:      database.Float4(44),
						},
					}, nil)

				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-2"})).Once().
					Return([]*repositories.CalculateHighestScoreResponse{}, nil)

				mockShuffleQuizSetRepo.On("CalculateHighestSubmissionScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-3"})).Once().
					Return([]*repositories.CalculateHighestScoreResponse{
						{
							StudyPlanItemID: database.Text("spi-3"),
							Percentage:      database.Float4(15),
						},
					}, nil)
			},
		},
		{
			name: "incomplete items doesn't include in average score",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequest{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     "ClassId",
			},
			expectedResp: &pb.RetrieveCourseStatisticResponse{
				CourseStatisticItems: []*pb.RetrieveCourseStatisticResponse_CourseStatisticItem{
					{
						TopicId:              "topic-1",
						CompletedStudent:     0,
						TotalAssignedStudent: 3,
						AverageScore:         -1,
						StudyPlanItemStatisticItems: []*pb.RetrieveCourseStatisticResponse_CourseStatisticItem_StudyPlanItemStatisticItem{
							{
								StudyPlanItemId:      "root-1",
								CompletedStudent:     0,
								TotalAssignedStudent: 1,
								AverageScore:         -1,
							},
							{
								StudyPlanItemId:      "root-2",
								CompletedStudent:     0,
								TotalAssignedStudent: 2,
								AverageScore:         -1,
							},
						},
					},
					{
						TopicId:              "topic-2",
						CompletedStudent:     1,
						TotalAssignedStudent: 1,
						AverageScore:         44,
						StudyPlanItemStatisticItems: []*pb.RetrieveCourseStatisticResponse_CourseStatisticItem_StudyPlanItemStatisticItem{
							{
								StudyPlanItemId:      "root-3",
								CompletedStudent:     1,
								TotalAssignedStudent: 1,
								AverageScore:         44,
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItems", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItem{
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-1",
						Status:              "STUDY_PLAN_ITEM_STATUS_ARCHIVED",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-2",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-3",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-4",
						Status:              "STUDY_PLAN_ITEM_STATUS_ARCHIVED",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-3",
						},
						StudentID:           "student-3",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-6",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-2",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-3",
						StudyPlanItemID:     "spi-5",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
				}, nil)

				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-5"})).
					Return([]*repositories.CalculateHighestScoreResponse{
						{
							StudyPlanItemID: database.Text("spi-5"),
							Percentage:      database.Float4(44),
						},
					}, nil)

				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-5"})).Once().
					Return([]*repositories.CalculateHighestScoreResponse{}, nil)

				mockShuffleQuizSetRepo.On("CalculateHighestSubmissionScore", mock.Anything, mock.Anything, database.TextArray(nil)).Once().
					Return([]*repositories.CalculateHighestScoreResponse{}, nil)
			},
		},
		{
			name: "empty request - response",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequest{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     "ClassId",
			},
			expectedResp: &pb.RetrieveCourseStatisticResponse{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItems", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItem{}, nil)
				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{}, nil)
				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{}, nil)
				mockShuffleQuizSetRepo.On("CalculateHighestSubmissionScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{}, nil)
			},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(ctx)
			}
			resp, err := svc.RetrieveCourseStatistic(test.ctx, test.req.(*pb.RetrieveCourseStatisticRequest))
			assert.Equal(t, test.expectedErr, err)
			if err == nil {
				assert.Equal(t, test.expectedResp.(*pb.RetrieveCourseStatisticResponse), resp)
			}
		})
	}
}

func TestCourseReaderService_RetrieveCourseStatisticV2(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockCourseStudentRepo := &mock_repositories.MockCourseStudyPlanRepo{}
	mockAssignmentRepo := &mock_repositories.MockAssignmentRepo{}
	mockShuffleQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}
	svc := CourseReaderService{
		CourseStudyPlanRepo: mockCourseStudentRepo,
		AssignmentRepo:      mockAssignmentRepo,
		ShuffledQuizSetRepo: mockShuffleQuizSetRepo,
	}

	cases := []TestCase{
		{
			name: "calculate assignment score error",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequestV2{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     []string{"ClassId"},
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, "AssignmentRepo.CalculateHigestScore %v", pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItemsV2", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItemV2{}, nil)
				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "calculate task assignment score error",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequestV2{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     []string{"ClassId"},
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, "AssignmentRepo.CalculateTaskAssignmentHighestScore %v", pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItemsV2", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItemV2{}, nil)
				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{}, nil)
				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "calculate shuffle quiz set score error",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequestV2{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     []string{"ClassId"},
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, "ShuffledQuizSetRepo.CalculateHighestSubmissionScore %v", pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItemsV2", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItemV2{}, nil)
				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{}, nil)
				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{}, nil)
				mockShuffleQuizSetRepo.On("CalculateHighestSubmissionScore", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "correct score all active item with no class ID",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequestV2{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     []string{},
			},
			expectedResp: &pb.RetrieveCourseStatisticResponseV2{
				TopicStatistic: []*pb.RetrieveCourseStatisticResponseV2_TopicStatistic{
					{
						TopicId:              "topic-1",
						CompletedStudent:     2,
						TotalAssignedStudent: 3,
						AverageScore:         28,
						LearningMaterialStatistic: []*pb.RetrieveCourseStatisticResponseV2_TopicStatistic_LearningMaterialStatistic{
							{
								StudyPlanItemId:      "root-1",
								CompletedStudent:     2,
								TotalAssignedStudent: 2,
								AverageScore:         38,
							},
							{
								StudyPlanItemId:      "root-2",
								CompletedStudent:     2,
								TotalAssignedStudent: 3,
								AverageScore:         17,
							},
						},
					},
					{
						TopicId:              "topic-2",
						CompletedStudent:     1,
						TotalAssignedStudent: 1,
						AverageScore:         91,
						LearningMaterialStatistic: []*pb.RetrieveCourseStatisticResponseV2_TopicStatistic_LearningMaterialStatistic{
							{
								StudyPlanItemId:      "root-3",
								CompletedStudent:     1,
								TotalAssignedStudent: 1,
								AverageScore:         91,
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItemsV2", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItemV2{
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-1",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-2",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-3",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-4",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-3",
						},
						StudentID:           "student-3",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-6",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-2",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-3",
						StudyPlanItemID:     "spi-5",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
				}, nil)

				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-1", "spi-2", "spi-5"})).
					Return([]*repositories.CalculateHighestScoreResponse{
						{
							StudyPlanItemID: database.Text("spi-1"),
							Percentage:      database.Float4(31),
						},
						{
							StudyPlanItemID: database.Text("spi-2"),
							Percentage:      database.Float4(44),
						},
					}, nil)

				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-1", "spi-2", "spi-5"})).Once().
					Return([]*repositories.CalculateHighestScoreResponse{
						{
							StudyPlanItemID: database.Text("spi-5"),
							Percentage:      database.Float4(91),
						},
					}, nil)

				mockShuffleQuizSetRepo.On("CalculateHighestSubmissionScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-3", "spi-4"})).Once().
					Return([]*repositories.CalculateHighestScoreResponse{
						{

							StudyPlanItemID: database.Text("spi-3"),
							Percentage:      database.Float4(15),
						},
						{

							StudyPlanItemID: database.Text("spi-4"),
							Percentage:      database.Float4(19),
						},
					}, nil)
			},
		},

		{
			name: "correct score all active item",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequestV2{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     []string{"ClassId1", "ClassId2"},
			},
			expectedResp: &pb.RetrieveCourseStatisticResponseV2{
				TopicStatistic: []*pb.RetrieveCourseStatisticResponseV2_TopicStatistic{
					{
						TopicId:              "topic-1",
						CompletedStudent:     2,
						TotalAssignedStudent: 3,
						AverageScore:         28,
						LearningMaterialStatistic: []*pb.RetrieveCourseStatisticResponseV2_TopicStatistic_LearningMaterialStatistic{
							{
								StudyPlanItemId:      "root-1",
								CompletedStudent:     2,
								TotalAssignedStudent: 2,
								AverageScore:         38,
							},
							{
								StudyPlanItemId:      "root-2",
								CompletedStudent:     2,
								TotalAssignedStudent: 3,
								AverageScore:         17,
							},
						},
					},
					{
						TopicId:              "topic-2",
						CompletedStudent:     1,
						TotalAssignedStudent: 1,
						AverageScore:         91,
						LearningMaterialStatistic: []*pb.RetrieveCourseStatisticResponseV2_TopicStatistic_LearningMaterialStatistic{
							{
								StudyPlanItemId:      "root-3",
								CompletedStudent:     1,
								TotalAssignedStudent: 1,
								AverageScore:         91,
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItemsV2", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItemV2{
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-1",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-2",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-3",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-4",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-3",
						},
						StudentID:           "student-3",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-6",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-2",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-3",
						StudyPlanItemID:     "spi-5",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
				}, nil)

				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-1", "spi-2", "spi-5"})).
					Return([]*repositories.CalculateHighestScoreResponse{
						{
							StudyPlanItemID: database.Text("spi-1"),
							Percentage:      database.Float4(31),
						},
						{
							StudyPlanItemID: database.Text("spi-2"),
							Percentage:      database.Float4(44),
						},
					}, nil)

				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-1", "spi-2", "spi-5"})).Once().
					Return([]*repositories.CalculateHighestScoreResponse{
						{
							StudyPlanItemID: database.Text("spi-5"),
							Percentage:      database.Float4(91),
						},
					}, nil)

				mockShuffleQuizSetRepo.On("CalculateHighestSubmissionScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-3", "spi-4"})).Once().
					Return([]*repositories.CalculateHighestScoreResponse{
						{

							StudyPlanItemID: database.Text("spi-3"),
							Percentage:      database.Float4(15),
						},
						{

							StudyPlanItemID: database.Text("spi-4"),
							Percentage:      database.Float4(19),
						},
					}, nil)
			},
		},
		{
			name: "archived items are ignored",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequestV2{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     []string{"ClassId1", "ClassId2"},
			},
			expectedResp: &pb.RetrieveCourseStatisticResponseV2{
				TopicStatistic: []*pb.RetrieveCourseStatisticResponseV2_TopicStatistic{
					{
						TopicId:              "topic-1",
						CompletedStudent:     2,
						TotalAssignedStudent: 3,
						AverageScore:         30,
						LearningMaterialStatistic: []*pb.RetrieveCourseStatisticResponseV2_TopicStatistic_LearningMaterialStatistic{
							{
								StudyPlanItemId:      "root-1",
								CompletedStudent:     1,
								TotalAssignedStudent: 1,
								AverageScore:         44,
							},
							{
								StudyPlanItemId:      "root-2",
								CompletedStudent:     1,
								TotalAssignedStudent: 2,
								AverageScore:         15,
							},
						},
					},
					{
						TopicId:      "topic-2",
						AverageScore: -1,
						LearningMaterialStatistic: []*pb.RetrieveCourseStatisticResponseV2_TopicStatistic_LearningMaterialStatistic{
							{
								StudyPlanItemId: "root-3",
								AverageScore:    -1,
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItemsV2", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItemV2{
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-1",
						Status:              "STUDY_PLAN_ITEM_STATUS_ARCHIVED",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-2",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-3",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-4",
						Status:              "STUDY_PLAN_ITEM_STATUS_ARCHIVED",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-3",
						},
						StudentID:           "student-3",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-6",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-2",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-3",
						StudyPlanItemID:     "spi-5",
						Status:              "STUDY_PLAN_ITEM_STATUS_ARCHIVED",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
				}, nil)

				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-2"})).
					Return([]*repositories.CalculateHighestScoreResponse{
						{
							StudyPlanItemID: database.Text("spi-2"),
							Percentage:      database.Float4(44),
						},
					}, nil)

				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-2"})).Once().
					Return([]*repositories.CalculateHighestScoreResponse{}, nil)

				mockShuffleQuizSetRepo.On("CalculateHighestSubmissionScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-3"})).Once().
					Return([]*repositories.CalculateHighestScoreResponse{
						{
							StudyPlanItemID: database.Text("spi-3"),
							Percentage:      database.Float4(15),
						},
					}, nil)
			},
		},
		{
			name: "incomplete items doesn't include in average score",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequestV2{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     []string{"ClassId1", "ClassId2"},
			},
			expectedResp: &pb.RetrieveCourseStatisticResponseV2{
				TopicStatistic: []*pb.RetrieveCourseStatisticResponseV2_TopicStatistic{
					{
						TopicId:              "topic-1",
						CompletedStudent:     0,
						TotalAssignedStudent: 3,
						AverageScore:         -1,
						LearningMaterialStatistic: []*pb.RetrieveCourseStatisticResponseV2_TopicStatistic_LearningMaterialStatistic{
							{
								StudyPlanItemId:      "root-1",
								CompletedStudent:     0,
								TotalAssignedStudent: 1,
								AverageScore:         -1,
							},
							{
								StudyPlanItemId:      "root-2",
								CompletedStudent:     0,
								TotalAssignedStudent: 2,
								AverageScore:         -1,
							},
						},
					},
					{
						TopicId:              "topic-2",
						CompletedStudent:     1,
						TotalAssignedStudent: 1,
						AverageScore:         44,
						LearningMaterialStatistic: []*pb.RetrieveCourseStatisticResponseV2_TopicStatistic_LearningMaterialStatistic{
							{
								StudyPlanItemId:      "root-3",
								CompletedStudent:     1,
								TotalAssignedStudent: 1,
								AverageScore:         44,
							},
						},
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItemsV2", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItemV2{
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-1",
						Status:              "STUDY_PLAN_ITEM_STATUS_ARCHIVED",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-1",
							AssignmentID: "ass-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-1",
						StudyPlanItemID:     "spi-2",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-3",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-2",
						},
						StudentID:           "student-2",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-4",
						Status:              "STUDY_PLAN_ITEM_STATUS_ARCHIVED",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID: "topic-1",
							LoID:    "lo-3",
						},
						StudentID:           "student-3",
						RootStudyPlanItemID: "root-2",
						StudyPlanItemID:     "spi-6",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         pgtype.Timestamptz{Status: pgtype.Null},
					},
					{
						ContentStructure: entities.ContentStructure{
							TopicID:      "topic-2",
							AssignmentID: "ass-1",
						},
						StudentID:           "student-1",
						RootStudyPlanItemID: "root-3",
						StudyPlanItemID:     "spi-5",
						Status:              "STUDY_PLAN_ITEM_STATUS_ACTIVE",
						CompletedAt:         database.Timestamptz(time.Now()),
					},
				}, nil)

				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-5"})).
					Return([]*repositories.CalculateHighestScoreResponse{
						{
							StudyPlanItemID: database.Text("spi-5"),
							Percentage:      database.Float4(44),
						},
					}, nil)

				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, database.TextArray([]string{"spi-5"})).Once().
					Return([]*repositories.CalculateHighestScoreResponse{}, nil)

				mockShuffleQuizSetRepo.On("CalculateHighestSubmissionScore", mock.Anything, mock.Anything, database.TextArray(nil)).Once().
					Return([]*repositories.CalculateHighestScoreResponse{}, nil)
			},
		},
		{
			name: "empty request - response",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequestV2{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     []string{"ClassId"},
			},
			expectedResp: &pb.RetrieveCourseStatisticResponseV2{},
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItemsV2", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CourseStatisticItemV2{}, nil)
				mockAssignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{}, nil)
				mockAssignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{}, nil)
				mockShuffleQuizSetRepo.On("CalculateHighestSubmissionScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{}, nil)
			},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(ctx)
			}
			resp, err := svc.RetrieveCourseStatisticV2(test.ctx, test.req.(*pb.RetrieveCourseStatisticRequestV2))
			assert.Equal(t, test.expectedErr, err)
			if err == nil {
				assert.Equal(t, test.expectedResp.(*pb.RetrieveCourseStatisticResponseV2), resp)
			}
		})
	}
}

func TestCourseReaderService_ListCourseStatistic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	mockCourseStudentRepo := &mock_repositories.MockCourseStudyPlanRepo{}
	mockAssignmentRepo := &mock_repositories.MockAssignmentRepo{}
	mockShuffleQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}
	svc := CourseReaderService{
		CourseStudyPlanRepo: mockCourseStudentRepo,
		AssignmentRepo:      mockAssignmentRepo,
		ShuffledQuizSetRepo: mockShuffleQuizSetRepo,
	}

	cases := []TestCase{
		{
			name: "invalid request missing course id",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequestV2{
				CourseId:    "",
				StudyPlanId: "StudyPlanId",
				ClassId:     []string{"ClassId"},
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.InvalidArgument, "Missing course"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "invalid request missing study plan",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequestV2{
				CourseId:    "CourseId",
				StudyPlanId: "",
				ClassId:     []string{"ClassId"},
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.InvalidArgument, "Missing study plan"),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "list topic statistic error",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequestV2{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     []string{"ClassId"},
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, "CourseStudyPlanRepo.ListCourseStatisticItemsV2 %v", pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItemsV2", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "list topic statistic error",
			ctx:  ctx,
			req: &pb.RetrieveCourseStatisticRequestV2{
				CourseId:    "CourseId",
				StudyPlanId: "StudyPlanId",
				ClassId:     []string{"ClassId"},
			},
			expectedResp: nil,
			expectedErr:  status.Errorf(codes.Internal, "CourseStudyPlanRepo.ListCourseStatisticItemsV2 %v", pgx.ErrNoRows.Error()),
			setup: func(ctx context.Context) {
				mockCourseStudentRepo.On("ListCourseStatisticItemsV2", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if test.setup != nil {
				test.setup(ctx)
			}
			resp, err := svc.RetrieveCourseStatisticV2(test.ctx, test.req.(*pb.RetrieveCourseStatisticRequestV2))
			assert.Equal(t, test.expectedErr, err)
			if err == nil {
				assert.Equal(t, test.expectedResp.(*pb.RetrieveCourseStatisticResponseV2), resp)
			}
		})
	}
}

func TestCourseReaderService_GetLOsByCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	mockCourseBookRepo := &mock_repositories.MockCourseBookRepo{}
	mockChapterRepo := &mock_repositories.MockChapterRepo{}
	mockTopicRepo := &mock_repositories.MockTopicRepo{}
	mockLORepo := &mock_repositories.MockLearningObjectiveRepoV2{}
	mockAssessmentRepo := &mock_repositories.MockAssessmentRepo{}
	db := new(mock_database.Ext)
	var listResponseLOsNil []*pb.GetLOsByCourseResponse_LearningObject

	defer cancel()

	s := CourseReaderService{
		DB:                      db,
		CourseBookRepo:          mockCourseBookRepo,
		ChapterRepo:             mockChapterRepo,
		TopicRepo:               mockTopicRepo,
		LearningObjectiveRepoV2: mockLORepo,
		AssessmentRepo:          mockAssessmentRepo,
	}

	courseBook1 := entities.CoursesBooks{BookID: database.Text("book-1")}
	courseBook2 := entities.CoursesBooks{BookID: database.Text("book-2")}
	listBook := []*entities.CoursesBooks{&courseBook1, &courseBook2}
	listBookNil := []*entities.CoursesBooks{}

	bookChapter1 := entities.Chapter{ID: database.Text("chapter-1")}
	bookChapter2 := entities.Chapter{ID: database.Text("chapter-2")}
	listBookChapter := []*entities.Chapter{&bookChapter1, &bookChapter2}
	listBookChapterNil := []*entities.Chapter{}

	topic1 := entities.Topic{
		ID:   database.Text("topic-1"),
		Name: database.Text("topic-name-1"),
	}
	topic2 := entities.Topic{
		ID:   database.Text("topic-2"),
		Name: database.Text("topic-name-2"),
	}
	listTopics := []*entities.Topic{&topic1, &topic2}
	listTopicsNil := []*entities.Topic{}
	learningMaterial1 := entities.LearningMaterial{
		ID:      database.Text("lo-1"),
		Name:    database.Text("lo-name-1"),
		TopicID: database.Text("topic-1"),
	}

	lo1 := entities.LearningObjectiveV2{
		LearningMaterial: learningMaterial1,
	}

	listLOs := []*entities.LearningObjectiveV2{&lo1}
	listLOsNil := []*entities.LearningObjectiveV2{}

	responseLO1 := pb.GetLOsByCourseResponse_LearningObject{
		ActivityId:         "as-1",
		TopicName:          "topic-name-1",
		LoName:             "lo-name-1",
		LearningMaterialId: "lo-1",
	}

	listResponseLOs := []*pb.GetLOsByCourseResponse_LearningObject{&responseLO1, &responseLO1, &responseLO1, &responseLO1}

	assessment1 := entities.Assessment{
		ID: database.Text("as-1"),
	}

	listAssessments := []*entities.Assessment{&assessment1}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &pb.GetLOsByCourseRequest{
				CourseId: []string{"course-1", "course-2"},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &pb.GetLOsByCourseResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				LOs:        listResponseLOs,
				TotalItems: 2,
			},
			setup: func(ctx context.Context) {
				mockCourseBookRepo.On("FindByCourseIDsV2", ctx, db, mock.Anything, mock.Anything).
					Return(listBook, nil).Once()
				mockChapterRepo.On("FindByBookIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(listBookChapter, nil).Once()
				mockTopicRepo.On("FindByChapterIDs", ctx, db, mock.Anything, mock.Anything).
					Return(listTopics, nil).Once()
				mockTopicRepo.On("FindByChapterIDs", ctx, db, mock.Anything, mock.Anything).
					Return(listTopics, nil).Once()
				mockLORepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listLOs, nil).Once()
				mockLORepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listLOs, nil).Once()
				mockLORepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listLOs, nil).Once()
				mockLORepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listLOs, nil).Once()
				mockAssessmentRepo.On("GetAssessmentByCourseAndLearningMaterial", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listAssessments, nil).Once()
				mockAssessmentRepo.On("GetAssessmentByCourseAndLearningMaterial", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listAssessments, nil).Once()
				mockAssessmentRepo.On("GetAssessmentByCourseAndLearningMaterial", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listAssessments, nil).Once()
				mockAssessmentRepo.On("GetAssessmentByCourseAndLearningMaterial", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listAssessments, nil).Once()
				mockLORepo.On("CountLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(2, nil).Once()
			},
		},
		{
			name: "happy case when get empty list topics by chapters",
			ctx:  ctx,
			req: &pb.GetLOsByCourseRequest{
				CourseId: []string{"course-1", "course-2"},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &pb.GetLOsByCourseResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				LOs:        listResponseLOsNil,
				TotalItems: 2,
			},
			setup: func(ctx context.Context) {
				mockCourseBookRepo.On("FindByCourseIDsV2", ctx, db, mock.Anything, mock.Anything).
					Return(listBook, nil).Once()
				mockChapterRepo.On("FindByBookIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(listBookChapter, nil).Once()
				mockTopicRepo.On("FindByChapterIDs", ctx, db, mock.Anything, mock.Anything).
					Return(listTopicsNil, nil).Once()
				mockTopicRepo.On("FindByChapterIDs", ctx, db, mock.Anything, mock.Anything).
					Return(listTopics, nil).Once()
				mockLORepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listLOsNil, nil).Once()
				mockLORepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listLOsNil, nil).Once()
				mockLORepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listLOsNil, nil).Once()
				mockLORepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listLOsNil, nil).Once()
				mockLORepo.On("CountLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(2, nil).Once()
			},
		},
		{
			name: "happy case when get empty list books by course",
			ctx:  ctx,
			req: &pb.GetLOsByCourseRequest{
				CourseId: []string{"course-1", "course-2"},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &pb.GetLOsByCourseResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				LOs:        listResponseLOsNil,
				TotalItems: 2,
			},
			setup: func(ctx context.Context) {
				mockCourseBookRepo.On("FindByCourseIDsV2", ctx, db, mock.Anything, mock.Anything).
					Return(listBookNil, nil).Once()
				mockChapterRepo.On("FindByBookIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(listBookChapterNil, nil).Once()
				mockTopicRepo.On("FindByChapterIDs", ctx, db, mock.Anything, mock.Anything).
					Return(listTopicsNil, nil).Once()
				mockLORepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listLOsNil, nil).Once()
				mockLORepo.On("CountLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(2, nil).Once()
			},
		},
		{
			name: "happy case when get empty list LOs by topics",
			ctx:  ctx,
			req: &pb.GetLOsByCourseRequest{
				CourseId: []string{"course-1", "course-2"},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &pb.GetLOsByCourseResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				LOs:        listResponseLOsNil,
				TotalItems: 2,
			},
			setup: func(ctx context.Context) {
				mockCourseBookRepo.On("FindByCourseIDsV2", ctx, db, mock.Anything, mock.Anything).
					Return(listBook, nil).Once()
				mockChapterRepo.On("FindByBookIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(listBookChapter, nil).Once()
				mockTopicRepo.On("FindByChapterIDs", ctx, db, mock.Anything, mock.Anything).
					Return(listTopics, nil).Once()
				mockLORepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listLOsNil, nil).Once()
				mockLORepo.On("CountLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(2, nil).Once()
			},
		},
		{
			name: "happy case when get empty list chapters by books",
			ctx:  ctx,
			req: &pb.GetLOsByCourseRequest{
				CourseId: []string{"course-1", "course-2"},
				Paging: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			expectedErr: nil,
			expectedResp: &pb.GetLOsByCourseResponse{
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
				LOs:        listResponseLOsNil,
				TotalItems: 2,
			},
			setup: func(ctx context.Context) {
				mockCourseBookRepo.On("FindByCourseIDsV2", ctx, db, mock.Anything, mock.Anything).
					Return(listBook, nil).Once()
				mockChapterRepo.On("FindByBookIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(listBookChapterNil, nil).Once()
				mockTopicRepo.On("FindByChapterIDs", ctx, db, mock.Anything, mock.Anything).
					Return(listTopicsNil, nil).Once()
				mockLORepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(listLOsNil, nil).Once()
				mockLORepo.On("CountLearningObjectivesByTopicIDs", ctx, db, mock.Anything, mock.Anything, mock.Anything).
					Return(2, nil).Once()
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.GetLOsByCourseRequest)
			resp, err := s.GetLOsByCourse(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
