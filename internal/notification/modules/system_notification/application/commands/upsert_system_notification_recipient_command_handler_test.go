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

func Test_UpsertSystemNotificationRecipient(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	systemNotificationRecipientRepo := &mock_repositories.MockSystemNotificationRecipientRepo{}
	hdl := &SystemNotificationRecipientCommandHandler{
		SystemNotificationRecipientRepo: systemNotificationRecipientRepo,
	}
	testCases := []struct {
		Name    string
		Err     error
		Payload *payloads.UpsertSystemNotificationRecipientPayload
		Setup   func(ctx context.Context, t *testing.T)
	}{
		{
			Name: "happy case",
			Payload: &payloads.UpsertSystemNotificationRecipientPayload{
				SystemNotificationID: "event",
				Recipients: []*dto.SystemNotificationRecipient{
					{
						UserID: "1",
					},
					{
						UserID: "2",
					},
				},
			},
			Err: nil,
			Setup: func(ctx context.Context, t *testing.T) {
				systemNotificationRecipientRepo.On("SoftDeleteBySystemNotificationID",
					ctx, mockDB.DB,
					mock.Anything,
				).Once().Return(nil)
				systemNotificationRecipientRepo.On("BulkInsertSystemNotificationRecipients",
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
			err := hdl.UpsertSystemNotificationRecipients(ctx, mockDB.DB, tc.Payload)
			assert.Equal(t, tc.Err, err)
		})
	}
}
