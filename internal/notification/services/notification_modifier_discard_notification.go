package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DiscardNotification only discard draft, scheduled notification
func (svc *NotificationModifierService) DiscardNotification(ctx context.Context, req *npb.DiscardNotificationRequest) (*npb.DiscardNotificationResponse, error) {
	notification, err := svc.findDiscardableNotification(ctx, req.NotificationId)
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
		}

		softDeleteNotificationTagFilter := repositories.NewSoftDeleteNotificationTagFilter()
		_ = softDeleteNotificationTagFilter.NotificationIDs.Set([]string{notification.NotificationID.String})
		err = svc.InfoNotificationTagRepo.SoftDelete(ctx, tx, softDeleteNotificationTagFilter)
		if err != nil {
			return fmt.Errorf("InfoNotificationTagRepo.SoftDelete %v", err)
		}

		status := []string{cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String(), cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()}
		err = svc.InfoNotificationRepo.DiscardNotification(ctx, tx, database.Text(notification.NotificationID.String), database.TextArray(status))
		if err != nil {
			return fmt.Errorf("InfoNotificationRepo.DiscardDraftNotification %v", err)
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
		return nil, status.Error(codes.Internal, fmt.Sprintf("ExecInTxWithRetry: %v", err))
	}

	resp := &npb.DiscardNotificationResponse{}
	return resp, nil
}

func (svc *NotificationModifierService) findDiscardableNotification(ctx context.Context, notificationID string) (*entities.InfoNotification, error) {
	noti, err := svc.findUndeletedNotification(ctx, notificationID)
	if err != nil {
		return nil, err
	}

	if noti.Status.String == cpb.NotificationStatus_NOTIFICATION_STATUS_SENT.String() {
		return nil, fmt.Errorf("the notification has been sent, you can no longer discard this notification")
	}

	if noti.Status.String != cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT.String() && noti.Status.String != cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String() {
		return nil, fmt.Errorf("notification %s has status %s and not discardable", noti.NotificationID.String, noti.Status.String)
	}

	return noti, nil
}
