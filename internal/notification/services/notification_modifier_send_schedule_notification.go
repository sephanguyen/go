package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (svc *NotificationModifierService) SendScheduledNotification(ctx context.Context, req *npb.SendScheduledNotificationRequest) (*npb.SendScheduledNotificationResponse, error) {
	logger := ctxzap.Extract(ctx)

	var tenantIDs []string
	var err error
	if req.IsRunningForAllTenants {
		tenantIDs, err = svc.OrganizationRepo.GetOrganizations(ctx, svc.DB)
		if err != nil {
			return nil, fmt.Errorf("query tenant has error %v", err)
		}
	} else {
		tenantIDs = req.TenantIds
	}

	logger.Sugar().Info(fmt.Sprintf("Starting process scheduled notification for tenants: [%s]", strings.Join(tenantIDs, ", ")))

	var wg sync.WaitGroup
	for _, tenant := range tenantIDs {
		wg.Add(1)
		go func(tentIdString string) {
			defer wg.Done()
			logger.Sugar().Info("fake clamis to start job")
			// make tenant context with a RSL resource path
			tenantContext := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: tentIdString,
					UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
				},
			})

			// use tenant context to get internal user
			internalUser, err := svc.NotificationInternalUserRepo.GetByOrgID(tenantContext, svc.DB, tentIdString)
			internalUserID := ""
			// temporarily ignore internal user not found (err == pgx.ErrNoRows)
			if err != nil && err != pgx.ErrNoRows {
				logger.Sugar().Errorf("query internal user of tenant %v has err %v", tentIdString, err)
			} else if err == pgx.ErrNoRows {
				logger.Sugar().Warnf("query internal user of tenant %v has err %v", tentIdString, err)
			} else {
				internalUserID = internalUser.UserID.String
			}

			tenantWithInternalUserContext := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: tentIdString,
					UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
					UserID:       internalUserID,
				},
			})

			// send notification of tenant
			err = svc.sendScheduledNotifyOfTenant(tenantWithInternalUserContext, tentIdString, req.To, logger, internalUserID)
			if err != nil {
				// TODO send to slack
				logger.Sugar().Error(fmt.Sprintf("send scheduled notification of tenant %v has err %v", tentIdString, err))
			}
		}(tenant)
	}
	wg.Wait()
	logger.Sugar().Infof("process scheduled notification of [%s] done", strings.Join(tenantIDs, ", "))
	return &npb.SendScheduledNotificationResponse{}, nil
}

func (svc *NotificationModifierService) sendScheduledNotifyOfTenant(tenantContext context.Context, tenant string, triggerTime *timestamppb.Timestamp, logger *zap.Logger, internalUserID string) error {
	filter := repositories.NewFindNotificationFilter()
	filter.Status = database.TextArray([]string{cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED.String()})
	filter.ToScheduled = database.TimestamptzFromPb(triggerTime)
	filter.ResourcePath = database.Text(tenant)

	notifications, err := svc.InfoNotificationRepo.Find(tenantContext, svc.DB, filter)
	if err != nil {
		return err
	}

	if len(notifications) == 0 {
		return nil
	}

	notifyMsgIDs := make([]string, 0)
	notifyMsgMap := make(map[string]*entities.InfoNotification)
	for _, n := range notifications {
		notifyMsgIDs = append(notifyMsgIDs, n.NotificationMsgID.String)
		notifyMsgMap[n.NotificationMsgID.String] = n
	}

	notifyMsgs, err := svc.InfoNotificationMsgRepo.GetByIDs(tenantContext, svc.DB, database.TextArray(notifyMsgIDs))
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex
	failedNotification := make([]string, 0)
	logger.Sugar().Info("start to send %v scheduled notification of tenant_id: %v", len(notifications), tenant)
	for i := 0; i < len(notifyMsgs); i++ {
		wg.Add(1)
		go func(msg *entities.InfoNotificationMsg) {
			defer wg.Done()
			notify, ok := notifyMsgMap[msg.NotificationMsgID.String]

			if !ok {
				logger.Sugar().Errorf("scheduled notification message %v failed with error notification not found", msg.NotificationMsgID)
				mutex.Lock()
				failedNotification = append(failedNotification, msg.NotificationMsgID.String)
				mutex.Unlock()
			} else {
				logger.Sugar().Infof("process scheduled notification [%v|%v]", notify.NotificationID.String, msg.NotificationMsgID.String)

				organizationID := notify.Owner.Int
				userIDForLog := "cron-job-send-scheduled-notification+" + internalUserID

				// use EditedUserCtx to help detect granted locations changed and deal with edge cases
				editedUserCtx := interceptors.ContextWithJWTClaims(tenantContext, &interceptors.CustomClaims{
					Manabie: &interceptors.ManabieClaims{
						UserID:       notify.EditorID.String,
						ResourcePath: strconv.Itoa(int(organizationID)),
					},
				})

				err = svc.sendNotification(editedUserCtx, notify, msg, organizationID, userIDForLog)
				if err != nil {
					logger.Sugar().Errorf("scheduled notification %v failed with error %v", msg.NotificationMsgID, err.Error())
					mutex.Lock()
					failedNotification = append(failedNotification, msg.NotificationMsgID.String)
					mutex.Unlock()
				} else {
					logger.Sugar().Infof("send notification id %v success", notify.NotificationID.String)
				}
			}
		}(notifyMsgs[i])
		if i%500 == 0 {
			wg.Wait()
		}
	}
	wg.Wait()
	logger.Sugar().Infof("End send scheduled notification of %v", tenant)
	if len(failedNotification) > 0 {
		return fmt.Errorf("scheduled notifications failed %v", strings.Join(failedNotification, ","))
	}
	return nil
}
