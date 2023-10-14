package subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	nats_org "github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type JprepCourseClass struct {
	Logger *zap.Logger
	Subs   []nats_org.Subscription
	JSM    nats.JetStreamManagement

	CourseClassService interface {
		SyncCourseClass(ctx context.Context, req *npb.EventMasterRegistration) error
	}
}

func (j *JprepCourseClass) Subscribe(ctx context.Context) error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamSyncMasterRegistration, constants.DurableSyncCourseClass),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverSyncMasterRegistrationCourseClass),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err := j.JSM.QueueSubscribe(constants.SubjectSyncMasterRegistration,
		constants.QueueSyncCourseClass, option, j.syncCourseClassHandler)
	if err != nil {
		return fmt.Errorf("syncCourseClassSub.Subscribe: %w", err)
	}

	return nil
}

func (j *JprepCourseClass) syncCourseClassHandler(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var req npb.EventMasterRegistration
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncCourseClassHandler proto.Unmarshal: %w", err)
	}
	if len(req.Classes) == 0 {
		return false, fmt.Errorf("syncCourseClassHandler length of classes = 0")
	}
	if err := nats.ChunkHandler(len(req.Classes), constants.MaxRecordProcessPertime, func(start, end int) error {
		return j.CourseClassService.SyncCourseClass(ctx, &npb.EventMasterRegistration{
			Classes: req.Classes[start:end],
		})
	}); err != nil {
		return true, fmt.Errorf("syncCourseClassHandler err syncCourseClass: %w", err)
	}

	return false, nil
}
