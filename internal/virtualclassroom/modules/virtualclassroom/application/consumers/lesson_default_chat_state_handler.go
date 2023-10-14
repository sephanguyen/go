package consumers

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type LessonDefaultChatStateHandler struct {
	Logger            *zap.Logger
	WrapperConnection *support.WrapperDBConnection
	JSM               nats.JetStreamManagement

	LessonMemberRepo infrastructure.LessonMemberRepo
}

func (l *LessonDefaultChatStateHandler) Handle(ctx context.Context, msg []byte) (bool, error) {
	conn, err := l.WrapperConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return false, err
	}
	l.Logger.Info("[LessonDefaultChatStateEvent]: Received message on",
		zap.String("data", string(msg)),
		zap.String("subject", constants.SubjectLessonUpdated),
		zap.String("queue", constants.QueueLessonDefaultChatState),
	)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	lessonEventData := &bpb.EvtLesson{}
	if err := proto.Unmarshal(msg, lessonEventData); err != nil {
		l.Logger.Error("[LessonDefaultChatStateEvent] Failed to parse bpb.EvtLesson: ", zap.Error(err))
		return false, err
	}

	if _, isValidType := lessonEventData.Message.(*bpb.EvtLesson_JoinLesson_); isValidType {
		joinLessonData := lessonEventData.GetJoinLesson()
		// from join live lesson API, user group can only be teacher or student
		userGroup := joinLessonData.GetUserGroup().String()
		lessonID := joinLessonData.GetLessonId()
		userID := joinLessonData.GetUserId()

		if userGroup == constant.UserGroupStudent {
			state := &repo.LessonMemberStateDTO{}
			database.AllNullEntity(state)
			now := time.Now()

			if err := multierr.Combine(
				state.LessonID.Set(lessonID),
				state.UserID.Set(userID),
				state.StateType.Set(domain.LearnerStateTypeChat),
				state.CreatedAt.Set(now),
				state.UpdatedAt.Set(now),
				state.BoolValue.Set(true),
			); err != nil {
				actualErr := fmt.Errorf("failed to set lesson member state, lesson %s user %s: %w", lessonID, userID, err)
				l.Logger.Error("[LessonDefaultChatStateEvent]: ", zap.Error(actualErr))
				return false, actualErr
			}

			if err := l.LessonMemberRepo.InsertLessonMemberState(ctx, conn, state); err != nil {
				actualErr := fmt.Errorf("error in LessonMemberRepo.InsertLessonMemberState, lesson %s user %s: %w", lessonID, userID, err)
				l.Logger.Error("[LessonDefaultChatStateEvent]: ", zap.Error(actualErr))
				return false, actualErr
			}
		} else {
			if err := l.LessonMemberRepo.InsertMissingLessonMemberStateByState(ctx, conn, lessonID, domain.LearnerStateTypeChat,
				&repo.StateValueDTO{
					BoolValue:        database.Bool(true),
					StringArrayValue: database.TextArray([]string{}),
				},
			); err != nil {
				actualErr := fmt.Errorf("error in LessonMemberRepo.InsertMissingLessonMemberStateByState, lesson %s: %w", lessonID, err)
				l.Logger.Error("[LessonDefaultChatStateEvent]: ", zap.Error(actualErr))
				return false, actualErr
			}
		}
	}

	return true, nil
}
