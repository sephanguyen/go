package consumers

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_commands "github.com/manabie-com/backend/mock/notification/modules/system_notification/application/commands"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func setupLogsCapture() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)
	return zap.New(core), logs
}

func Test_UpsertSystemNotification(t *testing.T) {
	t.Parallel()

	systemNotificationCmdHdl := &mock_commands.MockSystemNotificationCommandHandler{}
	systemNotificationRecipientCmdHdl := &mock_commands.MockSystemNotificationRecipientCommandHandler{}
	systemNotificationContentCmdHdl := &mock_commands.MockSystemNotificationContentHandler{}
	logger, logs := setupLogsCapture()

	mockDB := testutil.NewMockDB()
	handler := &UpsertSystemNotificationConsumerHandler{
		DB:                               mockDB.DB,
		SystemNotificationCommandHandler: systemNotificationCmdHdl,
		SystemNotificationRecipientCommandHandler: systemNotificationRecipientCmdHdl,
		SystemNotificationContentCommandHandler:   systemNotificationContentCmdHdl,
		Logger:                                    logger,
	}

	testCases := []struct {
		Name               string
		SystemNotification *dto.SystemNotification
		Err                error
		LogMsg             string
		Setup              func(ctx context.Context)
	}{
		{
			Name: "happy case",
			SystemNotification: &dto.SystemNotification{
				ReferenceID: "id",
				ValidFrom:   time.Now(),
				Recipients: []*dto.SystemNotificationRecipient{
					{
						UserID: "user",
					},
				},
			},
			Err:    nil,
			LogMsg: "",
			Setup: func(ctx context.Context) {
				mockTx := &mock_database.Tx{}
				mockDB.DB.On("Begin", ctx, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				systemNotificationCmdHdl.On("UpsertSystemNotification", ctx, mockTx, mock.Anything).Once().Return(nil)
				systemNotificationContentCmdHdl.On("UpsertSystemNotificationContents", ctx, mockTx, mock.Anything).Once().Return(nil)
				systemNotificationRecipientCmdHdl.On("UpsertSystemNotificationRecipients", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "error required fields",
			SystemNotification: &dto.SystemNotification{
				ReferenceID: "id",
				Recipients: []*dto.SystemNotificationRecipient{
					{
						UserID: "user-id",
					},
				},
			},
			Err:    nil,
			LogMsg: "UpsertSystemNotification message failed validation with: %+v",
			Setup:  func(ctx context.Context) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			tc.Setup(ctx)

			err := handler.UpsertSystemNotification(ctx, tc.SystemNotification)
			assert.Equal(t, tc.Err, err)
			if len(tc.LogMsg) > 0 {
				entry := logs.All()[0]
				assert.Equal(t, tc.LogMsg, entry.Message)
			}
		})
	}
}
