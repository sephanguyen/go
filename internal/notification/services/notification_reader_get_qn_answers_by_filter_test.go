package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
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

func TestNotificationModifierService_GetAnswersByFilter(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	qnRepo := &mock_repositories.MockQuestionnaireRepo{}

	svc := &NotificationReaderService{
		DB:                  db,
		QuestionnaireRepo:   qnRepo,
	}

	questionnaireID := idutil.ULIDNow()
	userIDSearched := idutil.ULIDNow()
	targetIDAnswered := idutil.ULIDNow()
	questionnaireQuestions := utils.GenQNQuestions(questionnaireID)

	testCases := []struct {
		Name  string
		Req   *npb.GetAnswersByFilterRequest
		Err   error
		Setup func(ctx context.Context)
	}{
		{
			Name: "success case",
			Req: &npb.GetAnswersByFilterRequest{
				QuestionnaireId: questionnaireID,
				Keyword:         "",
				Paging: &cpb.Paging{
					Limit: 100,
				},
			},
			Setup: func(ctx context.Context) {
				qnRepo.On("FindQuestionsByQnID", ctx, db, questionnaireID).Once().Return(questionnaireQuestions, nil)

				questionnaireQuestionIDs := make([]string, 0)
				for _, q := range questionnaireQuestions {
					questionnaireQuestionIDs = append(questionnaireQuestionIDs, q.QuestionnaireQuestionID.String)
				}

				findQuestionnaireRespondersFilter := repositories.NewFindQuestionnaireRespondersFilter()
				findQuestionnaireRespondersFilter.QuestionnaireID = database.Text(questionnaireID)
				findQuestionnaireRespondersFilter.UserName = database.Text("")
				findQuestionnaireRespondersFilter.Offset.Set(0)
				findQuestionnaireRespondersFilter.Limit.Set(100)

				qnRepo.On("FindQuestionnaireResponders", ctx, db, &findQuestionnaireRespondersFilter).Once().Return(uint32(1), []*repositories.QuestionnaireResponder{
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
			Req: &npb.GetAnswersByFilterRequest{
				QuestionnaireId: questionnaireID,
				Keyword:         "",
				Paging: &cpb.Paging{
					Limit: 100,
				},
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
			_, err := svc.GetAnswersByFilter(ctx, testCase.Req)
			if testCase.Err == nil {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, testCase.Err, err)
			}
		})
	}
}
