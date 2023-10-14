package commands

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"
	mock_repositories "github.com/manabie-com/backend/mock/notification/modules/system_notification/infrastructure/repo"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_UpsertSystemNotification(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	systemNotificationRepo := &mock_repositories.MockSystemNotificationRepo{}
	hdl := &SystemNotificationCommandHandler{
		SystemNotificationRepo: systemNotificationRepo,
	}
	referenceID := "referenceID"
	systemNotificationID := "systemNotificationID"
	testCases := []struct {
		Name    string
		Err     error
		Payload *payloads.UpsertSystemNotificationPayload
		Setup   func(ctx context.Context, t *testing.T)
	}{
		{
			Name: "happy case",
			Payload: &payloads.UpsertSystemNotificationPayload{
				SystemNotification: &dto.SystemNotification{
					ReferenceID: referenceID,
				},
			},
			Err: nil,
			Setup: func(ctx context.Context, t *testing.T) {
				systemNotificationRepo.On("FindByReferenceID", ctx, mockDB.DB, referenceID).Once().
					Return(
						&model.SystemNotification{
							SystemNotificationID: database.Text(systemNotificationID),
						},
						nil,
					)
				systemNotificationRepo.On("UpsertSystemNotification",
					ctx, mockDB.DB,
					mock.MatchedBy(func(e *model.SystemNotification) bool {
						if len(e.SystemNotificationID.String) == 0 {
							return false
						}
						return true
					}),
				).Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			tc.Setup(ctx, t)
			err := hdl.UpsertSystemNotification(ctx, mockDB.DB, tc.Payload)
			assert.Equal(t, tc.Err, err)
			assert.Equal(t, tc.Payload.SystemNotification.SystemNotificationID, systemNotificationID)
		})
	}
}
