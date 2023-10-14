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

type AssignmentsCreatedHandler struct {
	Logger          *zap.Logger
	JSM             nats.JetStreamManagement
	StudyPlanWriter interface {
		SyncStudyPlanItemsOnAssignmentsCreated(ctx context.Context, data *npb.EventAssignmentsCreated) error
	}
}

func (h *AssignmentsCreatedHandler) Subscribe() error {
	h.Logger = h.Logger.With(
		zap.String("subject", constants.SubjectAssignmentsCreated),
		zap.String("group", constants.QueueAssignmentsCreated),
		zap.String("durable", constants.DurableAssignmentsCreated),
	)
	h.Logger.Info("AssignmentsCreatedHandler: subscribing to")
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamAssignments, constants.DurableAssignmentsCreated),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverAssignmentCreated),
			nats.AckWait(380 * time.Second),
		},
	}
	_, err := h.JSM.QueueSubscribe(constants.SubjectAssignmentsCreated, constants.QueueAssignmentsCreated, opts, h.handleAssignmentsCreated)
	if err != nil {
		return fmt.Errorf("l.JSM.QueueSubscribe: %w", err)
	}
	return nil
}

func (h *AssignmentsCreatedHandler) handleAssignmentsCreated(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 360*time.Second)
	defer cancel()

	h.Logger = h.Logger.With(
		zap.String("method", "handleAssignmentsCreated"),
	)

	var req npb.EventAssignmentsCreated
	if err := proto.Unmarshal(data, &req); err != nil {
		h.Logger.Error("proto.Unmarshal", zap.Error(err))
		return false, err
	}

	err := h.StudyPlanWriter.SyncStudyPlanItemsOnAssignmentsCreated(ctx, &req)
	if err != nil {
		h.Logger.Error("l.StudyPlanWriter.SyncStudyPlanItems", zap.Error(err))
		return true, err
	}

	return false, nil
}
