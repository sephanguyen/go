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
	mock_services "github.com/manabie-com/backend/mock/eureka/services"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	"github.com/pkg/errors"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var ErrSomethingWentWrong = fmt.Errorf("something went wrong")

func TestListToDoItemsByTopicsV2(t *testing.T) {

	t.Parallel()
	studyPlanItemRepo := &mock_repositories.MockStudyPlanItemRepo{}
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	assignmentRepo := &mock_repositories.MockAssignmentRepo{}
	learningObjectiveRepo := &mock_repositories.MockLearningObjectiveRepo{}
	bobCourseReaderClient := &mock_services.MockCourseReaderService{}
	bookRepo := &mock_repositories.MockBookRepo{}
	db := &mock_database.Ext{}

	topicReaderService := &TopicReaderService{
		StudyPlanItemRepo:     studyPlanItemRepo,
		StudyPlanRepo:         studyPlanRepo,
		BobCourseReaderClient: bobCourseReaderClient,
		AssignmentRepo:        assignmentRepo,
		LearningObjectiveRepo: learningObjectiveRepo,
		DB:                    db,
		BookRepo:              bookRepo,
	}

	req := &pb.ListToDoItemsByTopicsRequest{
		TopicIds: []string{
			"topic-1",
		},
		StudyPlanId: wrapperspb.String("study_plan-1"),
	}
	studyPlanItems := []*entities.StudyPlanItem{
		{
			StudyPlanID: database.Text("study_plan_id"),
			ContentStructure: database.JSONB(map[string]interface{}{
				"lo_id": "lo-id",
			}),
		},
		{
			StudyPlanID: database.Text("study_plan_id-2"),
			ContentStructure: database.JSONB(map[string]interface{}{
				"assignment_id": "assignment-id",
			}),
		},
	}
	testCases := []TestCase{
		{
			name: "happy case",
			req:  req,
			setup: func(ctx context.Context) {
				studyPlanRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				bookRepo.On("RetrieveBookTreeByTopicIDs", ctx, db, mock.Anything).Once().Return([]*repositories.BookTreeInfo{
					{
						LoID:                database.Text("lo-id"),
						TopicID:             database.Text("topic-id"),
						ChapterID:           database.Text("chapter-id"),
						LoDisplayOrder:      database.Int2(1),
						TopicDisplayOrder:   database.Int2(2),
						ChapterDisplayOrder: database.Int2(3),
					},
				}, nil)
				assignmentRepo.On("RetrieveAssignmentsByTopicIDs", ctx, db, mock.Anything).Once().Return([]*entities.Assignment{
					{
						ID: database.Text("assignment-id"),
						Content: database.JSONB(map[string]interface{}{
							"topic_id": "topic-id",
							"lo_id":    []string{"lo-id"},
						}),
					},
				}, nil)
				learningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything).Once().Return([]*entities.LearningObjective{
					{
						ID:      database.Text("lo-id"),
						TopicID: database.Text("topic-id"),
					},
				}, nil)
				studyPlanItemRepo.On("FindWithFilter", ctx, db, mock.Anything).Once().Return(studyPlanItems, nil)
			},
		},
		{
			name: "err study plan is null",
			req: &pb.ListToDoItemsByTopicsRequest{
				TopicIds: []string{
					"topic-1",
				},
			},
			expectedErr: status.Errorf(codes.InvalidArgument, ErrMustHaveStudyPlanID.Error()),
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "err studyPlanRepo.FindByID",
			req:         req,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("unable to retrieve study plan: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				studyPlanRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        "err studyPlanRepo.FindByID not found",
			req:         req,
			expectedErr: status.Errorf(codes.NotFound, "study plan not exist"),
			setup: func(ctx context.Context) {
				studyPlanRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name:        "err studyPlanItemRepo.FindWithFilter",
			req:         req,
			expectedErr: status.Error(codes.Internal, fmt.Errorf("unable to retrieve study plan items: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				studyPlanRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				bookRepo.On("RetrieveBookTreeByTopicIDs", ctx, db, mock.Anything).Once().Return([]*repositories.BookTreeInfo{}, nil)
				learningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything).Once().Return([]*entities.LearningObjective{}, nil)
				assignmentRepo.On("RetrieveAssignmentsByTopicIDs", ctx, db, mock.Anything).Once().Return([]*entities.Assignment{}, nil)
				studyPlanItemRepo.On("FindWithFilter", ctx, db, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name:        "studyPlanItemRepo.FindWithFilter no rows",
			req:         req,
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studyPlanRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				studyPlanItemRepo.On("FindWithFilter", ctx, db, mock.Anything, mock.Anything).Once().Return([]*entities.StudyPlanItem{}, nil)
				bookRepo.On("RetrieveBookTreeByTopicIDs", ctx, db, mock.Anything).Once().Return([]*repositories.BookTreeInfo{}, nil)
				learningObjectiveRepo.On("RetrieveLearningObjectivesByTopicIDs", ctx, db, mock.Anything).Once().Return([]*entities.LearningObjective{}, nil)
				assignmentRepo.On("RetrieveAssignmentsByTopicIDs", ctx, db, mock.Anything).Once().Return([]*entities.Assignment{}, nil)
			},
		},
		{
			name:        "err BookRepo.RetrieveBookTreeByTopicIDs",
			req:         req,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("BookRepo.RetrieveBookTreeByTopicIDs: %w", ErrSomethingWentWrong).Error()),
			setup: func(ctx context.Context) {
				studyPlanRepo.On("FindByID", ctx, db, mock.Anything).Once().Return(nil, nil)
				studyPlanItemRepo.On("FindWithFilter", ctx, db, mock.Anything, mock.Anything).Once().Return(studyPlanItems, nil)
				bookRepo.On("RetrieveBookTreeByTopicIDs", ctx, db, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := interceptors.NewIncomingContext(context.Background())
			testCase.setup(ctx)

			rsp, err := topicReaderService.ListToDoItemsByTopicsV2(ctx, testCase.req.(*pb.ListToDoItemsByTopicsRequest))

			if (testCase.expectedErr != nil) != (err != nil) {
				t.Errorf("expected error = %v but got = %v", testCase.expectedErr, err)
			}

			if testCase.expectedErr != nil {
				t.Logf("expectedErr %v", testCase.expectedErr)
				assert.Equal(t, testCase.expectedErr, err)
			}
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, rsp)
			} else {
				assert.Equal(t, testCase.expectedResp.(*pb.ListToDoItemsByTopicsRequest), rsp)
			}
		})
	}

}

