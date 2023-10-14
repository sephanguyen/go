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

type JprepSyncStudentPackage struct {
	Logger                      *zap.Logger
	JSM                         nats.JetStreamManagement
	NotificationModifierService interface {
		SyncJprepStudentPackage(ctx context.Context, data []*npb.EventSyncStudentPackage_StudentPackage) error
	}
}

func (s *JprepSyncStudentPackage) StartToSubscribe() error {
	s.Logger.Info("JprepSyncStudentPackage: subscribing to",
		zap.String("stream", constants.StreamSyncJprepStudentPackageEventNats),
		zap.String("subject", constants.SubjectSyncJprepStudentPackageEventNats),
		zap.String("queue", constants.QueueNotificationSyncJprepStudentPackageEventNats),
		zap.String("durable", constants.DurableNotificationSyncJprepStudentPackageEventNats),
	)

	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamSyncJprepStudentPackageEventNats, constants.DurableNotificationSyncJprepStudentPackageEventNats),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverNotificationSyncJprepStudentPackageEventNats),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err := s.JSM.QueueSubscribe(constants.SubjectSyncJprepStudentPackageEventNats,
		constants.QueueNotificationSyncJprepStudentPackageEventNats,
		option, s.syncStudentPackageHandler)
	if err != nil {
		return fmt.Errorf("NotificationSyncJprepStudentPackage.Subscribe: %w", err)
	}

	return nil
}

func (s *JprepSyncStudentPackage) syncStudentPackageHandler(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventSyncStudentPackage
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncStudentPackageHandler proto.Unmarshal: %w", err)
	}
	if len(req.StudentPackages) == 0 {
		return false, nil
	}

	err := s.NotificationModifierService.SyncJprepStudentPackage(ctx, req.StudentPackages)
	if err != nil {
		return true, fmt.Errorf("s.SyncJprepStudentPackage: %w", err)
	}

	return false, nil
}
