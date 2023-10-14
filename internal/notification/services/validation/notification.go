package validation

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/notification/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
)

func ValidateMessageRequiredField(noti *cpb.Notification) error {
	if noti.Message == nil {
		return fmt.Errorf("request Notification.Message is null")
	}

	if noti.Message.Title == "" {
		return fmt.Errorf("request Notification.Message.Title is empty")
	}

	if noti.Message.Content == nil || (noti.Message.Content.Raw == "" && noti.Message.Content.Rendered == "") {
		return fmt.Errorf("request Notification.Message.Content is null")
	}

	return nil
}

func ValidateNotification(notiMsg *entities.InfoNotificationMsg, noti *entities.InfoNotification) error {
	if notiMsg.Title.String == "" {
		return fmt.Errorf("validateNotification.NotificationMessage.Title is empty")
	}
	_, err := noti.GetTargetGroup()
	if err != nil {
		return fmt.Errorf("validateNotification.Notification.GetTargetGroup: %v", err)
	}

	return nil
}

func ValidateUpsertNotificationRequest(req *npb.UpsertNotificationRequest) error {
	if req == nil {
		return fmt.Errorf("request is null")
	}
	if req.Notification == nil {
		return fmt.Errorf("request notification is null")
	}

	err := ValidateTargetGroup(req.Notification)
	if err != nil {
		return err
	}

	err = ValidateMessageRequiredField(req.Notification)
	if err != nil {
		return err
	}

	if req.Notification.Status == cpb.NotificationStatus_NOTIFICATION_STATUS_NONE ||
		req.Notification.Status == cpb.NotificationStatus_NOTIFICATION_STATUS_SENT ||
		req.Notification.Status == cpb.NotificationStatus_NOTIFICATION_STATUS_DISCARD {
		return fmt.Errorf("do not allow req notication status is %v", req.Notification.Status)
	}

	if req.Questionnaire != nil {
		err = ValidateQuestionnairePb(req.Questionnaire, req.Notification)
		return err
	}

	return nil
}

func ValidateScheduledNotification(req *npb.UpsertNotificationRequest) error {
	if req.Notification.ScheduledAt == nil {
		return fmt.Errorf("request Notification.ScheduledAt time is empty")
	}

	scheduledAt := req.Notification.ScheduledAt.AsTime()
	now := time.Now().Truncate(time.Minute).Add(time.Minute)

	if scheduledAt.Before(now) {
		return fmt.Errorf("you cannot schedule at a time in the past")
	}
	return nil
}
