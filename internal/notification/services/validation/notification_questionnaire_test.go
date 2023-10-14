package validation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	utils "github.com/manabie-com/backend/internal/notification/services/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_ValidateSubmitQuestionnaire(t *testing.T) {
	// All questionnaire question generated is required
	questionnaireSamplePb := utils.GenSampleQuestionnaire()
	questionnaireEnt, err := mappers.PbToQuestionnaireEnt(questionnaireSamplePb)
	assert.NoError(t, err)

	quesionnaireQuestion, _ := mappers.PbToQuestionnaireQuestionEnts(questionnaireSamplePb)

	type testCase struct {
		Name                        string
		Questionnaire               *entities.Questionnaire
		QuestionnaireQuestions      entities.QuestionnaireQuestions
		OldUserAnswerQuestionnaires entities.QuestionnaireUserAnswers
		Err                         error
		GetReqSubmit                func() *ypb.SubmitQuestionnaireRequest
		ReqSubmit                   *npb.SubmitQuestionnaireRequest
		Setup                       func(ctx context.Context, this *testCase)
	}

	testCases := []testCase{
		{
			Name:                        "invalid answer question",
			Err:                         status.Error(codes.InvalidArgument, "you cannot answer the question not in questionnaire"),
			Questionnaire:               questionnaireEnt,
			QuestionnaireQuestions:      quesionnaireQuestion,
			OldUserAnswerQuestionnaires: nil,
			ReqSubmit: &npb.SubmitQuestionnaireRequest{
				QuestionnaireId: questionnaireSamplePb.QuestionnaireId,
			},
			Setup: func(ctx context.Context, this *testCase) {
				answersQuestionnairePb := utils.GenSampleAnswersForQuestionnaire(questionnaireSamplePb)
				this.ReqSubmit.Answers = append(answersQuestionnairePb, &cpb.Answer{
					// Add another answer for question out of quesionnaire
					QuestionnaireQuestionId: idutil.ULIDNow(),
					Answer:                  idutil.ULIDNow(),
				})
			},
		},
		{
			Name:                        "missing required question",
			Err:                         status.Error(codes.InvalidArgument, "missing required question, you need to fill all required question"),
			Questionnaire:               questionnaireEnt,
			QuestionnaireQuestions:      quesionnaireQuestion,
			OldUserAnswerQuestionnaires: nil,
			ReqSubmit: &npb.SubmitQuestionnaireRequest{
				QuestionnaireId: questionnaireSamplePb.QuestionnaireId,
			},
			Setup: func(ctx context.Context, this *testCase) {
				answersQuestionnairePb := utils.GenSampleAnswersForQuestionnaire(questionnaireSamplePb)

				// Fill empty answer
				qnQuestionRequiredId := ""
				for _, question := range quesionnaireQuestion {
					if question.IsRequired.Bool {
						qnQuestionRequiredId = question.QuestionnaireQuestionID.String
						break
					}
				}

				for _, answerQuestionnairePb := range answersQuestionnairePb {
					if answerQuestionnairePb.QuestionnaireQuestionId == qnQuestionRequiredId {
						answerQuestionnairePb.Answer = ""
					}
				}

				this.ReqSubmit.Answers = answersQuestionnairePb
			},
		},
		{
			Name:                        "missing required question in case empty answer",
			Err:                         status.Error(codes.InvalidArgument, "missing required question, you need to fill all required question"),
			Questionnaire:               questionnaireEnt,
			QuestionnaireQuestions:      quesionnaireQuestion,
			OldUserAnswerQuestionnaires: nil,
			ReqSubmit: &npb.SubmitQuestionnaireRequest{
				QuestionnaireId: questionnaireSamplePb.QuestionnaireId,
			},
			Setup: func(ctx context.Context, this *testCase) {
				answersQuestionnairePb := utils.GenSampleAnswersForQuestionnaire(questionnaireSamplePb)

				// Get index of required question and remove answer for that
				qnQuestionRequiredId := ""
				for _, question := range quesionnaireQuestion {
					if question.IsRequired.Bool {
						qnQuestionRequiredId = question.QuestionnaireQuestionID.String
						break
					}
				}

				indexRequired := 0
				for idx, answerQuestionnairePb := range answersQuestionnairePb {
					if answerQuestionnairePb.QuestionnaireQuestionId == qnQuestionRequiredId {
						indexRequired = idx
					}
				}

				answersQuestionnairePb = append(answersQuestionnairePb[:indexRequired], answersQuestionnairePb[indexRequired+1:]...)
				this.ReqSubmit.Answers = answersQuestionnairePb
			},
		},
		{
			Name:                        "invalid answer type",
			Err:                         status.Error(codes.InvalidArgument, "you cannot have multiple answer for multiple choices and free text question"),
			Questionnaire:               questionnaireEnt,
			QuestionnaireQuestions:      quesionnaireQuestion,
			OldUserAnswerQuestionnaires: nil,
			ReqSubmit: &npb.SubmitQuestionnaireRequest{
				QuestionnaireId: questionnaireSamplePb.QuestionnaireId,
			},
			Setup: func(ctx context.Context, this *testCase) {
				answersQuestionnairePb := utils.GenSampleAnswersForQuestionnaire(questionnaireSamplePb)

				qnQuestionIdMultichoices := ""
				qnQuestionMultichoices := make([]string, 0)
				for _, question := range quesionnaireQuestion {
					if question.Type.String == "QUESTION_TYPE_MULTIPLE_CHOICE" {
						qnQuestionIdMultichoices = question.QuestionnaireQuestionID.String
						qnQuestionMultichoices = database.FromTextArray(question.Choices)
						break
					}
				}
				answersQuestionnairePb = append(answersQuestionnairePb, &cpb.Answer{
					QuestionnaireQuestionId: qnQuestionIdMultichoices,
					Answer:                  qnQuestionMultichoices[1],
				})

				qnQuestionIdFreeText := ""
				for _, question := range quesionnaireQuestion {
					if question.Type.String == "QUESTION_TYPE_FREE_TEXT" {
						qnQuestionIdFreeText = question.QuestionnaireQuestionID.String
						break
					}
				}
				answersQuestionnairePb = append(answersQuestionnairePb, &cpb.Answer{
					QuestionnaireQuestionId: qnQuestionIdFreeText,
					Answer:                  idutil.ULIDNow(),
				})
				this.ReqSubmit.Answers = answersQuestionnairePb
			},
		},
		{
			Name:                        "invalid choice",
			Err:                         status.Error(codes.InvalidArgument, "your answer doesn't in questionnaire question choices"),
			Questionnaire:               questionnaireEnt,
			QuestionnaireQuestions:      quesionnaireQuestion,
			OldUserAnswerQuestionnaires: nil,
			ReqSubmit: &npb.SubmitQuestionnaireRequest{
				QuestionnaireId: questionnaireSamplePb.QuestionnaireId,
			},
			Setup: func(ctx context.Context, this *testCase) {
				answersQuestionnairePb := utils.GenSampleAnswersForQuestionnaire(questionnaireSamplePb)

				qnQuestionIdMultichoices := ""
				qnQuestionRequiredId := ""
				for _, question := range quesionnaireQuestion {
					if question.Type.String == "QUESTION_TYPE_MULTIPLE_CHOICE" {
						qnQuestionIdMultichoices = question.QuestionnaireQuestionID.String
						qnQuestionRequiredId = question.QuestionnaireQuestionID.String
						break
					}
				}

				indexMultichoices := 0
				for idx, answerQuestionnairePb := range answersQuestionnairePb {
					if answerQuestionnairePb.QuestionnaireQuestionId == qnQuestionRequiredId {
						indexMultichoices = idx
					}
				}
				// Remove current multichoices answer
				answersQuestionnairePb = append(answersQuestionnairePb[:indexMultichoices], answersQuestionnairePb[indexMultichoices+1:]...)

				// Add new answer that out of my questionnaire questions
				answersQuestionnairePb = append(answersQuestionnairePb, &cpb.Answer{
					QuestionnaireQuestionId: qnQuestionIdMultichoices,
					Answer:                  idutil.ULIDNow(),
				})

				qnQuestionIdFreeText := ""
				for _, question := range quesionnaireQuestion {
					if question.Type.String == "QUESTION_TYPE_CHECK_BOX" {
						qnQuestionIdFreeText = question.QuestionnaireQuestionID.String
						break
					}
				}
				answersQuestionnairePb = append(answersQuestionnairePb, &cpb.Answer{
					QuestionnaireQuestionId: qnQuestionIdFreeText,
					Answer:                  idutil.ULIDNow(),
				})
				this.ReqSubmit.Answers = answersQuestionnairePb
			},
		},
		{
			Name:                   "invalid re-submit",
			Err:                    status.Error(codes.InvalidArgument, "resubmit not allowed, you cannot re-submit this questionnaire"),
			Questionnaire:          questionnaireEnt,
			QuestionnaireQuestions: quesionnaireQuestion,
			OldUserAnswerQuestionnaires: entities.QuestionnaireUserAnswers{
				{
					AnswerID: database.Text(idutil.ULIDNow()),
				},
			},
			ReqSubmit: &npb.SubmitQuestionnaireRequest{
				QuestionnaireId: questionnaireSamplePb.QuestionnaireId,
				Answers:         utils.GenSampleAnswersForQuestionnaire(questionnaireSamplePb),
			},
			Setup: func(ctx context.Context, this *testCase) {
				this.Questionnaire.ResubmitAllowed = database.Bool(false)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(context.Background(), &testCase)
			err := ValidateSubmitQuestionnaire(testCase.ReqSubmit, testCase.Questionnaire, testCase.QuestionnaireQuestions, testCase.OldUserAnswerQuestionnaires)
			if testCase.Err != nil {
				assert.ErrorIs(t, err, testCase.Err)
			}
		})
	}
}