func TestRetrieveTopics(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	topicRepo := new(mock_repositories.MockTopicRepo)
	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &pb.RetrieveTopicsRequest{
				TopicIds: []string{
					"1",
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]*entities.Topic{
					generateEnTopic(),
				}, nil)
			},
		},
		{
			name: "repo err case",
			ctx:  ctx,
			req: &pb.RetrieveTopicsRequest{
				TopicIds: []string{
					"1",
				},
			},
			expectedErr: errors.Wrap(pgx.ErrNoRows, "c.TopicRepo.RetrieveByIDs"),
			setup: func(ctx context.Context) {
				topicRepo.On("RetrieveByIDs", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
	}
	s := &TopicReaderService{
		TopicRepo: topicRepo,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.RetrieveTopicsRequest)
			_, err := s.RetrieveTopics(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func generateEnTopic() *entities.Topic {
	e := new(entities.Topic)
	e.ID.Set("1")
	e.Name.Set("Some topic name")
	e.Country.Set("COUNTRY_VN")
	e.Grade.Set("2")
	e.Subject.Set("SUBJECT_MATHS")
	e.TopicType.Set("TOPIC_TYPE_LEARNING")
	e.CreatedAt.Set(time.Now())
	e.UpdatedAt.Set(time.Now())
	e.DeletedAt.Set(nil)
	e.Status.Set("TOPIC_STATUS_DRAFT")
	e.PublishedAt.Set(nil)
	return e
}
