package controller

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/queries"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
	systemNotification "github.com/manabie-com/backend/internal/notification/modules/system_notification/util/mapper/systemnotification"
	mock_queries "github.com/manabie-com/backend/mock/notification/modules/system_notification/application/queries"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_RetrieveSystemNotifications(t *testing.T) {
	t.Parallel()

	systemNotificationQueryHdl := &mock_queries.MockSystemNotificationQueryHandler{}

	svc := &SystemNotificationReaderService{
		SystemNotificationQueryHandler: systemNotificationQueryHdl,
	}

	systemNotifications := []*dto.SystemNotification{}

	systemNotificationsPb := systemNotification.ToSystemNotificationPb(systemNotifications)

	var totalCount uint32 = 2
	userID := "user-id"
	c := context.Background()
	ctx := interceptors.ContextWithUserID(c, userID)

	t.Run("happy case", func(t *testing.T) {
		response := &queries.RetrieveSystemNotificationResponse{
			TotalCount:          totalCount,
			SystemNotifications: systemNotifications,
			Error:               nil,
		}
		systemNotificationQueryHdl.
			On("RetrieveSystemNotifications", ctx, mock.Anything).
			Once().
			Return(response, nil)

		res, err := svc.RetrieveSystemNotifications(ctx, &npb.RetrieveSystemNotificationsRequest{
			Paging: &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0}},
		})

		assert.Nil(t, err)
		assert.Equal(t, systemNotificationsPb, res.SystemNotifications)
		assert.Equal(t, totalCount, res.TotalItems)
		assert.Equal(t, &cpb.Paging{
			Limit:  10,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: int64(len(systemNotificationsPb))},
		}, res.NextPage)
		assert.Equal(t, &cpb.Paging{
			Limit:  10,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
		}, res.PreviousPage)
	})

	t.Run("retrieve when paging is nil", func(t *testing.T) {
		response := &queries.RetrieveSystemNotificationResponse{
			TotalCount:          totalCount,
			SystemNotifications: systemNotifications,
			Error:               nil,
		}
		systemNotificationQueryHdl.
			On("RetrieveSystemNotifications", ctx, mock.MatchedBy(func(payload *queries.RetrieveSystemNotificationPayload) bool {
				if payload.Limit == 100 && payload.Offset == 0 {
					return true
				}
				return false
			})).
			Once().
			Return(response, nil)

		res, err := svc.RetrieveSystemNotifications(ctx, &npb.RetrieveSystemNotificationsRequest{
			Paging: nil,
		})

		assert.Nil(t, err)
		assert.Equal(t, systemNotificationsPb, res.SystemNotifications)
		assert.Equal(t, totalCount, res.TotalItems)
		assert.Equal(t, &cpb.Paging{
			Limit:  100,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: int64(len(systemNotificationsPb))},
		}, res.NextPage)
		assert.Equal(t, &cpb.Paging{
			Limit:  100,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
		}, res.PreviousPage)
	})

	t.Run("retrieve system notification error", func(t *testing.T) {
		response := &queries.RetrieveSystemNotificationResponse{
			TotalCount:          totalCount,
			SystemNotifications: []*dto.SystemNotification{},
			Error:               pgx.ErrNoRows,
		}
		systemNotificationQueryHdl.
			On("RetrieveSystemNotifications", ctx, mock.Anything).
			Once().
			Return(response)

		res, err := svc.RetrieveSystemNotifications(ctx, &npb.RetrieveSystemNotificationsRequest{
			Paging: &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0}},
		})

		assert.Nil(t, res)
		assert.Equal(t, err, status.Error(codes.Internal, fmt.Sprintf("svc.SystemNotificationQueryHandler.RetrieveSystemNotifications: %v", pgx.ErrNoRows)))
	})
}
