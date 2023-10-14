package services

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateQuizTest(t *testing.T) {
}

func TestDuplicateBook(t *testing.T) {
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
		{"label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"qwewqeqweqwe\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": false},

		{"label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"hello\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true},

		{"label":"","configs":[],"content":{"raw": "{\"blocks\":[{\"key\":\"56i3l\",\"text\":\"goodbye\",\"type\":\"unstyled\",\"depth\":0,\"inlineStyleRanges\":[],\"entityRanges\":[],\"data\":{}}],\"entityMap\":{}}", "rendered_url": "https://storage.googleapis.com/stag-manabie-backend/content/0e739b785fd16378cebccb8fba0f6586.html"}, "correctness": true}
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

	return quiz
}

func TestCourseModifierService_CreateOfflineLearningRecords(t *testing.T) {
}

func TestMultipleChoiceQuizType_CheckCorrectness(t *testing.T) {
	t.Parallel()
	mockDB := &mock_database.Ext{}
	shuffledQuizSetRepo := &mock_repositories.MockShuffledQuizSetRepo{}
	quiz := generateQuiz()
	multipleChoiceQuiz := &MultipleChoiceQuiz{
		Quiz:                &quiz,
		ShuffledQuizSetRepo: shuffledQuizSetRepo,
	}
	testCases := []struct {
		Name        string
		Request     []*bpb.Answer
		ResponseErr error
		Setup       func(ctx context.Context)
	}{
		{
			Name:        "happy case",
			Request:     make([]*bpb.Answer, 0),
			ResponseErr: nil,
			Setup: func(ctx context.Context) {
				seed := database.Text(strconv.FormatInt(time.Now().UnixNano(), 10))
				idx := database.Int4(1)
				shuffledQuizSetRepo.On("GetSeed", ctx, mock.Anything, mock.Anything).Once().Return(seed, nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(idx, nil)
			},
		},
		{
			Name: "happy case 2",
			Request: []*bpb.Answer{
				{
					Format: &bpb.Answer_SelectedIndex{
						SelectedIndex: 1,
					},
				},
				{
					Format: &bpb.Answer_SelectedIndex{
						SelectedIndex: 2,
					},
				},
				{
					Format: &bpb.Answer_SelectedIndex{
						SelectedIndex: 3,
					},
				},
			},
			ResponseErr: nil,
			Setup: func(ctx context.Context) {
				seed := database.Text(strconv.FormatInt(time.Now().UnixNano(), 10))
				idx := database.Int4(1)
				shuffledQuizSetRepo.On("GetSeed", ctx, mock.Anything, mock.Anything).Once().Return(seed, nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(idx, nil)
			},
		},
		{
			Name: "error not received right format",
			Request: []*bpb.Answer{
				{
					Format: &bpb.Answer_SelectedIndex{
						SelectedIndex: 1,
					},
				},
				{
					Format: &bpb.Answer_FilledText{
						FilledText: "some answer text",
					},
				},
				{
					Format: &bpb.Answer_SelectedIndex{
						SelectedIndex: 3,
					},
				},
			},
			ResponseErr: fmt.Errorf("your answer is not the multiple choice type"),
			Setup: func(ctx context.Context) {
				seed := database.Text(strconv.FormatInt(time.Now().UnixNano(), 10))
				idx := database.Int4(1)
				shuffledQuizSetRepo.On("GetSeed", ctx, mock.Anything, mock.Anything).Once().Return(seed, nil)
				shuffledQuizSetRepo.On("GetQuizIdx", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(idx, nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := multipleChoiceQuiz.CheckCorrectness(ctx, mockDB, testCase.Request)
			if testCase.ResponseErr != nil {
				assert.EqualError(t, err, testCase.ResponseErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestManualInputQuizType_CheckCorrectness(t *testing.T) {
	t.Parallel()
	mockDB := &mock_database.Ext{}
	quiz := generateQuiz()
	manualInputQuizType := &ManualInputQuiz{
		Quiz: &quiz,
	}
	testCases := []struct {
		Name        string
		Request     []*bpb.Answer
		ResponseErr error
		Setup       func(ctx context.Context)
	}{
		{
			Name:        "happy case",
			Request:     make([]*bpb.Answer, 0),
			ResponseErr: nil,
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "happy case user see they do quiz correctly",
			Request: []*bpb.Answer{
				{
					Format: &bpb.Answer_SelectedIndex{
						SelectedIndex: 1,
					},
				},
			},
			ResponseErr: nil,
			Setup: func(ctx context.Context) {
				// user manualy do quiz on paper
				// SelectedIndex == 1 => user see they do quiz correctly
				// SelectedIndex == 2 => user see they do quiz incorrectly
			},
		},
		{
			Name: "happy case user see they do quiz incorrectly",
			Request: []*bpb.Answer{
				{
					Format: &bpb.Answer_SelectedIndex{
						SelectedIndex: 2,
					},
				},
			},
			ResponseErr: nil,
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "error when user answer not selected index format",
			Request: []*bpb.Answer{
				{
					Format: &bpb.Answer_FilledText{
						FilledText: "some answer text",
					},
				},
			},
			ResponseErr: fmt.Errorf("your answer is not the multiple choice type"),
			Setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := manualInputQuizType.CheckCorrectness(ctx, mockDB, testCase.Request)
			if testCase.ResponseErr != nil {
				assert.EqualError(t, err, testCase.ResponseErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestFillInTheBlankQuizType_CheckCorrectness(t *testing.T) {
	t.Parallel()
	mockDB := &mock_database.Ext{}
	quiz := generateQuiz()
	fillInTheBlankQuiz := &FillInTheBlankQuiz{
		Quiz: &quiz,
	}
	testCases := []struct {
		Name        string
		Request     []*bpb.Answer
		ResponseErr error
		Setup       func(ctx context.Context)
	}{
		{
			Name:        "happy case",
			Request:     make([]*bpb.Answer, 0),
			ResponseErr: nil,
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "happy case",
			Request: []*bpb.Answer{
				{
					Format: &bpb.Answer_FilledText{
						FilledText: "hello",
					},
				},
				{
					Format: &bpb.Answer_FilledText{
						FilledText: "goodbye",
					},
				},
				{
					Format: &bpb.Answer_FilledText{
						FilledText: "bonjour",
					},
				},
			},
			ResponseErr: nil,
			Setup: func(ctx context.Context) {
			},
		},
		{
			Name: "err when user answer not right format",
			Request: []*bpb.Answer{
				{
					Format: &bpb.Answer_SelectedIndex{
						SelectedIndex: 1,
					},
				},
				{
					Format: &bpb.Answer_FilledText{
						FilledText: "goodbye",
					},
				},
				{
					Format: &bpb.Answer_FilledText{
						FilledText: "bonjour",
					},
				},
			},
			ResponseErr: fmt.Errorf("your answer is not the fill in the blank type"),
			Setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.Setup(ctx)
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := fillInTheBlankQuiz.CheckCorrectness(ctx, mockDB, testCase.Request)
			if testCase.ResponseErr != nil {
				assert.EqualError(t, err, testCase.ResponseErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
