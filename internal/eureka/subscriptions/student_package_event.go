package subscriptions

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

type StudentPackageEvent struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement

	StudentPackageEventService interface {
		HandleStudentPackageEvent(ctx context.Context, req *npb.EventStudentPackage) error
		HandleStudentPackageEventV2(ctx context.Context, req *npb.EventStudentPackageV2) error
		HandleStudentPackageEventV3(ctx context.Context, req *npb.EventStudentPackageV2) error
	}
}

func (e *StudentPackageEvent) Subscribe(ctx context.Context) error {
	e.Logger.Info("StudentPackageEvent: subscribing to",
		zap.String("subject", constants.SubjectStudentPackageEventNats),
		zap.String("group", constants.QueueStudentPackageEventNats),
		zap.String("durable", constants.DurableStudentPackageEventNats),
	)

	opt := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamStudentPackageEventNats, constants.DurableStudentPackageEventNats),
			nats.DeliverSubject(constants.DeliverStudentPackageEventNats),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err := e.JSM.QueueSubscribe(constants.SubjectStudentPackageEventNats,
		constants.QueueStudentPackageEventNats, opt, e.handleStudentPackageEvent)

	if err != nil {
		return fmt.Errorf("handleStudentEvent.Subscribe: %w", err)
	}

	return nil
}

func (e *StudentPackageEvent) SubscribeV2(ctx context.Context) error {
	e.Logger.Info("StudentPackageEventV2: subscribing to",
		zap.String("subject", constants.SubjectStudentPackageV2EventNats),
		zap.String("group", constants.QueueStudentPackageEventNatsV2),
		zap.String("durable", constants.DurableStudentPackageEventNatsV2),
	)

	opt := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamStudentPackageEventNatsV2, constants.DurableStudentPackageEventNatsV2),
			nats.DeliverSubject(constants.DeliverStudentPackageEventNatsV2),
			nats.MaxDeliver(10),
			nats.AckWait(200 * time.Second),
		},
	}

	_, err := e.JSM.QueueSubscribe(constants.SubjectStudentPackageV2EventNats,
		constants.QueueStudentPackageEventNatsV2, opt, e.handleStudentPackageEventV2)

	if err != nil {
		return fmt.Errorf("handleStudentEventV2.Subscribe: %w", err)
	}

	return nil
}

func (e *StudentPackageEvent) handleStudentPackageEvent(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventStudentPackage
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, err
	}

	err := e.StudentPackageEventService.HandleStudentPackageEvent(ctx, &req)
	if err != nil {
		return true, err
	}

	return false, nil
}

func (e *StudentPackageEvent) handleStudentPackageEventV2(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	var req npb.EventStudentPackageV2
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, err
	}

	err := e.StudentPackageEventService.HandleStudentPackageEventV2(ctx, &req)
	if err != nil {
		return true, err
	}

	return false, nil
}
