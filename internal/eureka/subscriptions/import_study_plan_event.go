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

type StudyPlanItemsImportedHandler struct {
	Logger *zap.Logger
	Subs   []natsJS.Subscription
	JSM    nats.JetStreamManagement

	StudyPlanWriter interface {
		ImportStudyPlanItems(ctx context.Context, data *npb.EventImportStudyPlan) error
	}
}

func (l *StudyPlanItemsImportedHandler) Subscribe() error {
	l.Logger = l.Logger.With(
		zap.String("subject", constants.SubjectStudyPlanItemsImported),
		zap.String("group", constants.QueueGroupSubjectStudyPlanItemsImported),
		zap.String("durable", constants.DurableStudyPlanItemsImported),
	)
	l.Logger.Info("StudyPlanItemsImportedHandler: subscribing to")

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamStudyPlanItems, constants.DurableStudyPlanItemsImported),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverStudyPlanItemsImported),
			nats.AckWait(380 * time.Second),
		},
	}

	_, err := l.JSM.QueueSubscribe(constants.SubjectStudyPlanItemsImported, constants.QueueGroupSubjectStudyPlanItemsImported, opts, l.handleStudyPlanItemsImported)
	if err != nil {
		return fmt.Errorf("l.JSM.QueueSubscribe: %w", err)
	}

	return nil
}

func (l *StudyPlanItemsImportedHandler) handleStudyPlanItemsImported(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 360*time.Second)
	defer cancel()

	l.Logger = l.Logger.With(
		zap.String("method", "handleStudyPlanItemsImported"),
	)

	var req npb.EventImportStudyPlan
	if err := proto.Unmarshal(data, &req); err != nil {
		l.Logger.Error("proto.Unmarshal", zap.Error(err))
		return false, err
	}

	err := l.StudyPlanWriter.ImportStudyPlanItems(ctx, &req)
	if err != nil {
		l.Logger.Error("l.StudyPlanWriter.ImportStudyPlanItems", zap.Error(err))
		return true, err
	}

	return false, nil
}
