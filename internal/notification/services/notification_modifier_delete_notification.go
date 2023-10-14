package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *NotificationModifierService) DeleteNotification(ctx context.Context, req *npb.DeleteNotificationRequest) (*npb.DeleteNotificationResponse, error) {
	if req.NotificationId == "" {
		return nil, status.Error(codes.InvalidArgument, "notificationID is required")
	}
	notification, err := svc.findUndeletedNotification(ctx, req.NotificationId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = database.ExecInTxWithRetry(ctx, svc.DB, func(ctx context.Context, tx pgx.Tx) error {
		if notification.QuestionnaireID.String != "" {
			err = svc.QuestionnaireRepo.SoftDelete(ctx, tx, []string{notification.QuestionnaireID.String})
			if err != nil {
				return fmt.Errorf("QuestionnaireRepo.SoftDelete %v", err)
			}

			err = svc.QuestionnaireQuestionRepo.SoftDelete(ctx, tx, []string{notification.QuestionnaireID.String})
			if err != nil {
				return fmt.Errorf("QuestionnaireQuestionRepo.SoftDelete %v", err)
			}

			err = svc.QuestionnaireUserAnswer.SoftDeleteByQuestionnaireID(ctx, tx, []string{notification.QuestionnaireID.String})
			if err != nil {
				return fmt.Errorf("QuestionnaireUserAnswer.SoftDeleteByQuestionnaireID %v", err)
			}
		}

		softDeleteNotificationTagFilter := repositories.NewSoftDeleteNotificationTagFilter()
		_ = softDeleteNotificationTagFilter.NotificationIDs.Set([]string{notification.NotificationID.String})
		err = svc.InfoNotificationTagRepo.SoftDelete(ctx, tx, softDeleteNotificationTagFilter)
		if err != nil {
			return fmt.Errorf("InfoNotificationTagRepo.SoftDelete %v", err)
		}

		err = svc.UserNotificationRepo.SoftDeleteByNotificationID(ctx, tx, notification.NotificationID.String)
		if err != nil {
			return fmt.Errorf("UserNotificationRepo.SoftDeleteByNotificationID %v", err)
		}

		// soft delete regardless of state
		err = svc.InfoNotificationRepo.DiscardNotification(ctx, tx, database.Text(notification.NotificationID.String), database.TextArray(nil))
		if err != nil {
			return fmt.Errorf("InfoNotificationRepo.DiscardNotification %v", err)
		}

		err = svc.InfoNotificationMsgRepo.SoftDelete(ctx, tx, []string{notification.NotificationMsgID.String})
		if err != nil {
			return fmt.Errorf("InfoNotificationMsgRepo.SoftDelete %v", err)
		}

		softDeleteNotificationAccessPathFilter := repositories.NewSoftDeleteNotificationAccessPathFilter()
		_ = softDeleteNotificationAccessPathFilter.NotificationIDs.Set([]string{notification.NotificationID.String})
		err = svc.InfoNotificationAccessPathRepo.SoftDelete(ctx, tx, softDeleteNotificationAccessPathFilter)
		if err != nil {
			return fmt.Errorf("InfoNotificationAccessPathRepo.SoftDelete %v", err)
		}

		targetGroup := &entities.InfoNotificationTarget{}
		err = notification.TargetGroups.AssignTo(targetGroup)
		if err != nil {
			return fmt.Errorf("error to assign target group: %v", err)
		}

		if targetGroup.LocationFilter.Type == consts.TargetGroupSelectTypeList.String() {
			err := svc.NotificationLocationFilterRepo.SoftDeleteByNotificationID(ctx, tx, notification.NotificationID.String)
			if err != nil {
				return fmt.Errorf("error when deleting location filter: %v", err)
			}
		}
		if targetGroup.CourseFilter.Type == consts.TargetGroupSelectTypeList.String() {
			err := svc.NotificationCourseFilterRepo.SoftDeleteByNotificationID(ctx, tx, notification.NotificationID.String)
			if err != nil {
				return fmt.Errorf("error when deleting course filter: %v", err)
			}
		}
		if targetGroup.ClassFilter.Type == consts.TargetGroupSelectTypeList.String() {
			err := svc.NotificationClassFilterRepo.SoftDeleteByNotificationID(ctx, tx, notification.NotificationID.String)
			if err != nil {
				return fmt.Errorf("error when deleting class filter: %v", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed DeleteNotification: %+v", err))
	}
	return &npb.DeleteNotificationResponse{}, nil
}

func (svc *NotificationModifierService) findUndeletedNotification(ctx context.Context, notificationID string) (*entities.InfoNotification, error) {
	noti, err := svc.findNotificationByID(ctx, svc.DB, notificationID)

	if err != nil {
		isDeleted, err := svc.InfoNotificationRepo.IsNotificationDeleted(ctx, svc.DB, database.Text(notificationID))

		if err != nil {
			return nil, fmt.Errorf("svc.InfoNotificationRepo.IsNotificationDeleted: %w", err)
		}
		if isDeleted {
			return nil, fmt.Errorf("the notification has been deleted")
		}
	}

	return noti, nil
}
