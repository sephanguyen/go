package consumers

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type LessonUpdatedHandler struct {
	Logger            *zap.Logger
	WrapperConnection *support.WrapperDBConnection
	JSM               nats.JetStreamManagement

	LiveLessonSentNotificationRepo infrastructure.LiveLessonSentNotificationRepo
}

func (l *LessonUpdatedHandler) Handle(ctx context.Context, msg []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	r := &bpb.EvtLesson{}
	if err := proto.Unmarshal(msg, r); err != nil {
		return false, err
	}

	if _, ok := r.Message.(*bpb.EvtLesson_UpdateLesson_); ok {
		lessonMsg := r.GetUpdateLesson()
		if err := l.handleLessonUpdateEvent(ctx, lessonMsg); err != nil {
			return false, err
		}
	}

	return true, nil
}

func (l *LessonUpdatedHandler) handleLessonUpdateEvent(ctx context.Context, msg *bpb.EvtLesson_UpdateLesson) error {
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return err
	}
	// if lesson is not published, then we don't need to send notification
	// we can also disregard changes to teaching medium and can freely delete records on sent notification table since all of the records are live lessons
	if msg.SchedulingStatusAfter != cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED {
		return nil
	}

	lessonID := msg.GetLessonId()
	startAtBefore := msg.GetStartAtBefore().AsTime()
	startAtAfter := msg.GetStartAtAfter().AsTime()

	// check if new start time is in the future
	// if so, then delete sent notification record for the cronjob to resend notification for this lesson
	// we also buffer the time to make sure that the cronjob has enough time to resend the notification
	if !startAtBefore.Equal(startAtAfter) && startAtAfter.After(time.Now().Add(1*time.Minute)) {
		if err := l.LiveLessonSentNotificationRepo.SoftDeleteLiveLessonSentNotificationRecord(ctx, conn, lessonID); err != nil {
			return fmt.Errorf("LiveLessonSentNotificationRepo.SoftDeleteLiveLessonSentNotificationRecord(lesson id: %s): %v", lessonID, err)
		}
	}

	return nil
}
