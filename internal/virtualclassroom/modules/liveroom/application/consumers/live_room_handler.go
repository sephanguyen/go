package consumers

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type LiveRoomHandler struct {
	Logger       *zap.Logger
	LessonmgmtDB database.Ext
	JSM          nats.JetStreamManagement

	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
}

func (l *LiveRoomHandler) Handle(ctx context.Context, msg []byte) (bool, error) {
	l.Logger.Info("[LiveRoomEvent]: Received message on",
		zap.String("data", string(msg)),
		zap.String("subject", constants.SubjectLiveRoomUpdated),
		zap.String("queue", constants.QueueLiveRoom),
	)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	eventData := &vpb.LiveRoomEvent{}
	if err := proto.Unmarshal(msg, eventData); err != nil {
		l.Logger.Error("[LiveRoomEvent] Failed to parse vpb.LiveRoomEvent: ", zap.Error(err))
		return false, err
	}

	switch eventData.Message.(type) {
	case *vpb.LiveRoomEvent_JoinLiveRoom_:
		joinLiveRoomData := eventData.GetJoinLiveRoom()
		channelID := joinLiveRoomData.GetChannelId()
		userID := joinLiveRoomData.GetUserId()
		// from join live room API, user group can only be teacher or student
		userGroup := joinLiveRoomData.GetUserGroup().String()

		if userGroup == constant.UserGroupStudent {
			state := &vc_domain.StateValue{
				BoolValue: true,
			}

			// enable chat permission
			if err := l.LiveRoomMemberStateRepo.CreateLiveRoomMemberState(ctx, l.LessonmgmtDB, channelID, userID, vc_domain.LearnerStateTypeChat, state); err != nil {
				actualErr := fmt.Errorf("error in LiveRoomMemberStateRepo.CreateLiveRoomMemberState, channel %s user %s chat permission: %w", channelID, userID, err)
				l.Logger.Error("[LiveRoomEvent]: ", zap.Error(actualErr))
				return false, actualErr
			}

			// enable annotation permission
			if err := l.LiveRoomMemberStateRepo.CreateLiveRoomMemberState(ctx, l.LessonmgmtDB, channelID, userID, vc_domain.LearnerStateTypeAnnotation, state); err != nil {
				actualErr := fmt.Errorf("error in LiveRoomMemberStateRepo.CreateLiveRoomMemberState, channel %s user %s annotation permission: %w", channelID, userID, err)
				l.Logger.Error("[LiveRoomEvent]: ", zap.Error(actualErr))
				return false, actualErr
			}
		}
	default:
		l.Logger.Info(fmt.Sprintf("[LiveRoomEvent]: live room event type not supported %T", eventData.Message))
	}

	return true, nil
}
