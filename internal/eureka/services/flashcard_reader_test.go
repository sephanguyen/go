package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	bpb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCourseReaderService_RetrieveFlashCardStudyProgress(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	flashcardProgressionRepo := &mock_repositories.MockFlashcardProgressionRepo{}
	quizRepo := &mock_repositories.MockQuizRepo{}

	s := &FlashCardReaderService{
		DB:                       db,
		QuizRepo:                 quizRepo,
		FlashcardProgressionRepo: flashcardProgressionRepo,
	}
	testCases := []TestCase{
		{
			name:        "missing study set id",
			ctx:         ctx,
			req:         &bpb.RetrieveFlashCardStudyProgressRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "req must have study set id"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "missing student id",
			ctx:  ctx,
			req: &bpb.RetrieveFlashCardStudyProgressRequest{
				StudySetId: "study-set-id",
			},
			expectedErr: status.Error(codes.InvalidArgument, "req must have student id"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "missing paging",
			ctx:  ctx,
			req: &bpb.RetrieveFlashCardStudyProgressRequest{
				StudySetId: "study-set-id",
				StudentId:  "student-id",
			},
			expectedErr: status.Error(codes.InvalidArgument, "req must have paging field"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "offset must be positive",
			ctx:  ctx,
			req: &bpb.RetrieveFlashCardStudyProgressRequest{
				StudySetId: "study-set-id",
				StudentId:  "student-id",
				Paging:     &cpb.Paging{},
			},
			expectedErr: status.Error(codes.InvalidArgument, "offset must be positive"),
			setup:       func(ctx context.Context) {},
		},
		{
			name: "ErrNoRows when get FlashcardProgression",
			ctx:  ctx,
			req: &bpb.RetrieveFlashCardStudyProgressRequest{
				StudySetId: "study-set-id",
				StudentId:  "student-id",
				Paging: &cpb.Paging{
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, "getFlashcardProgressionWithPaging.FlashcardProgressionRepo.Get: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				flashcardProgressionRepo.On("Get", ctx, s.DB, mock.Anything).Once().Return(nil, pgx.ErrNoRows)
			},
		},
		{
			name: "ErrNoRows when get Quizzed by externalIDs",
			ctx:  ctx,
			req: &bpb.RetrieveFlashCardStudyProgressRequest{
				StudySetId: "study-set-id",
				StudentId:  "student-id",
				Paging: &cpb.Paging{
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
			},
			expectedErr: status.Errorf(codes.Internal, "getFlashcardProgressionWithPaging.QuizRepo.GetByExternalIDsAndLmID: %v", pgx.ErrNoRows),
			setup: func(ctx context.Context) {
				flashcardProgressionRepo.On("Get", ctx, s.DB, mock.Anything).Once().Return(&entities.FlashcardProgression{
					LoID:            database.Text("lo-id"),
					QuizExternalIDs: database.TextArray([]string{"quiz-id-1", "quiz-id-2"}),
				}, nil)
				quizRepo.On("GetByExternalIDsAndLmID", ctx, s.DB, database.TextArray([]string{"quiz-id-1", "quiz-id-2"}), database.Text("lo-id")).Once().Return(entities.Quizzes{}, pgx.ErrNoRows)
			},
		},
		{
			name: "happy case",
			ctx:  ctx,
			req: &bpb.RetrieveFlashCardStudyProgressRequest{
				StudySetId: "study-set-id",
				StudentId:  "student-id",
				Paging: &cpb.Paging{
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 1,
					},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				flashcardProgressionRepo.On("Get", ctx, s.DB, mock.Anything).Once().Return(&entities.FlashcardProgression{
					LoID:            database.Text("lo-id"),
					QuizExternalIDs: database.TextArray([]string{"quiz-id-1", "quiz-id-2"}),
				}, nil)
				quizRepo.On("GetByExternalIDsAndLmID", ctx, s.DB, database.TextArray([]string{"quiz-id-1", "quiz-id-2"}), database.Text("lo-id")).Once().Return(entities.Quizzes{}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.RetrieveFlashCardStudyProgress(testCase.ctx, testCase.req.(*bpb.RetrieveFlashCardStudyProgressRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func TestCourseReaderService_RetrieveLastFlashCardStudyProgress(t *testing.T) {
}

func TestToQuizCore(t *testing.T) {
	t.Parallel()
	quiz := generateQuiz()

	quizPb, err := toQuizCore(&quiz)
	assert.Nil(t, err)
	assert.Equal(t, quizPb.Info.Id, quiz.ID.String)
	assert.Equal(t, quizPb.ExternalId, quiz.ExternalID.String)
	assert.Equal(t, quiz.Country.String, cpb.Country_name[int32(quizPb.Info.Country.Number())])
	assert.Equal(t, quizPb.Info.SchoolId, quiz.SchoolID.Int)
	assert.Equal(t, quiz.Kind.String, cpb.QuizType_name[int32(quizPb.Kind.Number())])
	assert.Equal(t, quizPb.Point.Value, quiz.Point.Int)

	assert.Equal(t, len(quiz.TaggedLOs.Elements), len(quizPb.TaggedLos))
	for i := range quiz.TaggedLOs.Elements {
		assert.Equal(t, quiz.TaggedLOs.Elements[i].String, quizPb.TaggedLos[i])
	}

	question := &entities.RichText{}
	err = quiz.Question.AssignTo(question)
	assert.Nil(t, err)
	assert.Equal(t, question.Raw, quizPb.Question.Raw)
	assert.Equal(t, question.RenderedURL, quizPb.Question.Rendered)

	explanation := &entities.RichText{}
	err = quiz.Explanation.AssignTo(explanation)
	assert.Nil(t, err)
	assert.Equal(t, explanation.Raw, quizPb.Explanation.Raw)
	assert.Equal(t, explanation.RenderedURL, quizPb.Explanation.Rendered)

	options := []*entities.QuizOption{}
	err = quiz.Options.AssignTo(&options)
	assert.Nil(t, err)

	for i, opt := range options {
		assert.Equal(t, opt.Content.Raw, quizPb.Options[i].Content.Raw)
		assert.Equal(t, opt.Content.RenderedURL, quizPb.Options[i].Content.Rendered)
		assert.Equal(t, opt.Correctness, quizPb.Options[i].Correctness)
		assert.Equal(t, opt.Label, quizPb.Options[i].Label)
	}

	assert.True(t, quiz.CreatedAt.Time.Equal(quizPb.Info.CreatedAt.AsTime()))
	assert.True(t, quiz.UpdatedAt.Time.Equal(quizPb.Info.UpdatedAt.AsTime()))
	assert.True(t, quiz.DeletedAt.Time.Equal(quizPb.Info.DeletedAt.AsTime()))
}

func generateQuiz() entities.Quiz {
	quiz := entities.Quiz{}

	quiz.ID.Set(idutil.ULIDNow())
	quiz.ExternalID.Set("external-id-1")
	quiz.Country.Set(cpb.Country_name[int32(cpb.Country_COUNTRY_VN)])
	quiz.SchoolID.Set(1)
	quiz.Kind.Set(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MCQ)])
	quiz.Question.Set(`{"raw":"{\"blocks\":[{\"key\":\"2bsgi\",\"text\":\"qeqweqewq\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}","rendered_url":"https://storage.googleapis.com/stag-manabie-backend/content/150cb1b73bc9d3bbe4011a55476a6913.html"}`)
	quiz.Explanation.Set(`{"raw":"{\"blocks\":[{\"key\":\"f5lms\",\"text\":\"\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}","rendered_url":"https://storage.googleapis.com/stag-manabie-backend/content/24061416a35eb51f403307148c5f4cef.html"}`)
	quiz.Options.Set(`[
		{"key":"1", "label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"qwewqeqweqwe\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": false},

		{"key":"2", "label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"hello\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"key":"3", "label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true}
	]`)
	quiz.TaggedLOs.Set([]string{"tagLO1", "tagLO2", "tagLO3"})
	quiz.CreatedBy.Set("adminid")
	quiz.ApprovedBy.Set("adminid")
	quiz.Status.Set("approved")
	quiz.DifficultLevel.Set(2)
	createdAt := database.Timestamptz(time.Now())
	updatedAt := database.Timestamptz(time.Now())
	quiz.CreatedAt.Set(createdAt)
	quiz.UpdatedAt.Set(updatedAt)
	quiz.Point.Set(10)
	return quiz
}
