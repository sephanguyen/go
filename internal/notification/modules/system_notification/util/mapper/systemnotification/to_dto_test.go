package systemnotification

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"

	"github.com/stretchr/testify/assert"
)

func Test_KafkaPayloadToDTO(t *testing.T) {
	t.Parallel()

	referenceID := "referenceID"
	url := "url"
	validFrom := time.Now()

	testCases := []struct {
		Name    string
		Payload *payload.UpsertSystemNotification
		DTO     *dto.SystemNotification
	}{
		{
			Name: "happy case",
			Payload: &payload.UpsertSystemNotification{
				ReferenceID: referenceID,
				URL:         "///" + url,
				Content: []payload.SystemNotificationContent{
					{
						Language: "en",
						Text:     "",
					},
					{
						Language: "vi",
						Text:     "<p>xin chào</p>",
					},
				},
				ValidFrom: validFrom,
				Recipients: []payload.SystemNotificationRecipient{
					{
						UserID: "user-id1",
					},
					{
						UserID: "user-id2",
					},
				},
				Status: payload.SystemNotificationStatusNew,
			},
			DTO: &dto.SystemNotification{
				ReferenceID: referenceID,
				URL:         url,
				Content: []*dto.SystemNotificationContent{
					{
						Language: "en",
						Text:     "",
					},
					{
						Language: "vi",
						Text:     "<p>xin chào</p>",
					},
				},
				ValidFrom: validFrom,
				Recipients: []*dto.SystemNotificationRecipient{
					{
						UserID: "user-id1",
					},
					{
						UserID: "user-id2",
					},
				},
				Status: string(payload.SystemNotificationStatusNew),
			},
		},
		{
			Name: "case deleted",
			Payload: &payload.UpsertSystemNotification{
				ReferenceID: referenceID,
				URL:         url,
				Content: []payload.SystemNotificationContent{
					{
						Language: "en",
						Text:     "<p>hello world</p>",
					},
				},
				ValidFrom: validFrom,
				Recipients: []payload.SystemNotificationRecipient{
					{
						UserID: "user-id1",
					},
					{
						UserID: "user-id2",
					},
				},
				IsDeleted: true,
				Status:    payload.SystemNotificationStatusDone,
			},
			DTO: &dto.SystemNotification{
				ReferenceID: referenceID,
				URL:         url,
				Content: []*dto.SystemNotificationContent{
					{
						Language: "en",
						Text:     "<p>hello world</p>",
					},
				},
				ValidFrom: validFrom,
				Recipients: []*dto.SystemNotificationRecipient{
					{
						UserID: "user-id1",
					},
					{
						UserID: "user-id2",
					},
				},
				IsDeleted: true,
				Status:    string(payload.SystemNotificationStatusDone),
			},
		},
		{
			Name: "empty content",
			Payload: &payload.UpsertSystemNotification{
				ReferenceID: referenceID,
				URL:         url,
				Content:     []payload.SystemNotificationContent{},
				ValidFrom:   validFrom,
				Recipients: []payload.SystemNotificationRecipient{
					{
						UserID: "user-id1",
					},
					{
						UserID: "user-id2",
					},
				},
				IsDeleted: true,
				Status:    payload.SystemNotificationStatusDone,
			},
			DTO: &dto.SystemNotification{
				ReferenceID: referenceID,
				URL:         url,
				Content:     []*dto.SystemNotificationContent{},
				ValidFrom:   validFrom,
				Recipients: []*dto.SystemNotificationRecipient{
					{
						UserID: "user-id1",
					},
					{
						UserID: "user-id2",
					},
				},
				IsDeleted: true,
				Status:    string(payload.SystemNotificationStatusDone),
			},
		},
	}

	for _, tc := range testCases {
		res := KafkaPayloadToDTO(tc.Payload)
		assert.Equal(t, tc.DTO, res)
	}
}