func Test_ValidateQuestionnairePb(t *testing.T) {
	type testCase struct {
		Name    string
		Req     *ypb.UpsertNotificationRequest
		RespErr error
		Setup   func(ctx context.Context, this *testCase)
	}
	testCases := []*testCase{
		{
			Name: "happy case questionnaire",
			Req: &ypb.UpsertNotificationRequest{
				Notification:  utils.GenSampleNotification(),
				Questionnaire: utils.GenSampleQuestionnaire(),
			},
			RespErr: nil,
			Setup: func(ctx context.Context, this *testCase) {
			},
		},
		{
			Name: "error when questionnaire have null of endtime",
			Req: &ypb.UpsertNotificationRequest{
				Notification:  utils.GenSampleNotification(),
				Questionnaire: utils.GenSampleQuestionnaire(),
			},
			RespErr: fmt.Errorf("you cannot set expiration date of questionnaire is empty"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Questionnaire.ExpirationDate = nil
			},
		},
		{
			Name: "error when questionnaire have endtime in the past",
			Req: &ypb.UpsertNotificationRequest{
				Notification:  utils.GenSampleNotification(),
				Questionnaire: utils.GenSampleQuestionnaire(),
			},
			RespErr: fmt.Errorf("you cannot set expiration date of questionnaire at a time in the past"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Questionnaire.ExpirationDate = timestamppb.New(time.Now().Add(-3 * time.Minute))
			},
		},
		{
			Name: "error when questionnaire have endtime before notification schedule time",
			Req: &ypb.UpsertNotificationRequest{
				Notification:  utils.GenSampleNotification(),
				Questionnaire: utils.GenSampleQuestionnaire(),
			},
			RespErr: fmt.Errorf("you cannot set the expiration date of a questionnaire at a time before notification's schedule time"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Notification.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED
				this.Req.Notification.ScheduledAt = timestamppb.New(time.Now().Add(3 * time.Minute))
				this.Req.Questionnaire.ExpirationDate = timestamppb.New(time.Now().Add(2 * time.Minute))
			},
		},
		{
			Name: "error when questionnaire have null of questions",
			Req: &ypb.UpsertNotificationRequest{
				Notification:  utils.GenSampleNotification(),
				Questionnaire: utils.GenSampleQuestionnaire(),
			},
			RespErr: fmt.Errorf("you cannot upsert a questionnaire without any questions"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Questionnaire.Questions = nil
			},
		},
		{
			Name: "error when questionnaire have empty of questions",
			Req: &ypb.UpsertNotificationRequest{
				Notification:  utils.GenSampleNotification(),
				Questionnaire: utils.GenSampleQuestionnaire(),
			},
			RespErr: fmt.Errorf("you cannot upsert a questionnaire without any questions"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Questionnaire.Questions = []*cpb.Question{}
			},
		},
		{
			Name: "error when questionnaire have empty of question's title",
			Req: &ypb.UpsertNotificationRequest{
				Notification:  utils.GenSampleNotification(),
				Questionnaire: utils.GenSampleQuestionnaire(),
			},
			RespErr: fmt.Errorf("you cannot set question's title is empty"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Questionnaire.Questions[0].Title = ""
			},
		},
		{
			Name: "error when questionnaire question have one of question's choices - multiple choice",
			Req: &ypb.UpsertNotificationRequest{
				Notification:  utils.GenSampleNotification(),
				Questionnaire: utils.GenSampleQuestionnaire(),
			},
			RespErr: fmt.Errorf("question with type is multiple choice or check box must have least two choices"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Questionnaire.Questions[0].Type = cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE
				this.Req.Questionnaire.Questions[0].Choices = []string{idutil.ULIDNow()}
			},
		},
		{
			Name: "error when questionnaire question have one of question's choices - check box",
			Req: &ypb.UpsertNotificationRequest{
				Notification:  utils.GenSampleNotification(),
				Questionnaire: utils.GenSampleQuestionnaire(),
			},
			RespErr: fmt.Errorf("question with type is multiple choice or check box must have least two choices"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Questionnaire.Questions[0].Type = cpb.QuestionType_QUESTION_TYPE_CHECK_BOX
				this.Req.Questionnaire.Questions[0].Choices = []string{idutil.ULIDNow()}
			},
		},
		{
			Name: "error when questionnaire question have question's choices is empty",
			Req: &ypb.UpsertNotificationRequest{
				Notification:  utils.GenSampleNotification(),
				Questionnaire: utils.GenSampleQuestionnaire(),
			},
			RespErr: fmt.Errorf("you cannot set question's choice is empty"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Questionnaire.Questions[0].Type = cpb.QuestionType_QUESTION_TYPE_CHECK_BOX
				this.Req.Questionnaire.Questions[0].Choices = []string{"", ""}
			},
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		tc.Setup(ctx, tc)
		err := ValidateQuestionnairePb(tc.Req.Questionnaire, tc.Req.Notification)
		assert.Equal(t, tc.RespErr, err)
	}
}

func Test_ValidateExportQuestionnaireAnswersCSV(t *testing.T) {
	supportedLangs := []string{"vi", "en", "ja"}
	t.Run("happy case", func(t *testing.T) {
		req := npb.GetQuestionnaireAnswersCSVRequest{
			QuestionnaireId: "questionnaire-id",
			Timezone:        "Asia/Ho_Chi_Minh",
			Language:        "en",
		}

		_, err := ValidateExportQuestionnaireAnswersCSV(supportedLangs, &req)

		assert.Nil(t, err)
	})

	t.Run("failed convert language", func(t *testing.T) {
		req := npb.GetQuestionnaireAnswersCSVRequest{
			QuestionnaireId: "questionnaire-id",
			Timezone:        "Asia/Ho_Chi_Minh",
			Language:        "invalid",
		}

		_, err := ValidateExportQuestionnaireAnswersCSV(supportedLangs, &req)

		assert.EqualError(t, err, fmt.Sprintf("your language: %s is not supported", req.Language))
	})

	t.Run("failed convert timezone", func(t *testing.T) {
		req := npb.GetQuestionnaireAnswersCSVRequest{
			QuestionnaireId: "questionnaire-id",
			Timezone:        "Asia/Ho_Chi_Minh_invalid",
			Language:        "en",
		}

		_, err := ValidateExportQuestionnaireAnswersCSV(supportedLangs, &req)

		assert.EqualError(t, err, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid client timezone: unknown time zone %s", req.Timezone)).Error())
	})
}
