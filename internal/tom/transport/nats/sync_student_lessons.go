package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/tom/configurations"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// Jprep lesson's students are synced by this queue
type StudentLessonsSubscriptions struct {
	Config             *configurations.Config
	JSM                nats.JetStreamManagement
	Logger             *zap.Logger
	LessonChatModifier interface {
		SyncLessonConversationStudents(context.Context, []*npb.EventSyncUserCourse_StudentLesson) error
	}
}

func (rcv *StudentLessonsSubscriptions) Subscribe() error {

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamSyncStudentLessons, constants.DurableSyncStudentLessonsConversations),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverSyncStudentLessonsConversations),
			nats.AckWait(30 * time.Second),
		},
	}

	_, err := rcv.JSM.QueueSubscribe(constants.SubjectSyncStudentLessons, constants.QueueSyncStudentLessonsConversations, opts, rcv.HandleStudentLessonsCreated)
	if err != nil {
		return fmt.Errorf("rcv.JSM.QueueSubscribe: %w", err)
	}

	return nil
}

func (rcv *StudentLessonsSubscriptions) HandleStudentLessonsCreated(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventSyncUserCourse
	if err := proto.Unmarshal(data, &req); err != nil {
		rcv.Logger.Error("proto.Unmarshal", zap.Error(err))
		return false, fmt.Errorf("proto.Unmarshal: %w", err)
	}

	logger := rcv.Logger.With(
		zap.String("subject", constants.SubjectSyncStudentLessons),
		zap.String("queue", constants.QueueSyncStudentLessonsConversations),
	)

	err := nats.ChunkHandler(len(req.StudentLessons), constants.MaxRecordProcessPertime, func(start, end int) error {
		return rcv.LessonChatModifier.SyncLessonConversationStudents(ctx, req.StudentLessons[start:end])
	})
	if err != nil {
		logger.Error("LessonChatModifier.SyncLessonConversationStudents", zap.Error(err))
		return true, err
	}
	return false, nil
}
