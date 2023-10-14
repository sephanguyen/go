package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	lr_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure/repo"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userJoinsANewLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentChannelName = fmt.Sprintf("channel-name-%s", s.NewID())

	return s.userJoinLiveRoom(StepStateToContext(ctx, stepState), stepState.CurrentChannelName)
}

func (s *suite) userJoinsAnExistingLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return s.userJoinLiveRoom(StepStateToContext(ctx, stepState), stepState.CurrentChannelName)
}

func (s *suite) userJoinLiveRoom(ctx context.Context, channelName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		JoinLiveRoom(helper.GRPCContext(ctx, "token", stepState.AuthToken), &vpb.JoinLiveRoomRequest{
			ChannelName: channelName,
			RtmUserId:   stepState.CurrentUserID,
		})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userReceivesChannelAndRoomIDAndOtherTokens(ctx context.Context, user string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.JoinLiveRoomResponse)

	liveRoomRepo := lr_repo.LiveRoomRepo{}
	actualLiveRoom, err := liveRoomRepo.GetLiveRoomByChannelName(ctx, s.CommonSuite.LessonmgmtDBTrace, stepState.CurrentChannelName)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("liveRoomRepo.GetLiveRoomByChannelName: %w", err)
	}

	if len(response.ChannelId) == 0 || response.ChannelId != actualLiveRoom.ChannelID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect channel id (%s) when %s join live room", response.ChannelId, user)
	}
	stepState.CurrentChannelID = actualLiveRoom.ChannelID

	if len(response.RoomId) == 0 || response.RoomId != actualLiveRoom.WhiteboardRoomID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect room id (%s) when %s join live room", response.RoomId, user)
	}

	if len(response.WhiteboardToken) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing whiteboard token when %s join live room", user)
	}

	if len(response.WhiteboardAppId) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing whiteboard app ID when %s join live room", user)
	}

	if len(response.StmToken) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing stm token when %s join live room", user)
	}

	if len(response.StreamToken) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing stream token when %s join live room", user)
	}

	if user == "teacher" || user == "staff" {
		if len(response.VideoToken) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("missing video token when teacher join live room")
		}

		if len(response.ScreenRecordingToken) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("missing screen recording token when teacher join live room")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentWhoIsPartOfTheLiveRoom(ctx context.Context) (context.Context, error) {
	return s.userSignedAsStudentWhoBelongToLesson(ctx)
}
