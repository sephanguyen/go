package controller

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/application/consumers"

	"go.uber.org/zap"
)

type StudentCourseSlotInfoSubscriber struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	consumers.SubscriberHandler
}

func (s *StudentCourseSlotInfoSubscriber) Subscribe() error {
	s.Logger.Info("[StudentCourseSlotInfoSubscriber]: Subscribing to ",
		zap.String("subject", constants.SubjectStudentCourseEventSync),
		zap.String("group", constants.QueueLessonSyncStudentCourseSlotInfo),
		zap.String("durable", constants.DurableLessonSyncStudentCourseSlotInfo),
	)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(), nats.AckWait(30 * time.Second),
			nats.MaxDeliver(10),
			nats.Bind(constants.StreamStudentCourseEventSync, constants.DurableLessonSyncStudentCourseSlotInfo),
			nats.DeliverSubject(constants.DeliverLessonSyncStudentCourseSlotInfo),
		},
		SpanName: "subscribeStudentCourseSlotInfo",
	}
	_, err := s.JSM.QueueSubscribe(
		constants.SubjectStudentCourseEventSync,
		constants.QueueLessonSyncStudentCourseSlotInfo,
		opts,
		s.Handle,
	)
	if err != nil {
		return fmt.Errorf("error subscribing to subject `%s` on `%s` queue: %w",
			constants.SubjectStudentCourseEventSync,
			constants.QueueLessonSyncStudentCourseSlotInfo,
			err,
		)
	}
	return nil
}
