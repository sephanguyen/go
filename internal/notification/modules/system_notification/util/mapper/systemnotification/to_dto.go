package systemnotification

import (
	"strings"

	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"
)

func KafkaPayloadToDTO(kafkaPayload *payload.UpsertSystemNotification) *dto.SystemNotification {
	// trim slash from the beginning of the string
	kafkaPayload.URL = strings.TrimLeft(kafkaPayload.URL, "/")

	systemNotification := &dto.SystemNotification{
		ReferenceID: kafkaPayload.ReferenceID,
		URL:         kafkaPayload.URL,
		Content:     []*dto.SystemNotificationContent{},
		ValidFrom:   kafkaPayload.ValidFrom,
		IsDeleted:   kafkaPayload.IsDeleted,
		Status:      string(kafkaPayload.Status),
	}

	for _, recipient := range kafkaPayload.Recipients {
		systemNotification.Recipients = append(systemNotification.Recipients, &dto.SystemNotificationRecipient{
			UserID: recipient.UserID,
		})
	}

	for _, snContent := range kafkaPayload.Content {
		systemNotification.Content = append(systemNotification.Content,
			&dto.SystemNotificationContent{
				Language: snContent.Language,
				Text:     snContent.Text,
			},
		)
	}

	return systemNotification
}

func EntitiesToDTO(systemNotifications model.SystemNotifications, systemNotificationContents model.SystemNotificationContents) ([]*dto.SystemNotification, error) {
	mapSystemNotificationContent := make(map[string][]*dto.SystemNotificationContent, 0)
	for _, snc := range systemNotificationContents {
		mapSystemNotificationContent[snc.SystemNotificationID.String] = append(mapSystemNotificationContent[snc.SystemNotificationID.String],
			&dto.SystemNotificationContent{
				Language: snc.Language.String,
				Text:     snc.Text.String,
			})
	}

	ret := make([]*dto.SystemNotification, 0, len(systemNotifications))
	for _, sn := range systemNotifications {
		snDTO := &dto.SystemNotification{SystemNotificationID: sn.SystemNotificationID.String,
			ReferenceID: sn.ReferenceID.String,
			URL:         sn.URL.String,
			Status:      sn.Status.String,
			ValidFrom:   sn.ValidFrom.Time,
		}
		if contents, found := mapSystemNotificationContent[sn.SystemNotificationID.String]; found {
			snDTO.Content = contents
		}
		ret = append(ret, snDTO)
	}
	return ret, nil
}
