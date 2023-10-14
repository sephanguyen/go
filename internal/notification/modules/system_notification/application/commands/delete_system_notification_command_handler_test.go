package commands

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/model"
	mock_repositories "github.com/manabie-com/backend/mock/notification/modules/system_notification/infrastructure/repo"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_SoftDeleteSystemNotification(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	systemNotificationRepo := &mock_repositories.MockSystemNotificationRepo{}
	systemNotificationRecipientRepo := &mock_repositories.MockSystemNotificationRecipientRepo{}
	systemNotificationContentRepo := &mock_repositories.MockSystemNotificationContentRepo{}
	hdl := &DeleteSystemNotificationCommandHandler{
		SystemNotificationRepo:          systemNotificationRepo,
		SystemNotificationRecipientRepo: systemNotificationRecipientRepo,
		SystemNotificationContentRepo:   systemNotificationContentRepo,
	}
	referenceID := "referenceID"
	systemNotificationID := "systemNotificationID"

	testCases := []struct {
		Name    string
		Err     error
		Payload *payloads.SoftDeleteSystemNotificationPayload
		Setup   func(ctx context.Context, t *testing.T)
	}{
		{
			Name: "case ReferenceID not exist",
			Payload: &payloads.SoftDeleteSystemNotificationPayload{
				ReferenceID: referenceID,
			},
			Err: errors.New(fmt.Sprintf("not found System Notification by ReferenceID %s", referenceID)),
			Setup: func(ctx context.Context, t *testing.T) {
				systemNotificationRepo.On("FindByReferenceID", ctx, mockDB.DB, referenceID).Once().
					Return(
						nil,
						nil,
					)
				systemNotificationRecipientRepo.On("SoftDeleteBySystemNotificationID", ctx, mockDB.DB, systemNotificationID).Once().
					Return(nil)
				systemNotificationContentRepo.On("SoftDeleteBySystemNotificationID", ctx, mockDB.DB, systemNotificationID).Once().
					Return(nil)
				systemNotificationRepo.On("UpsertSystemNotification",
					ctx, mockDB.DB,
					mock.MatchedBy(func(e *model.SystemNotification) bool {
						if len(e.SystemNotificationID.String) == 0 {
							return false
						}
						if e.DeletedAt.Status != pgtype.Present {
							return false
						}
						return true
					}),
				).Once().Return(nil)
			},
		},
		{
			Name: "happy case",
			Payload: &payloads.SoftDeleteSystemNotificationPayload{
				ReferenceID: referenceID,
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
				systemNotificationRecipientRepo.On("SoftDeleteBySystemNotificationID", ctx, mockDB.DB, systemNotificationID).Once().
					Return(nil)
				systemNotificationContentRepo.On("SoftDeleteBySystemNotificationID", ctx, mockDB.DB, systemNotificationID).Once().
					Return(nil)
				systemNotificationRepo.On("UpsertSystemNotification",
					ctx, mockDB.DB,
					mock.MatchedBy(func(e *model.SystemNotification) bool {
						if len(e.SystemNotificationID.String) == 0 {
							return false
						}
						if e.DeletedAt.Status != pgtype.Present {
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
			err := hdl.SoftDeleteSystemNotification(ctx, mockDB.DB, tc.Payload)
			assert.Equal(t, tc.Err, err)
		})
	}
}
