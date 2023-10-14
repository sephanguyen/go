package consumers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type CreateLiveLessonRoomHandler struct {
	Logger            *zap.Logger
	WrapperConnection *support.WrapperDBConnection
	JSM               nats.JetStreamManagement

	WhiteboardService infrastructure.WhiteboardPort
	LessonRepo        infrastructure.VirtualLessonRepo
}

func (c *CreateLiveLessonRoomHandler) Handle(ctx context.Context, msg []byte) (bool, error) {
	conn, err := c.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, err
	}
	c.Logger.Info("[CreateLiveLessonRoomEvent]: Received message on",
		zap.String("data", string(msg)),
		zap.String("subject", constants.SubjectLessonCreated),
		zap.String("queue", constants.QueueCreateLiveLessonRoom),
	)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	lessonEventData := &bpb.EvtLesson{}
	if err := proto.Unmarshal(msg, lessonEventData); err != nil {
		c.Logger.Error("[CreateLiveLessonRoomEvent] Failed to parse bpb.EvtLesson: ", zap.Error(err))
		return false, err
	}

	if _, isValidType := lessonEventData.Message.(*bpb.EvtLesson_CreateLessons_); isValidType {
		createLessonData := lessonEventData.GetCreateLessons()
		if createLessonData == nil {
			c.Logger.Error("[CreateLiveLessonRoomEvent] create lesson data is empty", zap.Any("message", createLessonData))
			return false, fmt.Errorf("create lesson data is empty")
		}

		if err := nats.ChunkHandler(len(createLessonData.Lessons), 10, func(start, end int) error {
			var chunkHandlerErrors error
			for _, lesson := range createLessonData.Lessons[start:end] {
				lessonID := lesson.LessonId

				room, err := c.WhiteboardService.CreateRoom(ctx, &whiteboard.CreateRoomRequest{
					Name:     lessonID,
					IsRecord: false,
				})
				if err != nil {
					actualErr := fmt.Errorf("error in WhiteboardService.CreateRoom, lesson %s: %w", lessonID, err)
					chunkHandlerErrors = multierr.Append(chunkHandlerErrors, actualErr)
					continue
				}

				if err = c.LessonRepo.UpdateRoomID(ctx, conn, lessonID, room.UUID); err != nil && !strings.Contains(err.Error(), "cannot update lesson") {
					actualErr := fmt.Errorf("error in LessonRepo.UpdateRoomID, lesson %s: %w", lessonID, err)
					chunkHandlerErrors = multierr.Append(chunkHandlerErrors, actualErr)
				}
			}
			return chunkHandlerErrors
		}); err != nil {
			c.Logger.Error("[CreateLiveLessonRoomEvent]: ", zap.Error(err))
			return false, err
		}
	}

	return true, nil
}
