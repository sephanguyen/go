package validation

import (
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"

	"github.com/stretchr/testify/assert"
)

func Test_ValidateSystemNotificationRequiredFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name  string
		Model *dto.SystemNotification
		Err   error
	}{
		{
			Name: "happy case",
			Model: &dto.SystemNotification{
				ReferenceID: "id",
				Content: []*dto.SystemNotificationContent{
					{
						Language: "en",
						Text:     "hello world",
					},
				},
				Recipients: []*dto.SystemNotificationRecipient{
					{UserID: "user"},
				},
				URL:       "https://manabie.com",
				ValidFrom: time.Now(),
			},
			Err: nil,
		},
		{
			Name: "missing Valid From",
			Model: &dto.SystemNotification{
				ReferenceID: "id",
				Recipients: []*dto.SystemNotificationRecipient{
					{UserID: "user"},
				},
				Content: []*dto.SystemNotificationContent{
					{
						Language: "en",
						Text:     "hello world",
					},
				},
				URL: "https://manabie.com",
			},
			Err: fmt.Errorf(ErrMissingValidFrom),
		},
		{
			Name: "missing Reference ID",
			Model: &dto.SystemNotification{
				Recipients: []*dto.SystemNotificationRecipient{
					{UserID: "user"},
				},
				Content: []*dto.SystemNotificationContent{
					{
						Language: "en",
						Text:     "hello world",
					},
				},
				URL: "https://manabie.com",
			},
			Err: fmt.Errorf(ErrMissingReferenceID),
		},
		{
			Name: "empty recipient",
			Model: &dto.SystemNotification{
				Content: []*dto.SystemNotificationContent{
					{
						Language: "en",
						Text:     "hello world",
					},
				},
				ReferenceID: "id",
				URL:         "https://manabie.com",
				ValidFrom:   time.Now(),
			},
			Err: fmt.Errorf(ErrMissingRecipients),
		},
		{
			Name: "empty contents",
			Model: &dto.SystemNotification{
				Content:     []*dto.SystemNotificationContent{},
				ReferenceID: "id",
				Recipients: []*dto.SystemNotificationRecipient{
					{UserID: "user"},
				},
				URL:       "https://manabie.com",
				ValidFrom: time.Now(),
			},
			Err: fmt.Errorf(ErrMissingContents),
		},
		{
			Name: "empty contents 2",
			Model: &dto.SystemNotification{
				ReferenceID: "id",
				Recipients: []*dto.SystemNotificationRecipient{
					{UserID: "user"},
				},
				URL:       "https://manabie.com",
				ValidFrom: time.Now(),
			},
			Err: fmt.Errorf(ErrMissingContents),
		},
		{
			Name: "empty URL",
			Model: &dto.SystemNotification{
				Content: []*dto.SystemNotificationContent{
					{
						Language: "en",
						Text:     "hello world",
					},
				},
				ReferenceID: "id",
				Recipients: []*dto.SystemNotificationRecipient{
					{UserID: "user"},
				},
				URL:       "",
				ValidFrom: time.Now(),
			},
			Err: fmt.Errorf(ErrMissingURL),
		},
		{
			Name: "event delete, ignore other checks",
			Model: &dto.SystemNotification{
				ReferenceID: "id",
				IsDeleted:   true,
			},
			Err: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateSystemNotificationRequiredFields(tc.Model)
			assert.Equal(t, tc.Err, err)
		})
	}
}
