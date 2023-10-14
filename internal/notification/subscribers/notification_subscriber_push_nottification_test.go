package subscribers

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/notification/consts"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

func Test_ProcessPushNotification(t *testing.T) {
	tests := []struct {
		name    string
		data    *ypb.NatsCreateNotificationRequest
		wantErr bool
	}{
		{
			name: "case fail client",
			data: &ypb.NatsCreateNotificationRequest{
				ClientId: "client_id_fail",
				SendTime: &ypb.NatsNotificationSendTime{
					Type: consts.NotificationTypeScheduled,
				},
				Target: &ypb.NatsNotificationTarget{
					ReceivedUserIds: []string{"user_id"},
				},
				NotificationConfig: &ypb.NatsPushNotificationConfig{
					Mode: "notify",
					Data: map[string]string{
						"custom_data_type":  "test",
						"custom_data_value": "{}",
					},
					Notification: &ypb.NatsNotification{
						Title:   "title",
						Content: "<h1>hello world</h1>",
						Message: "hello",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "happy case",
			data: &ypb.NatsCreateNotificationRequest{
				ClientId: "bdd_testing_client_id",
				SendTime: &ypb.NatsNotificationSendTime{
					Type: consts.NotificationTypeScheduled,
				},
				Target: &ypb.NatsNotificationTarget{
					ReceivedUserIds: []string{"user_id"},
				},
				NotificationConfig: &ypb.NatsPushNotificationConfig{
					Mode: "notify",
					Data: map[string]string{
						"custom_data_type":  "test",
						"custom_data_value": "{}",
					},
					Notification: &ypb.NatsNotification{
						Title:   "title",
						Content: "<h1>hello world</h1>",
						Message: "hello",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notiSubscriber := &NotificationSubscriber{}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			err := notiSubscriber.ProcessPushNotification(ctx, tt.data)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("processMsg() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			}
		})
	}
}
