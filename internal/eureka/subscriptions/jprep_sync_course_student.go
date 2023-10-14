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

type JprepCourseStudent struct {
	JSM    nats.JetStreamManagement
	Logger *zap.Logger

	CourseStudentService interface {
		SyncCourseStudent(ctx context.Context, req *npb.EventSyncStudentPackage) error
	}
}

func (j *JprepCourseStudent) Subscribe(ctx context.Context) error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamSyncStudentPackage, constants.DurableSyncStudentPackageEureka),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverSyncStudentPackageEureka),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err := j.JSM.QueueSubscribe(constants.SubjectSyncStudentPackage,
		constants.QueueSyncStudentPackageEureka, option, j.syncCourseStudentHandler)
	if err != nil {
		return fmt.Errorf("syncCourseStudentSub.Subscribe: %w", err)
	}
	return nil
}

func (j *JprepCourseStudent) syncCourseStudentHandler(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventSyncStudentPackage
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncCourseStudentHandler proto.Unmarshal: %w", err)
	}
	if len(req.StudentPackages) == 0 {
		return false, fmt.Errorf("syncCourseStudentHandler length of StudentPackages = 0")
	}
	err := nats.ChunkHandler(len(req.StudentPackages), constants.MaxRecordProcessPertime, func(start, end int) error {
		return j.CourseStudentService.SyncCourseStudent(ctx, &npb.EventSyncStudentPackage{
			StudentPackages: req.StudentPackages[start:end],
		})
	})
	if err != nil {
		return true, fmt.Errorf("syncCourseStudentHandler err syncCourseStudentHandler: %w", err)
	}

	return false, nil
}
