package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type NotificationSyncStudentPackage struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement

	NotificationModifierService interface {
		SyncStudentPackageV2(ctx context.Context, data *npb.EventStudentPackageV2) error
	}
}

func (ss *NotificationSyncStudentPackage) StartToSubscribe() error {
	ss.Logger.Info("NotificationSyncStudentPackageV2: subscribing to",
		zap.String("subject", constants.SubjectStudentPackageV2EventNats),
		zap.String("group", constants.QueueNotificationSyncStudentPackageEventNatsV2),
		zap.String("durable", constants.DurableNotificationSyncStudentPackageEventNatsV2),
	)

	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamStudentPackageEventNatsV2, constants.DurableNotificationSyncStudentPackageEventNatsV2),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverNotificationSyncStudentPackageEventNatsV2),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err := ss.JSM.QueueSubscribe(constants.SubjectStudentPackageV2EventNats,
		constants.QueueNotificationSyncStudentPackageEventNatsV2,
		option, ss.syncStudentPackageHandler)
	if err != nil {
		return fmt.Errorf("NotificationSyncStudentPackage.Subscribe: %w", err)
	}

	return nil
}

func (ss *NotificationSyncStudentPackage) syncStudentPackageHandler(ctx context.Context, data []byte) (bool, error) {
	// set timeout for syncing job
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var studentPackages npb.EventStudentPackageV2
	if err := proto.Unmarshal(data, &studentPackages); err != nil {
		return false, fmt.Errorf("handleStudentPackageEvent proto.Unmarshal: %w", err)
	}

	err := ss.NotificationModifierService.SyncStudentPackageV2(ctx, &studentPackages)

	if err != nil {
		return true, fmt.Errorf("ss.handleStudentPackageEvent: %w", err)
	}

	return false, nil
}
