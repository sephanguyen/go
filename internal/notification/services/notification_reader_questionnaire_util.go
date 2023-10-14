package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

func (svc *NotificationReaderService) findAttachedQuestionnaireDetail(ctx context.Context, questionnaireID string, userNotiID string, isSubmitted bool) (*cpb.UserQuestionnaire, error) {
	qn, err := svc.QuestionnaireRepo.FindByID(ctx, svc.DB, questionnaireID)
	if err != nil {
		return nil, fmt.Errorf("QuestionnairRepo.FindByID %w", err)
	}

	questions, err := svc.QuestionnaireRepo.FindQuestionsByQnID(ctx, svc.DB, questionnaireID)
	if err != nil {
		return nil, fmt.Errorf("QuestionnairRepo.FindQuestionsByQnID %w", err)
	}

	var answers entities.QuestionnaireUserAnswers
	if isSubmitted {
		filterUserAnswers := repositories.NewFindUserAnswersFilter()
		filterUserAnswers.UserNotificationIDs = database.TextArray([]string{userNotiID})
		answers, err = svc.QuestionnaireRepo.FindUserAnswers(ctx, svc.DB, &filterUserAnswers)
		if err != nil {
			return nil, fmt.Errorf("QuestionnaireRepo.FindUserAnswers %w", err)
		}
	}
	return mappers.QNUserAnswerToPb(qn, questions, answers, isSubmitted), nil
}
