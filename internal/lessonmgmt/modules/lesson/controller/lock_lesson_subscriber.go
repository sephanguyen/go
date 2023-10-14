package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type LockLessonSubscription struct {
	Logger            *zap.Logger
	JSM               nats.JetStreamManagement
	wrapperConnection *support.WrapperDBConnection

	LessonRepo infrastructure.LessonRepo
}

func (l *LockLessonSubscription) Subscribe() error {
	l.Logger.Info("[LockLessonEvent]: Subscribing to ",
		zap.String("subject", constants.SubjectTimesheetLesson),
		zap.String("group", constants.QueueLockLesson),
		zap.String("durable", constants.DurableLockLesson))

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamTimesheetLesson, constants.DurableLockLesson),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverLockLesson),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "LockLessonSubscription",
	}
	_, err := l.JSM.QueueSubscribe(
		constants.SubjectTimesheetLesson,
		constants.QueueLockLesson,
		opts,
		l.handleLockLessonEvent,
	)
	if err != nil {
		return fmt.Errorf("error subscribing to subject `%s` on `%s` queue: %w",
			constants.SubjectTimesheetLesson,
			constants.QueueLockLesson,
			err,
		)
	}
	return nil
}

func (l *LockLessonSubscription) handleLockLessonEvent(ctx context.Context, msg []byte) (bool, error) {
	l.Logger.Info("[LockLessonEvent]: Received message on",
		zap.String("data", string(msg)),
		zap.String("subject", constants.SubjectTimesheetLesson),
		zap.String("queue", constants.QueueLockLesson),
	)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	var lessonEvent tpb.TimesheetLessonLockEvt
	err := proto.Unmarshal(msg, &lessonEvent)

	if err != nil {
		l.Logger.Error("Failed to parse tpb.TimesheetLessonLockEvt: ", zap.Error(err))
		return false, err
	}

	conn, err := l.wrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, err
	}

	if err = l.LessonRepo.LockLesson(ctx, conn, lessonEvent.GetLessonIds()); err != nil {
		return false, err
	}

	return true, nil
}
