package subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	natsJS "github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type LearningObjectivesCreatedHandler struct {
	Logger *zap.Logger
	Subs   []natsJS.Subscription
	JSM    nats.JetStreamManagement

	StudyPlanWriter interface {
		SyncStudyPlanItemsOnLOsCreated(ctx context.Context, data *npb.EventLearningObjectivesCreated) error
	}
}

func (l *LearningObjectivesCreatedHandler) Subscribe() error {
	l.Logger = l.Logger.With(
		zap.String("subject", constants.SubjectLearningObjectivesCreated),
		zap.String("group", constants.QueueGroupLearningObjectivesCreated),
		zap.String("durable", constants.DurableLearningObjectivesCreated),
	)
	l.Logger.Info("LearningObjectivesCreatedHandler: subscribing to")

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamLearningObjectives, constants.DurableLearningObjectivesCreated),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverLearningObjectivesCreated),
			nats.AckWait(380 * time.Second),
		},
	}

	_, err := l.JSM.QueueSubscribe(constants.SubjectLearningObjectivesCreated, constants.QueueGroupLearningObjectivesCreated, opts, l.handleLearningObjectivesCreated)
	if err != nil {
		return fmt.Errorf("l.JSM.QueueSubscribe: %w", err)
	}

	return nil
}

func (l *LearningObjectivesCreatedHandler) handleLearningObjectivesCreated(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 360*time.Second)
	defer cancel()

	l.Logger = l.Logger.With(
		zap.String("method", "handleLearningObjectivesCreated"),
	)

	var req npb.EventLearningObjectivesCreated
	if err := proto.Unmarshal(data, &req); err != nil {
		l.Logger.Error("proto.Unmarshal", zap.Error(err))
		return false, err
	}

	err := l.StudyPlanWriter.SyncStudyPlanItemsOnLOsCreated(ctx, &req)
	if err != nil {
		l.Logger.Error("l.StudyPlanWriter.SyncStudyPlanItems", zap.Error(err))
		return true, err
	}

	return false, nil
}
