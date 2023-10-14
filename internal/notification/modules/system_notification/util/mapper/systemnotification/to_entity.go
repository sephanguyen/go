package systemnotification

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"go.uber.org/multierr"
)

func ToEntity(dto *dto.SystemNotification) (*model.SystemNotification, error) {
	e := &model.SystemNotification{}
	err := multierr.Combine(
		e.SystemNotificationID.Set(dto.SystemNotificationID),
		e.ReferenceID.Set(dto.ReferenceID),
		e.URL.Set(dto.URL),
		e.ValidFrom.Set(dto.ValidFrom),
		e.ValidTo.Set(nil),
		e.CreatedAt.Set(time.Now()),
		e.UpdatedAt.Set(time.Now()),
		e.DeletedAt.Set(nil),
	)

	if dto.Status != "" {
		_ = e.Status.Set(dto.Status)
	} else {
		_ = e.Status.Set(npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW)
	}

	if err != nil {
		return nil, fmt.Errorf("failed ToSystemNotificationEntity: %+v", err)
	}
	return e, nil
}

func ToRecipientEntities(systemNotificationID string, recipients []*dto.SystemNotificationRecipient) (model.SystemNotificationRecipients, error) {
	es := make(model.SystemNotificationRecipients, 0)
	for _, rc := range recipients {
		e := &model.SystemNotificationRecipient{}
		err := multierr.Combine(
			e.SystemNotificationRecipientID.Set(idutil.ULIDNow()),
			e.SystemNotificationID.Set(systemNotificationID),
			e.UserID.Set(rc.UserID),
		)
		if err != nil {
			return nil, fmt.Errorf("failed ToSystemNotificationRecipientEntities: %+v", err)
		}
		es = append(es, e)
	}
	return es, nil
}

func ToSystemNotificationContentEntities(systemNotificationID string, contentList []*dto.SystemNotificationContent) (model.SystemNotificationContents, error) {
	es := make(model.SystemNotificationContents, 0)
	for _, content := range contentList {
		e := es.Add()
		err := multierr.Combine(
			e.(*model.SystemNotificationContent).SystemNotificationContentID.Set(idutil.ULIDNow()),
			e.(*model.SystemNotificationContent).SystemNotificationID.Set(systemNotificationID),
			e.(*model.SystemNotificationContent).Language.Set(content.Language),
			e.(*model.SystemNotificationContent).Text.Set(content.Text),
		)
		if err != nil {
			return nil, fmt.Errorf("failed ToSystemNotificationContentEntities: %+v", err)
		}
	}
	return es, nil
}
