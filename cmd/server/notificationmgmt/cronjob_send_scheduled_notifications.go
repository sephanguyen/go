package notificationmgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/notification/config"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	bootstrap.RegisterJob("notificationmgmt_send_scheduled_notification", sendScheduledNotification)
}

func sendScheduledNotification(ctx context.Context, c config.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	zapLogger.Info("Start process scheduled notification...")

	notificationmgmtConn := rsc.GRPCDial("notificationmgmt")

	// For example: from 13:00:00 to 13:00:59
	// After that, will send from 13:01:00 to 13:01:59
	from := time.Now().Truncate(time.Minute)
	to := from.Add(time.Second * 59)
	tenantIDs := make([]string, 0)

	zapLogger.Info(fmt.Sprintf("Notification scheduled time will send: From [%v] To [%v]", from, to))
	zapLogger.Info(fmt.Sprintf("Is it run for all tenants: [%t]", c.ScheduledNotificationConfig.IsRunningForAllTenant))
	if !c.ScheduledNotificationConfig.IsRunningForAllTenant {
		tenantIDs = c.ScheduledNotificationConfig.TenantIDs
		zapLogger.Info(fmt.Sprintf("TenantIDs: [%s]", strings.Join(tenantIDs, ", ")))
	}

	req := &npb.SendScheduledNotificationRequest{
		To:                     timestamppb.New(to),
		IsRunningForAllTenants: c.ScheduledNotificationConfig.IsRunningForAllTenant,
		TenantIds:              tenantIDs,
	}
	_, err := npb.NewNotificationModifierServiceClient(notificationmgmtConn).SendScheduledNotification(ctx, req)
	if err != nil {
		zapLogger.Error(err.Error())
	}

	zapLogger.Info("Process DONE!")
	return nil
}
