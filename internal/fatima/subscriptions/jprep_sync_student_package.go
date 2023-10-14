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

type JprepStudentPackage struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement

	CourseService interface {
		SyncStudentPackage(ctx context.Context, req *npb.EventSyncStudentPackage) error
	}
}

func (j *JprepStudentPackage) Subscribe() error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamSyncStudentPackage, constants.DurableSyncStudentPackageFatima),
			nats.DeliverSubject(constants.DeliverSyncStudentPackageFatima),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err := j.JSM.QueueSubscribe(constants.SubjectSyncStudentPackage,
		constants.QueueSyncStudentPackageFatima, option, j.syncStudentPackageHandler)
	if err != nil {
		return fmt.Errorf("syncStudentSub.Subscribe: %w", err)
	}
	return nil
}

func (j *JprepStudentPackage) syncStudentPackageHandler(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventSyncStudentPackage
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncStudentPackageHandler proto.Unmarshal: %w", err)
	}
	if len(req.StudentPackages) == 0 {
		return false, nil
	}
	err := nats.ChunkHandler(len(req.StudentPackages), constants.MaxRecordProcessPertime, func(start, end int) error {
		return j.CourseService.SyncStudentPackage(ctx, &npb.EventSyncStudentPackage{
			StudentPackages: req.StudentPackages[start:end],
		})
	})
	if err != nil {
		return true, fmt.Errorf("syncStudentPackageHandler err syncStudentPackage: %w", err)
	}

	return false, nil
}
