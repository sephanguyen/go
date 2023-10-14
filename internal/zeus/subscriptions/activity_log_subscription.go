package subscriptions

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/zeus/configurations"
	"github.com/manabie-com/backend/internal/zeus/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	natsOrg "github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var (
	batchSize = 10
)

type ActivityLogCreatedEventSubscriber struct {
	CentralizeLogsService interface {
		CreateLogs(ctx context.Context, msg *npb.ActivityLogEvtCreated) error
		BulkCreateLogs(ctx context.Context, logs []*entities.ActivityLog) error
	}
	JSM     nats.JetStreamManagement
	Logger  *zap.Logger
	Configs *configurations.Config
}

func (s *ActivityLogCreatedEventSubscriber) Subscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamActivityLog, constants.DurableActivityLogCreated),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverActivityLogCreated),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
		},
	}

	_, err := s.JSM.QueueSubscribe(
		constants.SubjectActivityLogCreated,
		constants.QueueActivityLogCreated,
		opts,
		s.CreateActivityLog,
	)

	if err != nil {
		return err
	}
	return nil
}

func (s *ActivityLogCreatedEventSubscriber) CreateActivityLog(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var req npb.ActivityLogEvtCreated
	var err error
	if err = proto.Unmarshal(data, &req); err != nil {
		return false, err
	}

	err = s.CentralizeLogsService.CreateLogs(ctx, &req)
	if err != nil {
		s.Logger.Error("c.CentralizeLogsServiceClient.CreateLogs", zap.Error(err))
		return true, err
	}
	return false, nil
}

func (s *ActivityLogCreatedEventSubscriber) PullConsumer() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamActivityLog, constants.DurableActivityLogCreatedPull),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
		},
		PullOpt: nats.PullSubscribeOption{
			FetchSize: 500,
			BatchSize: batchSize,
		},
	}

	return s.JSM.PullSubscribe(constants.SubjectActivityLogCreated, constants.DurableActivityLogCreatedPull,
		s.BatchCreateActivityLog, opts)
}

func (s *ActivityLogCreatedEventSubscriber) BatchCreateActivityLog(msgs []*natsOrg.Msg) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	activityLogs := make([]*entities.ActivityLog, 0, len(msgs))
	err := nats.ChunkHandler(len(msgs), batchSize, func(start, end int) error {
		messages := msgs[start:end]
		for _, msg := range messages {
			var dataInMsg npb.DataInMessage
			var err error
			if err = proto.Unmarshal(msg.Data, &dataInMsg); err != nil {
				if ackErr := msg.Ack(); ackErr != nil {
					s.Logger.Error("msg.Ack", zap.Error(ackErr))
				}

				continue
			}

			var req npb.ActivityLogEvtCreated
			if err = proto.Unmarshal(dataInMsg.Payload, &req); err != nil {
				return err
			}

			activityLog, err := entities.ToActivityLog(req.UserId, req.ActionType, string(req.Payload), req.ResourcePath, req.RequestAt.AsTime(), req.Status, req.FinishedAt.AsTime())
			if err != nil {
				return err
			}

			activityLogs = append(activityLogs, activityLog)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return s.CentralizeLogsService.BulkCreateLogs(ctx, activityLogs)
}
