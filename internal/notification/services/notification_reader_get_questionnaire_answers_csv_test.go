package services

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNotificationModifierService_GetQuestionnaireAnswersCSV(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}

	qnRepo := &mock_repositories.MockQuestionnaireRepo{}
	svc := &NotificationReaderService{
		DB:                db,
		QuestionnaireRepo: qnRepo,
	}

	questionnaireID := idutil.ULIDNow()
	userIDSearched := idutil.ULIDNow()
	targetIDAnswered := idutil.ULIDNow()
	questionnaireQuestions := utils.GenQNQuestions(questionnaireID)

	testCases := []struct {
		Name  string
		Req   *npb.GetQuestionnaireAnswersCSVRequest
		Err   error
		Setup func(ctx context.Context)
	}{
		{
			Name: "success case",
			Req: &npb.GetQuestionnaireAnswersCSVRequest{
				QuestionnaireId: questionnaireID,
				Timezone:        "Asia/Ho_Chi_Minh",
				Language:        "en",
			},
			Setup: func(ctx context.Context) {
				qnRepo.On("FindQuestionsByQnID", ctx, db, questionnaireID).Once().Return(questionnaireQuestions, nil)

				questionnaireQuestionIDs := make([]string, 0)
				for _, q := range questionnaireQuestions {
					questionnaireQuestionIDs = append(questionnaireQuestionIDs, q.QuestionnaireQuestionID.String)
				}

				qnRepo.On("FindQuestionnaireCSVResponders", ctx, db, questionnaireID).Once().Return([]*repositories.QuestionnaireCSVResponder{
					{
						UserID:   database.Text(userIDSearched),
						TargetID: database.Text(targetIDAnswered),
						Name:     database.Text(idutil.ULIDNow()),
					},
				}, nil)

				findUserAnswersFilter := repositories.NewFindUserAnswersFilter()
				findUserAnswersFilter.QuestionnaireQuestionIDs = database.TextArray(questionnaireQuestionIDs)
				findUserAnswersFilter.UserIDs = database.TextArray([]string{userIDSearched})
				findUserAnswersFilter.TargetIDs = database.TextArray([]string{targetIDAnswered})

				questionnaireUserAnswers := entities.QuestionnaireUserAnswers{}
				for i := 0; i < 3; i++ {
					answer := utils.GenQNUserAnswer(userIDSearched, targetIDAnswered, idutil.ULIDNow())
					questionnaireUserAnswers = append(questionnaireUserAnswers, &answer)
				}
				qnRepo.On("FindUserAnswers", ctx, db, &findUserAnswersFilter).Once().Return(questionnaireUserAnswers, nil)
			},
		},
		{
			Name: "doesn't find any questions with your questionnaire_id",
			Req: &npb.GetQuestionnaireAnswersCSVRequest{
				QuestionnaireId: questionnaireID,
				Timezone:        "Asia/Ho_Chi_Minh",
				Language:        "en",
			},
			Err: status.Error(codes.InvalidArgument, "No questions found with questionnaireId: "+questionnaireID),
			Setup: func(ctx context.Context) {
				qnRepo.On("FindQuestionsByQnID", ctx, db, questionnaireID).Once().Return(entities.QuestionnaireQuestions{}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			ctx = interceptors.ContextWithUserID(ctx, mock.Anything)
			ctx = metadata.AppendToOutgoingContext(ctx, "pkg", "manabie", "version", "1.0.0", "token", idutil.ULIDNow())
			testCase.Setup(ctx)
			_, err := svc.GetQuestionnaireAnswersCSV(ctx, testCase.Req)
			if testCase.Err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.Err, err)
			}
		})
	}
}

