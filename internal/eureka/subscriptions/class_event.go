package subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	bobproto "github.com/manabie-com/backend/pkg/genproto/bob"

	"go.uber.org/zap"
)

type ClassEvent struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement

	ClassEventService interface {
		HandleClassEvent(ctx context.Context, req *bobproto.EvtClassRoom) error
	}
}

func (j *ClassEvent) Subscribe(ctx context.Context) error {
	j.Logger.Info("ClassEvent: subscribing to",
		zap.String("subject", constants.SubjectClass),
		zap.String("durable", constants.DurableClassUpserted),
	)

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamClass, constants.DurableClassUpserted),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverClassEvent),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err := j.JSM.QueueSubscribe(
		constants.SubjectClass,
		constants.QueueClassUpserted,
		opts,
		j.handleClassEvent,
	)
	if err != nil {
		return fmt.Errorf("handleClassEvent.Subscribe: %w", err)
	}

	return nil
}

func (j *ClassEvent) handleClassEvent(ctx context.Context, data []byte) (bool, error) {
	var req bobproto.EvtClassRoom
	if err := req.Unmarshal(data); err != nil {
		j.Logger.Error("proto.Unmarshal", zap.Error(err))
		return false, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := j.ClassEventService.HandleClassEvent(ctx, &req)
	if err != nil {
		j.Logger.Error("err handleClassEvent", zap.Error(err))
		return true, err
	}
	return false, nil
}
