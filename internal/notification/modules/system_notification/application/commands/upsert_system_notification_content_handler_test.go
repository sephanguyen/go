package commands

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
	mock_repositories "github.com/manabie-com/backend/mock/notification/modules/system_notification/infrastructure/repo"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_UpsertSystemNotificationContent(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	systemNotificationContentRepo := &mock_repositories.MockSystemNotificationContentRepo{}
	hdl := &SystemNotificationContentHandler{
		SystemNotificationContentRepo: systemNotificationContentRepo,
	}
	testCases := []struct {
		Name    string
		Err     error
		Payload *payloads.UpsertSystemNotificationContentPayload
		Setup   func(ctx context.Context, t *testing.T)
	}{
		{
			Name: "happy case",
			Payload: &payloads.UpsertSystemNotificationContentPayload{
				SystemNotificationID: "event",
				SystemNotificationContents: []*dto.SystemNotificationContent{
					{
						Language: "en",
						Text:     "text_en",
					},
					{
						Language: "vi",
						Text:     "text_vi",
					},
				},
			},
			Err: nil,
			Setup: func(ctx context.Context, t *testing.T) {
				systemNotificationContentRepo.On("SoftDeleteBySystemNotificationID",
					ctx, mockDB.DB,
					"event",
				).Once().Return(nil)
				systemNotificationContentRepo.On("BulkInsertSystemNotificationContents",
					ctx, mockDB.DB,
					mock.Anything,
				).Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			tc.Setup(ctx, t)
			err := hdl.UpsertSystemNotificationContents(ctx, mockDB.DB, tc.Payload)
			assert.Equal(t, tc.Err, err)
		})
	}
}
