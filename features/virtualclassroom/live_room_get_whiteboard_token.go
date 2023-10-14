package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	lr_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure/repo"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userGetsWhiteboardTokenForANewChannel(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentChannelName = fmt.Sprintf("channel-name-%s", s.NewID())

	return s.userGetsWhiteboardToken(StepStateToContext(ctx, stepState), stepState.CurrentChannelName)
}

func (s *suite) userGetsWhiteboardTokenForAnExistingChannel(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return s.userGetsWhiteboardToken(StepStateToContext(ctx, stepState), stepState.CurrentChannelName)
}

func (s *suite) userGetsWhiteboardToken(ctx context.Context, channelName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomReaderServiceClient(s.VirtualClassroomConn).
		GetWhiteboardToken(helper.GRPCContext(ctx, "token", stepState.AuthToken), &vpb.GetWhiteboardTokenRequest{
			ChannelName: channelName,
		})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userReceivesWhiteboardTokenAndOtherChannelDetails(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.GetWhiteboardTokenResponse)

	liveRoomRepo := lr_repo.LiveRoomRepo{}
	actualLiveRoom, err := liveRoomRepo.GetLiveRoomByChannelName(ctx, s.CommonSuite.LessonmgmtDBTrace, stepState.CurrentChannelName)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("liveRoomRepo.GetLiveRoomByChannelName: %w", err)
	}

	if len(response.ChannelId) == 0 || response.ChannelId != actualLiveRoom.ChannelID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect channel id (%s) when get whiteboard token", response.ChannelId)
	}
	stepState.CurrentChannelID = actualLiveRoom.ChannelID

	if len(response.RoomId) == 0 || response.RoomId != actualLiveRoom.WhiteboardRoomID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect room id (%s) when get whiteboard token", response.RoomId)
	}

	if len(response.WhiteboardToken) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing whiteboard token when get whiteboard token")
	}

	if len(response.WhiteboardAppId) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing whiteboard app ID when get whiteboard token")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) existingLiveRoomHasNoWhiteboardRoomID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	channelName := stepState.CurrentChannelName
	query := `UPDATE live_room 
		SET whiteboard_room_id = NULL, updated_at = now()
		WHERE channel_name = $1`

	cmdTag, err := s.CommonSuite.LessonmgmtDBTrace.Exec(ctx, query, &channelName)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to update room ID of channel %s db.Exec: %w", channelName, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no room id was updated for channel %s", channelName)
	}

	return StepStateToContext(ctx, stepState), nil
}
