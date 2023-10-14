package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestListSubmissionsV3(t *testing.T) {
	t.Parallel()
	studentSubmissionRepo := &mock_repositories.MockStudentSubmissionRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	mockDB := &mock_database.Ext{}
	svc := &StudentSubmissionService{
		DB:                    mockDB,
		StudentSubmissionRepo: studentSubmissionRepo,
		StudentRepo:           studentRepo,
		StudyPlanRepo:         studyPlanRepo,
	}

	validReq := &sspb.ListSubmissionsV3Request{
		ClassIds: []string{"class-id-1,class-id-2"},
		CourseId: wrapperspb.String("course-id-1"),
		Start:    timestamppb.Now(),
		End:      timestamppb.Now(),
		Statuses: []sspb.SubmissionStatus{
			sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS,
			sspb.SubmissionStatus_SUBMISSION_STATUS_MARKED,
		},
		Paging: &cpb.Paging{
			Limit: 10,
		},
	}

	testCases := []TestCase{
		{
			name:        "error list submission",
			req:         validReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				studentSubmissionRepo.On("ListV3", ctx, mockDB, mock.Anything).Once().Return(
					nil, pgx.ErrNoRows,
				)
			},
		},
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {

				studyPlan1 := &entities.StudyPlan{}
				studyPlan1.ID.Set("spi1")

				studyPlan2 := &entities.StudyPlan{}
				studyPlan2.ID.Set("spi2")

				submission := []*repositories.StudentSubmissionInfo{
					{
						CourseID: database.Text("course-id"),
						StudentSubmission: entities.StudentSubmission{
							AssignmentID:       database.Text("assignment-id"),
							ID:                 database.Text("id"),
							StudentID:          database.Text("student-id"),
							LearningMaterialID: database.Text("learning_material-id"),
							StudyPlanID:        database.Text("study-plan-id"),
						},
						StartDate: database.Timestamptz(time.Now().Add(-1 * time.Hour)),
						EndDate:   database.Timestamptz(time.Now().Add(1 * time.Hour)),
					},
				}
				studentSubmissionRepo.On("ListV3", ctx, mockDB, mock.Anything).Once().Return(
					submission, nil,
				)

				studyPlanRepo.On("FindByIDs",
					ctx,
					mockDB,
					database.TextArray([]string{
						studyPlan1.ID.String,
						studyPlan2.ID.String,
					}),
				).Once().Return([]*entities.StudyPlan{studyPlan1, studyPlan2}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.ListSubmissionsV3(ctx, testCase.req.(*sspb.ListSubmissionsV3Request))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestListSubmissionsV4(t *testing.T) {
	t.Parallel()
	studentSubmissionRepo := &mock_repositories.MockStudentSubmissionRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	mockDB := &mock_database.Ext{}
	svc := &StudentSubmissionService{
		DB:                    mockDB,
		StudentSubmissionRepo: studentSubmissionRepo,
		StudentRepo:           studentRepo,
		StudyPlanRepo:         studyPlanRepo,
	}

	validReq := &sspb.ListSubmissionsV4Request{
		ClassIds: []string{"class-id-1,class-id-2"},
		CourseId: wrapperspb.String("course-id-1"),
		Start:    timestamppb.Now(),
		End:      timestamppb.Now(),
		Statuses: []sspb.SubmissionStatus{
			sspb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS,
			sspb.SubmissionStatus_SUBMISSION_STATUS_MARKED,
		},
		Paging: &cpb.Paging{
			Limit: 10,
		},
	}

	testCases := []TestCase{
		{
			name:        "error list submission",
			req:         validReq,
			expectedErr: status.Errorf(codes.Internal, fmt.Errorf("database.ExecInTx: %w", pgx.ErrNoRows).Error()),
			setup: func(ctx context.Context) {
				studentSubmissionRepo.On("ListV4", ctx, mockDB, mock.Anything).Once().Return(
					nil, pgx.ErrNoRows,
				)
			},
		},
		{
			name:        "happy case",
			req:         validReq,
			expectedErr: nil,
			setup: func(ctx context.Context) {

				studyPlan1 := &entities.StudyPlan{}
				studyPlan1.ID.Set("spi1")

				studyPlan2 := &entities.StudyPlan{}
				studyPlan2.ID.Set("spi2")

				submission := []*repositories.StudentSubmissionInfo{
					{
						CourseID: database.Text("course-id"),
						StudentSubmission: entities.StudentSubmission{
							AssignmentID:       database.Text("assignment-id"),
							ID:                 database.Text("id"),
							StudentID:          database.Text("student-id"),
							LearningMaterialID: database.Text("learning_material-id"),
							StudyPlanID:        database.Text("study-plan-id"),
						},
						StartDate: database.Timestamptz(time.Now().Add(-1 * time.Hour)),
						EndDate:   database.Timestamptz(time.Now().Add(1 * time.Hour)),
					},
				}
				studentSubmissionRepo.On("ListV4", ctx, mockDB, mock.Anything).Once().Return(
					submission, nil,
				)

				studyPlanRepo.On("FindByIDs",
					ctx,
					mockDB,
					database.TextArray([]string{
						studyPlan1.ID.String,
						studyPlan2.ID.String,
					}),
				).Once().Return([]*entities.StudyPlan{studyPlan1, studyPlan2}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		_, err := svc.ListSubmissionsV4(ctx, testCase.req.(*sspb.ListSubmissionsV4Request))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentSubmissionService_RetrieveSubmissionHistory(t *testing.T) {
	ctx := context.Background()
	studentSubmissionRepo := &mock_repositories.MockStudentSubmissionRepo{}
	studentRepo := &mock_repositories.MockStudentRepo{}
	studyPlanRepo := &mock_repositories.MockStudyPlanRepo{}
	mockDB := &mock_database.Ext{}
	quizRepo := &mock_repositories.MockQuizRepo{}
	shuffledQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}
	questionGroupRepo := &mock_repositories.MockQuestionGroupRepo{}
	loSubmissionAnswerRepo := &mock_repositories.MockLOSubmissionAnswerRepo{}
	flashCardSubmissionAnswerRepo := &mock_repositories.MockFlashCardSubmissionAnswerRepo{}
	svc := &StudentSubmissionService{
		DB:                            mockDB,
		StudentSubmissionRepo:         studentSubmissionRepo,
		StudentRepo:                   studentRepo,
		StudyPlanRepo:                 studyPlanRepo,
		QuizRepo:                      quizRepo,
		ShuffledQuizSetRepo:           shuffledQuizSetRepo,
		QuestionGroupRepo:             questionGroupRepo,
		LOSubmissionAnswerRepo:        loSubmissionAnswerRepo,
		FlashCardSubmissionAnswerRepo: flashCardSubmissionAnswerRepo,
	}
	now := time.Now()
	testCases := []TestCase{
		{
			name: "happy case LO",
			req: &sspb.RetrieveSubmissionHistoryRequest{
				SetId: "set-1",
			},
			setup: func(ctx context.Context) {
				shuffledQuizSetRepo.
					On("GetRelatedLearningMaterial", mock.Anything, mock.Anything, database.Text("set-1")).
					Once().
					Return(
						&entities.LearningMaterial{
							ID:   database.Text("lm-1"),
							Type: database.Text("LEARNING_MATERIAL_LEARNING_OBJECTIVE"),
						},
						nil,
					)
				loSubmissionAnswerRepo.
					On("ListSubmissionAnswers", mock.Anything, mock.Anything, database.Text("set-1"), database.Int8(100), database.Int8(0)).
					Once().Return(
					[]*entities.LOSubmissionAnswer{
						{
							QuizID:            database.Text("quiz-1"),
							ShuffledQuizSetID: database.Text("set-1"),
						},
					},
					[]pgtype.Text{database.Text("quiz-1")},
					nil,
				)
				quizRepo.On("GetByExternalIDs", mock.Anything, mock.Anything, database.TextArray([]string{"quiz-1"}), database.Text("lm-1")).
					Once().
					Return(entities.Quizzes{
						{
							ID:         database.Text("quiz-1"),
							ExternalID: database.Text("quiz-1"),
							UpdatedAt:  database.Timestamptz(now),
							CreatedAt:  database.Timestamptz(now),
							DeletedAt:  pgtype.Timestamptz{},
						},
					}, nil)
				shuffledQuizSetRepo.On("GetSeed", mock.Anything, mock.Anything, database.Text("set-1")).
					Once().Return(database.Text("123"), nil)
				shuffledQuizSetRepo.On("GetQuizIdx", mock.Anything, mock.Anything, database.Text("set-1"), database.Text("quiz-1")).
					Once().Return(database.Int4(1), nil)

			},
		},
		{
			name: "happy case Flashcard",
			req: &sspb.RetrieveSubmissionHistoryRequest{
				SetId: "set-1",
			},
			setup: func(ctx context.Context) {
				shuffledQuizSetRepo.
					On("GetRelatedLearningMaterial", mock.Anything, mock.Anything, database.Text("set-1")).
					Once().
					Return(
						&entities.LearningMaterial{
							ID:   database.Text("lm-1"),
							Type: database.Text("LEARNING_MATERIAL_FLASH_CARD"),
						},
						nil,
					)
				flashCardSubmissionAnswerRepo.
					On("ListSubmissionAnswers", mock.Anything, mock.Anything, database.Text("set-1"), database.Int8(100), database.Int8(0)).
					Once().Return(
					[]*entities.FlashCardSubmissionAnswer{
						{
							QuizID:            database.Text("quiz-1"),
							ShuffledQuizSetID: database.Text("set-1"),
						},
					},
					[]pgtype.Text{database.Text("quiz-1")},
					nil,
				)
				quizRepo.On("GetByExternalIDs", mock.Anything, mock.Anything, database.TextArray([]string{"quiz-1"}), database.Text("lm-1")).
					Once().
					Return(entities.Quizzes{
						{
							ID:         database.Text("quiz-1"),
							ExternalID: database.Text("quiz-1"),
							UpdatedAt:  database.Timestamptz(now),
							CreatedAt:  database.Timestamptz(now),
							DeletedAt:  pgtype.Timestamptz{},
						},
					}, nil)
				shuffledQuizSetRepo.On("GetSeed", mock.Anything, mock.Anything, database.Text("set-1")).
					Once().Return(database.Text("123"), nil)
				shuffledQuizSetRepo.On("GetQuizIdx", mock.Anything, mock.Anything, database.Text("set-1"), database.Text("quiz-1")).
					Once().Return(database.Int4(1), nil)

			},
		},
	}
	for _, testCase := range testCases {
		testCase.setup(ctx)
		_, err := svc.RetrieveSubmissionHistory(ctx, testCase.req.(*sspb.RetrieveSubmissionHistoryRequest))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
