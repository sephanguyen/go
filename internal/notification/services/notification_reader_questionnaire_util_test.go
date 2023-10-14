package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotificationReader_findAttachedQnDetail(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	qnRepo := &mock_repositories.MockQuestionnaireRepo{}

	svc := &NotificationReaderService{
		DB:                db,
		QuestionnaireRepo: qnRepo,
	}
	userID := "parent_1"
	targetID := "student_1"

	testCases := []struct {
		name        string
		qnID        string
		userNotiID  string
		isSubmitted bool
		err         error
		res         *cpb.UserQuestionnaire
		setup       func(ctx context.Context) *cpb.UserQuestionnaire
	}{
		{
			name:        "submitted",
			qnID:        "qn_1",
			userNotiID:  "noti_1",
			isSubmitted: true,
			setup: func(ctx context.Context) *cpb.UserQuestionnaire {
				qn := utils.GenQuestionaire()
				qnqs := utils.GenQNQuestions(qn.QuestionnaireID.String)
				anws := utils.GenQNUserAnswers(userID, targetID)
				qnRepo.On("FindByID", ctx, db, "qn_1").Once().Return(&qn, nil)
				qnRepo.On("FindQuestionsByQnID", ctx, db, "qn_1").Once().Return(qnqs, nil)

				filterUserAnswers := repositories.NewFindUserAnswersFilter()
				filterUserAnswers.UserNotificationIDs = database.TextArray([]string{"noti_1"})
				qnRepo.On("FindUserAnswers", ctx, db, &filterUserAnswers).Once().Return(anws, nil)
				return mappers.QNUserAnswerToPb(&qn, qnqs, anws, true)
			},
		},
		{
			name:        "not submitted",
			qnID:        "qn_2",
			userNotiID:  "noti_2",
			isSubmitted: false,
			setup: func(ctx context.Context) *cpb.UserQuestionnaire {
				qn := utils.GenQuestionaire()
				qnqs := utils.GenQNQuestions(qn.QuestionnaireID.String)
				qnRepo.On("FindByID", ctx, db, "qn_2").Once().Return(&qn, nil)
				qnRepo.On("FindQuestionsByQnID", ctx, db, "qn_2").Once().Return(qnqs, nil)
				return mappers.QNUserAnswerToPb(&qn, qnqs, nil, false)
			},
		},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			expectRes := c.setup(ctx)
			res, err := svc.findAttachedQuestionnaireDetail(ctx, c.qnID, c.userNotiID, c.isSubmitted)
			assert.Equal(t, c.err, err)
			assert.Equal(t, expectRes, res)
		})
	}
	mock.AssertExpectationsForObjects(t, qnRepo)
}
