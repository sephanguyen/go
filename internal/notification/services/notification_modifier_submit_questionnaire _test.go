package services

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	"github.com/manabie-com/backend/internal/yasuo/configurations"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

func TestNotificationModifierService_SubmitQuestionnaire(t *testing.T) {
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}
	config := &configurations.Config{
		Storage: configs.StorageConfig{
			Endpoint: "endpoint",
			Bucket:   "testBucket",
		},
	}
	userID := idutil.ULIDNow()

	userInfoNotificationID := idutil.ULIDNow()
	notificationID := idutil.ULIDNow()
	questionnaireID := idutil.ULIDNow()
	questionnaireSampleProto := utils.GenSampleQuestionnaire()
	answersQuestionnaireProto := utils.GenSampleAnswersForQuestionnaire(questionnaireSampleProto)

	reqSubmit := &npb.SubmitQuestionnaireRequest{
		UserInfoNotificationId: userInfoNotificationID,
		QuestionnaireId:        questionnaireID,
		Answers:                answersQuestionnaireProto,
	}

	userInfoNotificationRepo := &mock_repositories.MockUsersInfoNotificationRepo{}
	infoNotificationRepo := &mock_repositories.MockInfoNotificationRepo{}
	questionnaireRepo := &mock_repositories.MockQuestionnaireRepo{}
	questionnaireQuestionRepo := &mock_repositories.MockQuestionnaireQuestionRepo{}
	questionnaireUserAnswerRepo := &mock_repositories.MockQuestionnaireUserAnswerRepo{}

	svc := &NotificationModifierService{
		DB:                        mockDB,
		StorageConfig:             config.Storage,
		InfoNotificationRepo:      infoNotificationRepo,
		UserNotificationRepo:      userInfoNotificationRepo,
		QuestionnaireRepo:         questionnaireRepo,
		QuestionnaireQuestionRepo: questionnaireQuestionRepo,
		QuestionnaireUserAnswer:   questionnaireUserAnswerRepo,
	}

	testCases := []struct {
		Name      string
		Answers   []*cpb.Answer
		ReqSumbit *npb.SubmitQuestionnaireRequest
		Err       error
		Setup     func(ctx context.Context)
	}{
		{
			Name:      "happy case",
			Err:       nil,
			ReqSumbit: reqSubmit,
			Setup: func(ctx context.Context) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)

				userInfoNotification := &entities.UserInfoNotification{
					UserNotificationID: database.Text(userInfoNotificationID),
					NotificationID:     database.Text(notificationID),
				}
				userNotificationFilter := repositories.NewFindUserNotificationFilter()
				userNotificationFilter.UserNotificationIDs = database.TextArray([]string{userInfoNotificationID})
				userNotificationFilter.UserIDs = database.TextArray([]string{userID})
				userInfoNotificationRepo.On("Find", ctx, mockDB, userNotificationFilter).Once().Return(entities.UserInfoNotifications{userInfoNotification}, nil)

				infoNotification := entities.InfoNotifications{&entities.InfoNotification{
					NotificationID:  database.Text(notificationID),
					QuestionnaireID: database.Text(questionnaireID),
				}}
				filterInfoNotification := repositories.NewFindNotificationFilter()
				filterInfoNotification.Status.Set(nil)
				filterInfoNotification.NotiIDs.Set([]string{notificationID})
				infoNotificationRepo.On("Find", ctx, mockDB, filterInfoNotification).Once().Return(infoNotification, nil)

				questionnaire, _ := mappers.PbToQuestionnaireEnt(questionnaireSampleProto)
				questionnaire.QuestionnaireID.Set(questionnaireID)
				questionnaireRepo.On("FindByID", ctx, mockDB, questionnaireID).Once().Return(questionnaire, nil)

				quesionnaireQuestion, _ := mappers.PbToQuestionnaireQuestionEnts(questionnaireSampleProto)
				questionnaireRepo.On("FindQuestionsByQnID", ctx, mockDB, questionnaireID).Once().Return(quesionnaireQuestion, nil)

				filterUserAnswers := repositories.NewFindUserAnswersFilter()
				filterUserAnswers.UserNotificationIDs = database.TextArray([]string{userInfoNotificationID})
				questionnaireRepo.On("FindUserAnswers", ctx, mockDB, &filterUserAnswers).Once().Return(entities.QuestionnaireUserAnswers{}, nil)

				answersQuestionnaire, _ := mappers.PbToQuestionnaireUserAnswerEnts(userID, "", reqSubmit)
				for _, qnAnswer := range answersQuestionnaire {
					qnAnswer.SubmittedAt.Set(time.Now())
				}

				questionnaireUserAnswerRepo.On("BulkUpsert", ctx, mockTx, mock.MatchedBy(func(actualAnswers entities.QuestionnaireUserAnswers) bool {
					countEqual := 0
					for _, expectedAnswer := range answersQuestionnaire {
						for _, actualAnswer := range actualAnswers {
							if expectedAnswer.QuestionnaireQuestionID == actualAnswer.QuestionnaireQuestionID &&
								expectedAnswer.Answer == actualAnswer.Answer &&
								expectedAnswer.TargetID == actualAnswer.TargetID &&
								expectedAnswer.UserID == actualAnswer.UserID &&
								math.Abs((expectedAnswer.SubmittedAt.Time.Sub(actualAnswer.SubmittedAt.Time)).Seconds()) < 60 {
								countEqual++
							}
						}
					}
					return countEqual == len(answersQuestionnaire)
				})).Once().Return(nil)

				userInfoNotificationRepo.On("SetQuestionnareStatusAndSubmittedAt", ctx, mockTx, userInfoNotificationID, cpb.UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED.String(), mock.MatchedBy(func(submittedAt pgtype.Timestamptz) bool {
					timeNow := time.Now()
					return math.Abs((timeNow.Sub(submittedAt.Time)).Seconds()) < 60
				})).Once().Return(nil)
			},
		},
	}

	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx = interceptors.ContextWithUserID(ctx, userID)
			ctx = metadata.AppendToOutgoingContext(ctx, "pkg", "manabie", "version", "1.0.0", "token", idutil.ULIDNow())
			testCase.Setup(ctx)
			_, err := svc.SubmitQuestionnaire(ctx, testCase.ReqSumbit)
			if testCase.Err != nil {
				assert.ErrorIs(t, err, testCase.Err)
			}
		})
	}
}
