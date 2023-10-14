package services

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/notification/consts"
	consumer "github.com/manabie-com/backend/internal/notification/transports/nats"
	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type EntryExitNotifyDetails struct {
	Student    *entities.Student
	TouchEvent eepb.TouchEvent
	Message    string
	Title      string
	TouchTime  time.Time
	RecordType eepb.RecordType
}

func (s *EntryExitModifierService) Notify(ctx context.Context, details *EntryExitNotifyDetails) error {
	data := &ypb.NatsCreateNotificationRequest{
		ClientId:       constant.ClientIDNatsEntryexitmgmtService,
		SendingMethods: []string{consts.SendingMethodPushNotification},
		Target: &ypb.NatsNotificationTarget{
			ReceivedUserIds: []string{details.Student.ID.String},
		},
		TargetGroup: &ypb.NatsNotificationTargetGroup{
			UserGroupFilter: &ypb.NatsNotificationTargetGroup_UserGroupFilter{
				UserGroups: []string{consts.TargetUserGroupParent},
			},
		},
		NotificationConfig: &ypb.NatsPushNotificationConfig{
			Mode:             consts.NotificationModeNotify,
			PermanentStorage: false,
			Notification: &ypb.NatsNotification{
				Title:   details.Title,
				Message: details.Message,
				Content: "<h1>" + details.Message + "</h1>",
			},
			Data: map[string]string{
				"custom_data_type":  "entryexit",
				"custom_data_value": "",
			},
		},
		SendTime: &ypb.NatsNotificationSendTime{
			Type: consts.NotificationTypeImmediate,
		},
		TracingId: uuid.New().String(),
		SchoolId:  details.Student.SchoolID.Int,
	}
	msg, err := proto.Marshal(data)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	if err := try.Do(func(attempt int) (bool, error) {
		if _, err := s.JSM.PublishContext(ctx, consumer.SubjectNotificationCreated, msg); err == nil {
			return false, nil
		}
		time.Sleep(s.retryNotificationSleep)
		return attempt < 4, fmt.Errorf("publish error")
	}); err != nil {
		return status.Error(codes.Internal, fmt.Errorf("s.JSM.PublishContext error: %w", err).Error())
	}

	return nil
}
