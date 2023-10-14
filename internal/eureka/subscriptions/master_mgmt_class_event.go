package subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	pbv1 "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	"google.golang.org/protobuf/proto"

	"go.uber.org/zap"
)

type MasterMgmtClassEvent struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement

	MasterMgmtClassEventService interface {
		HandleMasterMgmtClassEvent(ctx context.Context, req *pbv1.EvtClass) error
	}
}

func (j *MasterMgmtClassEvent) Subscribe(ctx context.Context) error {
	j.Logger.Info("MasterMgmtClassEvent: subscribing to",
		zap.String("subject", constants.SubjectMasterMgmtClass),
		zap.String("durable", constants.DurableMasterMgmtClassUpserted),
	)

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamMasterMgmtClass, constants.DurableMasterMgmtClassUpserted),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverMasterMgmtClassEvent),
			nats.AckWait(10 * time.Second),
		},
	}

	_, err := j.JSM.QueueSubscribe(
		constants.SubjectMasterMgmtClass,
		constants.QueueMasterMgmtClassUpserted,
		opts,
		j.handleMasterMgmtClassEvent,
	)
	if err != nil {
		return fmt.Errorf("handleMasterMgmtClassEvent.Subscribe: %w", err)
	}

	return nil
}

func (j *MasterMgmtClassEvent) handleMasterMgmtClassEvent(ctx context.Context, data []byte) (bool, error) {
	req := &pbv1.EvtClass{}
	if err := proto.Unmarshal(data, req); err != nil {
		j.Logger.Error("proto.Unmarshal", zap.Error(err))
		return false, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err := j.MasterMgmtClassEventService.HandleMasterMgmtClassEvent(ctx, req)
	if err != nil {
		j.Logger.Error("err handleMasterMgmtClassEvent", zap.Error(err))
		return true, err
	}
	return false, nil
}
