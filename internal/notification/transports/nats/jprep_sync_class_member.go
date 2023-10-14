package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"go.uber.org/zap"
)

type JprepSyncClassMember struct {
	Logger                      *zap.Logger
	JSM                         nats.JetStreamManagement
	NotificationModifierService interface {
		SyncJprepClassMember(ctx context.Context, req *pb.EvtClassRoom) error
	}
}

func (s *JprepSyncClassMember) StartToSubscribe() error {
	s.Logger.Info("JprepSyncClassMember: subscribing to",
		zap.String("stream", constants.StreamClass),
		zap.String("subject", constants.SubjectClassUpserted),
		zap.String("queue", constants.QueueNotificationClassUpserted),
		zap.String("durable", constants.DurableNotificationClassUpserted),
		zap.String("deliver", constants.DeliverNotificationClassUpserted),
	)

	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamClass, constants.DurableNotificationClassUpserted),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverNotificationClassUpserted),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err := s.JSM.QueueSubscribe(constants.SubjectClassUpserted,
		constants.QueueNotificationClassUpserted,
		option, s.syncClassMemberHandler)
	if err != nil {
		return fmt.Errorf("NotificationJprepSyncClassMember.Subscribe: %w", err)
	}

	return nil
}

func (s *JprepSyncClassMember) syncClassMemberHandler(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req := &pb.EvtClassRoom{}
	if err := req.Unmarshal(data); err != nil {
		return false, err
	}

	err := s.NotificationModifierService.SyncJprepClassMember(ctx, req)
	if err != nil {
		return true, err
	}

	return false, nil
}