func TestNotificationModifierService_ExportCSVRespondersToCSVArrayData(t *testing.T) {
	t.Parallel()

	svc := &NotificationReaderService{}

	timeSubmitted := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	responders := []*repositories.QuestionnaireCSVResponder{
		{
			UserNotificationID: database.Text("user-notification-id-1"),
			UserID:             database.Text("student-id-1"),
			TargetID:           database.Text("student-id-1"),
			StudentID:          database.Text("student-id-1"),
			Name:               database.Text("student-name-1"),
			TargetName:         database.Text("student-name-1"),
			SubmittedAt:        database.Timestamptz(timeSubmitted),
			IsParent:           database.Bool(false),
			IsIndividual:       database.Bool(false),
			StudentExternalID:  database.Text("student-external-id-1"),
			LocationNames:      database.TextArray([]string{"location-1"}),
		},
		{
			UserNotificationID: database.Text("user-notification-id-2"),
			UserID:             database.Text("parent-id-1"),
			TargetID:           database.Text("student-id-1"),
			StudentID:          database.Text("student-id-1"),
			Name:               database.Text("parent-name-1"),
			TargetName:         database.Text("student-name-1"),
			SubmittedAt:        database.Timestamptz(timeSubmitted),
			StudentExternalID:  database.Text("student-external-id-1"),
			IsParent:           database.Bool(true),
			IsIndividual:       database.Bool(false),
			LocationNames:      database.TextArray([]string{"location-1", "location-2"}),
		},
		{
			UserNotificationID: database.Text("user-notification-id-3"),
			UserID:             database.Text("parent-id-2"),
			TargetID:           database.Text("parent-id-2"),
			StudentID:          database.Text(""),
			Name:               database.Text("parent-name-2"),
			TargetName:         database.Text("parent-name-2"),
			SubmittedAt:        database.Timestamptz(timeSubmitted),
			StudentExternalID:  database.Text(""),
			IsParent:           database.Bool(true),
			IsIndividual:       database.Bool(true),
			LocationNames:      database.TextArray([]string{"location-3"}),
		},
	}
	questionnaireUserAnswers := entities.QuestionnaireUserAnswers{}
	questionnaireQuestions := utils.GenQNQuestions(idutil.ULIDNow())

	freeTextAnswer := "Free text answer."

	csvArrayDataExpected := [][]string{
		{"1", "2023/01/01, 07:00:00", "location-1", "student-name-1", "", "student-id-1", "student-external-id-1", "A", "B", "Free text answer."},
		{"2", "2023/01/01, 07:00:00", "location-1, location-2", "parent-name-1", "student-name-1", "student-id-1", "student-external-id-1", "A", "B", "Free text answer."},
		{"3", "2023/01/01, 07:00:00", "location-3", "parent-name-2", "", "", "", "A", "B", "Free text answer."},
	}

	clientLocation, _ := time.LoadLocation("Asia/Ho_Chi_Minh")

	t.Run("happy case", func(t *testing.T) {
		for _, responder := range responders {
			for _, questionnaireQuestion := range questionnaireQuestions {
				answer := utils.GenQNUserAnswer(responder.UserID.String, responder.TargetID.String, questionnaireQuestion.QuestionnaireQuestionID.String)
				switch questionnaireQuestion.Type.String {
				case cpb.QuestionType_QUESTION_TYPE_CHECK_BOX.String():
					//answer is: A
					answer.Answer = database.Text(database.FromTextArray(questionnaireQuestion.Choices)[0])
				case cpb.QuestionType_QUESTION_TYPE_MULTIPLE_CHOICE.String():
					//answer is: B
					answer.Answer = database.Text(database.FromTextArray(questionnaireQuestion.Choices)[1])
				case cpb.QuestionType_QUESTION_TYPE_FREE_TEXT.String():
					//answer is: Free text answer.
					answer.Answer = database.Text(freeTextAnswer)
				}
				questionnaireUserAnswers = append(questionnaireUserAnswers, &answer)
			}
		}

		userAnswersCSV := mappers.QuestionnaireUserAnswersToExportedCSVResponders(responders, questionnaireUserAnswers, questionnaireQuestions)
		csvArrayData := svc.exportCSVRespondersToCSVArrayData(userAnswersCSV, questionnaireQuestions, clientLocation)

		assert.Equal(t, csvArrayDataExpected, csvArrayData)
	})
}
