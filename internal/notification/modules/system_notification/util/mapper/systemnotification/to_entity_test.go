package systemnotification

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func Test_ToEntity(t *testing.T) {
	t.Parallel()

	validFrom := time.Now()
	testCases := []struct {
		Name string
		In   *dto.SystemNotification
		Out  *model.SystemNotification
	}{
		{
			Name: "happy case",
			In: &dto.SystemNotification{
				SystemNotificationID: "1",
				ReferenceID:          "1",
				// Content:              "event",
				URL:       "url",
				ValidFrom: validFrom,
				IsDeleted: false,
				Status:    npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW.String(),
			},
			Out: &model.SystemNotification{
				SystemNotificationID: database.Text("1"),
				ReferenceID:          database.Text("1"),
				// Content:              database.Text("event"),
				URL:       database.Text("url"),
				ValidFrom: database.Timestamptz(validFrom),
				CreatedAt: database.Timestamptz(time.Now()),
				UpdatedAt: database.Timestamptz(time.Now()),
				DeletedAt: database.TimestamptzNull(time.Time{}),
				Status:    database.Text(npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_NEW.String()),
			},
		},
		{
			Name: "empty fields",
			In: &dto.SystemNotification{
				ReferenceID: "1",
				// Content:     "event",
				URL:       "url",
				ValidFrom: validFrom,
				Status:    npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String(),
			},
			Out: &model.SystemNotification{
				SystemNotificationID: database.Text(""),
				ReferenceID:          database.Text("1"),
				// Content:              database.Text("event"),
				URL:       database.Text("url"),
				ValidFrom: database.Timestamptz(validFrom),
				CreatedAt: database.Timestamptz(time.Now()),
				UpdatedAt: database.Timestamptz(time.Now()),
				DeletedAt: database.TimestamptzNull(time.Time{}),
				Status:    database.Text(npb.SystemNotificationStatus_SYSTEM_NOTIFICATION_STATUS_DONE.String()),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			m, err := ToEntity(tc.In)
			assert.Nil(t, err)
			assert.Equal(t, tc.Out.SystemNotificationID, m.SystemNotificationID)
			assert.Equal(t, tc.Out.ReferenceID, m.ReferenceID)
			// assert.Equal(t, tc.Out.Content, m.Content)
			assert.Equal(t, tc.Out.URL, m.URL)
			assert.Equal(t, tc.Out.Status, m.Status)
			assert.Equal(t, tc.Out.CreatedAt.Status, pgtype.Present)
			assert.Equal(t, tc.Out.UpdatedAt.Status, pgtype.Present)
			if tc.In.IsDeleted {
				assert.Equal(t, tc.Out.DeletedAt.Status, pgtype.Present)
			} else {
				assert.Equal(t, tc.Out.DeletedAt.Status, pgtype.Null)
			}
		})
	}
}

func Test_ToRecipientEntities(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name                 string
		SystemNotificationID string
		Recipients           []*dto.SystemNotificationRecipient
		OutRecipients        model.SystemNotificationRecipients
	}{
		{
			Name:                 "happy case",
			SystemNotificationID: "event",
			Recipients: []*dto.SystemNotificationRecipient{
				{
					UserID: "1",
				},
				{
					UserID: "2",
				},
				{
					UserID: "3",
				},
			},
			OutRecipients: model.SystemNotificationRecipients{
				{
					SystemNotificationID: database.Text("event"),
					UserID:               database.Text("1"),
				},
				{
					SystemNotificationID: database.Text("event"),
					UserID:               database.Text("2"),
				},
				{
					SystemNotificationID: database.Text("event"),
					UserID:               database.Text("3"),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			es, err := ToRecipientEntities(tc.SystemNotificationID, tc.Recipients)
			assert.Nil(t, err)
			for i, e := range es {
				assert.Equal(t, e.UserID, tc.OutRecipients[i].UserID)
				assert.Equal(t, e.SystemNotificationID, tc.OutRecipients[i].SystemNotificationID)
				assert.Equal(t, e.SystemNotificationRecipientID.Status, pgtype.Present)
			}
		})
	}
}

func Test_ToSystemNotificationContentEntities(t *testing.T) {
	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		systemNotificationID := "snID"
		contentList := []*dto.SystemNotificationContent{
			{
				Language: "en",
				Text:     "<p>hello world</p>",
			},
			{
				Language: "vi",
				Text:     "<p>xin chào</p>",
			},
			{
				Language: "",
				Text:     "",
			},
		}

		ents, err := ToSystemNotificationContentEntities(systemNotificationID, contentList)
		assert.Nil(t, err)
		assert.Equal(t, len(contentList), len(ents))
		for i, e := range ents {
			assert.NotEmpty(t, e.SystemNotificationContentID)
			assert.Equal(t, systemNotificationID, e.SystemNotificationID.String)
			assert.Equal(t, contentList[i].Language, e.Language.String)
			assert.Equal(t, contentList[i].Text, e.Text.String)
		}
	})
}
