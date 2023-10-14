package systemnotification

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_SystemNotificationsToPb(t *testing.T) {
	t.Parallel()
	time := time.Date(2023, time.Month(6), 05, 01, 02, 03, 0, time.UTC)
	systemNotifications := []*dto.SystemNotification{
		{
			SystemNotificationID: "system-notification-id-1",
			Content: []*dto.SystemNotificationContent{
				{
					Language: "lang1-1",
					Text:     "text1-1",
				},
				{
					Language: "lang1-2",
					Text:     "text1-2",
				},
			},
			URL:       "url-1",
			ValidFrom: time,
			IsDeleted: false,
			Status:    npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW.String(),
		},
		{
			SystemNotificationID: "system-notification-id-2",
			Content: []*dto.SystemNotificationContent{
				{
					Language: "lang2-1",
					Text:     "text2-1",
				},
				{
					Language: "lang2-2",
					Text:     "text2-2",
				},
			},
			URL:       "url-2",
			ValidFrom: time,
			IsDeleted: false,
			Status:    npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String(),
		},
	}
	t.Run("happy case", func(t *testing.T) {
		pb := ToSystemNotificationPb(systemNotifications)
		for i, sn := range systemNotifications {
			assert.Equal(t, sn.SystemNotificationID, pb[i].GetSystemNotificationId(), "unmatched SystemNotificationID")
			assert.Equal(t, sn.URL, pb[i].GetUrl(), "unmatched URL")
			assert.Equal(t, timestamppb.New(time), pb[i].ValidFrom, "unmatched ValidFrom")
			assert.Equal(t, sn.Status, pb[i].Status.String())
			for j, snc := range sn.Content {
				assert.Equal(t, snc.Language, pb[i].Content[j].Language)
				assert.Equal(t, snc.Text, pb[i].Content[j].Text)
			}
		}
	})
}
