package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_services "github.com/manabie-com/backend/mock/eureka/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestRetrieveStudyPlanByCourse(t *testing.T) {
	t.Parallel()
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	mockDB := &mock_database.Ext{}
	svc := &StudyPlanReaderService{
		DB:            mockDB,
		StudyPlanRepo: studyPlanRepo,
	}

	validReq := &pb.ListStudyPlanByCourseRequest{
		CourseId: "course-id-1",
		Paging:   &cpb.Paging{},
	}

	testCases := []TestCase{
		{
			name:        "error no rows find study plans by course id",
			req:         validReq,
			expectedErr: fmt.Errorf("StudyPlanRepo.RetrieveByCourseID: %w", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				studyPlanRepo.On("RetrieveByCourseID", ctx, mockDB, mock.Anything).Once().Return(
					nil,
					pgx.ErrNoRows,
				)
			},
		},
		{
			name: "error list study plan by course without course id ",
			req: &pb.ListStudyPlanByCourseRequest{
				CourseId: "",
				Paging:   &cpb.Paging{},
			},
			expectedErr: errors.New("rpc error: code = InvalidArgument desc = invalid argument: course id have to not empty"),
			setup: func(ctx context.Context) {
				studyPlans := []*entities.StudyPlan{
					{
						ID:   database.Text("study-plan-id-1"),
						Name: database.Text("study-plan-name-1"),
					},
				}
				studyPlanRepo.On("RetrieveByCourseID", ctx, mockDB, mock.Anything).Once().Return(
					studyPlans,
					nil,
				)
			},
		},
		{
			name:        "happy case ",
			req:         validReq,
			expectedErr: nil,
			expectedResp: &pb.ListStudyPlanByCourseResponse{
				StudyPlans: []*pb.StudyPlan{
					{
						StudyPlanId: "study-plan-id-1",
						Name:        "study-plan-name-1",
					},
				},
				NextPage: &cpb.Paging{
					Limit: 10,
					Offset: &cpb.Paging_OffsetMultipleCombined{
						&cpb.Paging_MultipleCombined{
							Combined: []*cpb.Paging_Combined{{OffsetString: "study-plan-name-1"}, {OffsetString: "study-plan-id-1"}},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				studyPlans := []*entities.StudyPlan{
					{
						ID:   database.Text("study-plan-id-1"),
						Name: database.Text("study-plan-name-1"),
					},
				}
				studyPlanRepo.On("RetrieveByCourseID", ctx, mockDB, mock.Anything).Once().Return(
					studyPlans,
					nil,
				)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		res, err := svc.ListStudyPlanByCourse(ctx, testCase.req.(*pb.ListStudyPlanByCourseRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, res)
		}
	}
}

func TestGetBookIDsBelongsToStudentStudyPlan(t *testing.T) {
	t.Parallel()
	studentStudyPlan := &mock_repositories.MockStudentStudyPlanRepo{}
	mockDB := &mock_database.Ext{}
	svc := &StudyPlanReaderService{
		DB:                   mockDB,
		StudentStudyPlanRepo: studentStudyPlan,
	}

	testCases := []TestCase{
		{
			name: "error no studentID",
			req: &pb.GetBookIDsBelongsToStudentStudyPlanRequest{
				StudentId: "",
				BookIds:   []string{},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("GetBookIDsBelongsToStudentStudyPlan: studentID is not provided").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "error no bookIDs",
			req: &pb.GetBookIDsBelongsToStudentStudyPlanRequest{
				StudentId: "student_id",
				BookIds:   []string{},
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("GetBookIDsBelongsToStudentStudyPlan: bookIDs is not provided").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "error get books belong to student study plan",
			req: &pb.GetBookIDsBelongsToStudentStudyPlanRequest{
				StudentId: "student_id",
				BookIds:   []string{"book_id"},
			},
			expectedErr: status.Error(codes.Internal, fmt.Errorf("StudentStudyPlanRepo.GetBookIDsBelongsToStudentStudyPlan: %s", "get error").Error()),
			setup: func(ctx context.Context) {
				studentStudyPlan.On("GetBookIDsBelongsToStudentStudyPlan", ctx, mockDB, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("get error"))
			},
		},
		{
			name: "happy case ",
			req: &pb.GetBookIDsBelongsToStudentStudyPlanRequest{
				StudentId: "student_id",
				BookIds:   []string{"book_id"},
			},
			expectedErr: nil,
			expectedResp: &pb.GetBookIDsBelongsToStudentStudyPlanResponse{
				BookIds: []string{"book_id"},
			},
			setup: func(ctx context.Context) {
				studentStudyPlan.On("GetBookIDsBelongsToStudentStudyPlan", ctx, mockDB, mock.Anything, mock.Anything).Once().Return([]string{"book_id"}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		res, err := svc.GetBookIDsBelongsToStudentStudyPlan(ctx, testCase.req.(*pb.GetBookIDsBelongsToStudentStudyPlanRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, res)
		}
	}
}

func TestStudentBookStudyProgress(t *testing.T) {
	t.Parallel()
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	assignmentRepo := &mock_repositories.MockAssignmentRepo{}
	internalReaderService := &mock_services.BobInternalReaderServiceClient{}
	shuffledQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}

	mockDB := &mock_database.Ext{}
	svc := &StudyPlanReaderService{
		DB:                    mockDB,
		StudyPlanItemRepo:     studyPlanItemRepo,
		AssignmentRepo:        assignmentRepo,
		InternalReaderService: internalReaderService,
		ShuffledQuizSetRepo:   shuffledQuizSetRepo,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			req: &pb.StudentBookStudyProgressRequest{
				BookId:    "book_id",
				CourseId:  "course_id",
				StudentId: "student_id",
			},
			expectedErr: nil,
			expectedResp: &pb.StudentBookStudyProgressResponse{
				TopicProgress: []*pb.StudentTopicStudyProgress{
					{
						TopicId:                "topic_id",
						CompletedStudyPlanItem: wrapperspb.Int32(2),
						TotalStudyPlanItem:     wrapperspb.Int32(2),
						AverageScore:           wrapperspb.Int32(50),
					},
				},
				ChapterProgress: []*pb.StudentChapterStudyProgress{
					{
						ChapterId: "chapter_id",
						//AverageScore: wrapperspb.Int32(50),
					},
				},
			},
			setup: func(ctx context.Context) {
				studyPlanItemRepo.On("FetchByStudyProgressRequest", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.StudyPlanItem{
					{
						ID: database.Text("study_plan_item_id_1"),
						ContentStructure: database.JSONB(`{
              "lo_id": "lo_id",
              "topic_id": "topic_id",
              "chapter_id": "chapter_id"
            }`),
						CompletedAt: database.Timestamptz(time.Now()),
					},
					{
						ID: database.Text("study_plan_item_id_2"),
						ContentStructure: database.JSONB(`{
              "assignment_id": "assignment_id",
              "topic_id": "topic_id",
              "chapter_id": "chapter_id"
            }`),
						CompletedAt: database.Timestamptz(time.Now()),
					},
				}, nil)
				assignmentRepo.On("CalculateHigestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{
					{
						StudyPlanItemID: database.Text("study_plan_item_id_1"),
						Percentage:      database.Float4(50),
					},
				}, nil)
				assignmentRepo.On("CalculateTaskAssignmentHighestScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{
					{
						StudyPlanItemID: database.Text("study_plan_item_id_2"),
						Percentage:      database.Float4(50),
					},
				}, nil)

				shuffledQuizSetRepo.On("CalculateHighestSubmissionScore", mock.Anything, mock.Anything, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{
					{
						StudyPlanItemID: database.Text("study_plan_item_id_1"),
						Percentage:      database.Float4(50),
					},
				}, nil)
				internalReaderService.On("RetrieveTopics", mock.Anything, mock.Anything).Once().Return(&bpb.RetrieveTopicsResponse{
					Items: []*cpb.Topic{},
				}, nil)
			},
		},
		{
			name: "error empty course id",
			req: &pb.StudentBookStudyProgressRequest{
				BookId:    "book_id",
				StudentId: "student_id",
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("StudyPlanReaderService.StudentBookStudyProgress: course_id is empty").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "error empty book id",
			req: &pb.StudentBookStudyProgressRequest{
				CourseId:  "course_id",
				StudentId: "student_id",
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("StudyPlanReaderService.StudentBookStudyProgress: book_id is empty").Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "error empty student id",
			req: &pb.StudentBookStudyProgressRequest{
				BookId:   "book_id",
				CourseId: "course_id",
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("StudyPlanReaderService.StudentBookStudyProgress: student_id is empty").Error()),
			setup: func(ctx context.Context) {
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
				"token":   []string{"token"},
				"pkg":     []string{"package"},
				"version": []string{"version"},
			})
			testCase.setup(ctx)
			resp, err := svc.StudentBookStudyProgress(ctx, testCase.req.(*pb.StudentBookStudyProgressRequest))
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})

	}
}

func TestStudyPlanReaderService_GetLOHighestScoresByStudyPlanItemIDs(t *testing.T) {
	t.Parallel()
	mockDB := &mock_database.Ext{}

	mockShuffedQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}
	svc := &StudyPlanReaderService{
		DB:                  mockDB,
		ShuffledQuizSetRepo: mockShuffedQuizSetRepo,
	}

	testCases := []TestCase{
		{
			name:        "happy case ",
			req:         &pb.GetLOHighestScoresByStudyPlanItemIDsRequest{},
			expectedErr: nil,
			expectedResp: &pb.GetLOHighestScoresByStudyPlanItemIDsResponse{
				LoHighestScores: []*pb.GetLOHighestScoresByStudyPlanItemIDsResponse_LOHighestScore{
					{
						StudyPlanItemId: "study-plan-item-id-1",
						Percentage:      3.14,
					},
					{
						StudyPlanItemId: "study-plan-item-id-1",
						Percentage:      42,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockShuffedQuizSetRepo.On("CalculateHighestSubmissionScore", ctx, mockDB, mock.Anything).Once().Return([]*repositories.CalculateHighestScoreResponse{
					{
						StudyPlanItemID: database.Text("study-plan-item-id-1"),
						Percentage:      database.Float4(3.14),
					},
					{
						StudyPlanItemID: database.Text("study-plan-item-id-1"),
						Percentage:      database.Float4(42),
					},
				}, nil)
			},
		},
		{
			name:        "ShuffedQuizSetRepo error",
			req:         &pb.GetLOHighestScoresByStudyPlanItemIDsRequest{},
			expectedErr: status.Errorf(codes.Internal, "StudyPlanReaderService.GetLOHighestScoresByStudyPlanItemIDs.CalculateHigestSubmissionScore: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				mockShuffedQuizSetRepo.On("CalculateHighestSubmissionScore", ctx, mockDB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := interceptors.NewIncomingContext(context.Background())
		testCase.setup(ctx)
		_, err := svc.GetLOHighestScoresByStudyPlanItemIDs(ctx, testCase.req.(*pb.GetLOHighestScoresByStudyPlanItemIDsRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
