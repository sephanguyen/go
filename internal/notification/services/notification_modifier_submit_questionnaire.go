package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	"github.com/manabie-com/backend/internal/notification/services/mappers"
	"github.com/manabie-com/backend/internal/notification/services/validation"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *NotificationModifierService) SubmitQuestionnaire(ctx context.Context, req *npb.SubmitQuestionnaireRequest) (*npb.SubmitQuestionnaireResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)

	// Get user notification
	userNotificationFilter := repositories.NewFindUserNotificationFilter()
	userNotificationFilter.UserNotificationIDs = database.TextArray([]string{req.UserInfoNotificationId})
	userNotificationFilter.UserIDs = database.TextArray([]string{userID})
	userNotifications, err := svc.UserNotificationRepo.Find(ctx, svc.DB, userNotificationFilter)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error on svc.FindUserNotifications: %v", err))
	}
	if len(userNotifications) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user notification doesn't exist")
	}
	userNotification := userNotifications[0]

	// Get notification from user notification
	infoNotification, err := svc.findNotificationByID(ctx, svc.DB, userNotification.NotificationID.String)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error on FindNotificationByID: %v", err))
	}
	if infoNotification.QuestionnaireID.String != req.QuestionnaireId {
		return nil, status.Error(codes.InvalidArgument, "invalid questionnaire_id")
	}

	// Get questionnanire
	questionnaire, err := svc.QuestionnaireRepo.FindByID(ctx, svc.DB, req.QuestionnaireId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error on QuestionnaireRepo.FindByID: %v", err))
	}
	if questionnaire.ExpirationDate.Time.Before(time.Now()) {
		return nil, status.Error(codes.InvalidArgument, "expired questionnaire, you cannot submit questionnaire after the expiration date has passed")
	}

	// Get questionnanire questions
	questionnaireQuestions, err := svc.QuestionnaireRepo.FindQuestionsByQnID(ctx, svc.DB, req.QuestionnaireId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error on QuestionnaireQuestionRepo.FindQuestionsByQnID: %v", err))
	}

	// Get old answers
	filterUserAnswers := repositories.NewFindUserAnswersFilter()
	filterUserAnswers.UserNotificationIDs = database.TextArray([]string{req.UserInfoNotificationId})
	oldUserAnswerQuestionnaires, err := svc.QuestionnaireRepo.FindUserAnswers(ctx, svc.DB, &filterUserAnswers)
	if err != nil {
		return nil, status.Error(codes.Internal, "an error occurred when finding old user answer submitting: "+err.Error())
	}

	// Validation zone
	err = validation.ValidateSubmitQuestionnaire(req, questionnaire, questionnaireQuestions, oldUserAnswerQuestionnaires)
	if err != nil {
		return nil, err
	}

	targetID := userNotification.StudentID.String
	if targetID == "" {
		targetID = userNotification.ParentID.String
	}
	newUserQuestionnaireAnswers, err := mappers.PbToQuestionnaireUserAnswerEnts(userID, targetID, req)
	if err != nil {
		return nil, status.Error(codes.Internal, "can not convert toQuestionnaireUserAnswersEnt")
	}

	err = database.ExecInTxWithRetry(ctx, svc.DB, func(ctx context.Context, tx pgx.Tx) error {
		return svc.submitQuestionnaire(ctx, tx, req.UserInfoNotificationId, oldUserAnswerQuestionnaires, newUserQuestionnaireAnswers)
	})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error on ExecInTxWithRetry: %v", err))
	}

	return &npb.SubmitQuestionnaireResponse{}, nil
}

func (svc *NotificationModifierService) submitQuestionnaire(ctx context.Context, tx pgx.Tx, userNotificationID string, oldQnUserAnswers entities.QuestionnaireUserAnswers, qnUserAnswers entities.QuestionnaireUserAnswers) error {
	submittedAt := time.Now()

	// soft delete old submiting if exist
	if len(oldQnUserAnswers) > 0 {
		answerIds := make([]string, 0)

		for _, oldAnswer := range oldQnUserAnswers {
			answerIds = append(answerIds, oldAnswer.AnswerID.String)
		}

		err := svc.QuestionnaireUserAnswer.SoftDelete(ctx, tx, answerIds)
		if err != nil {
			return fmt.Errorf("svc.QuestionnaireUserAnswer.SoftDelete: %v", err)
		}
	}

	for _, qnAnswer := range qnUserAnswers {
		_ = qnAnswer.SubmittedAt.Set(submittedAt)
	}

	// upsert new questionnaire user answer
	err := svc.QuestionnaireUserAnswer.BulkUpsert(ctx, tx, qnUserAnswers)
	if err != nil {
		return fmt.Errorf("svc.QuestionnaireUserAnswer.BulkUpsert: %v", err)
	}

	// set user notification to answered
	err = svc.UserNotificationRepo.SetQuestionnareStatusAndSubmittedAt(ctx, tx, userNotificationID, cpb.UserNotificationQuestionnaireStatus_USER_NOTIFICATION_QUESTIONNAIRE_STATUS_ANSWERED.String(), database.Timestamptz(submittedAt))
	if err != nil {
		return fmt.Errorf("svc.UserNotificationRepo.SetQuestionnareStatusAndSubmittedAt: %v", err)
	}

	return nil
}
