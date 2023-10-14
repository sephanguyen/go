package virtualclassroom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"google.golang.org/protobuf/types/known/durationpb"
)

func (s *suite) userShareAMaterialWithTypeInTheLiveRoom(ctx context.Context, mediaType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.ModifyLiveRoomStateRequest{
		ChannelId: stepState.CurrentChannelID,
	}

	switch mediaType {
	case "audio":
		req.Command = &vpb.ModifyLiveRoomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand{
				MediaId: stepState.MediaIDs[0],
				State: &vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand_AudioState{
					AudioState: &vpb.VirtualClassroomState_CurrentMaterial_AudioState{
						CurrentTime: durationpb.New(13 * time.Second),
						PlayerState: vpb.PlayerState_PLAYER_STATE_PLAYING,
					},
				},
			},
		}
	case "pdf":
		req.Command = &vpb.ModifyLiveRoomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand{
				MediaId: stepState.MediaIDs[0],
			},
		}
	case "video":
		req.Command = &vpb.ModifyLiveRoomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand{
				MediaId: stepState.MediaIDs[0],
				State: &vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand_VideoState{
					VideoState: &vpb.VirtualClassroomState_CurrentMaterial_VideoState{
						CurrentTime: durationpb.New(9 * time.Second),
						PlayerState: vpb.PlayerState_PLAYER_STATE_PAUSE,
					},
				},
			},
		}
	case "empty":
		req.Command = &vpb.ModifyLiveRoomStateRequest_StopSharingMaterial{}
	default:
		return nil, fmt.Errorf("media type is not supported in the share a material step")
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		ModifyLiveRoomState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userStopSharingMaterialInTheLiveRoom(ctx context.Context) (context.Context, error) {
	return s.userShareAMaterialWithTypeInTheLiveRoom(ctx, "empty")
}

func (s *suite) userGetsCurrentMaterialStateOfLiveRoom(ctx context.Context, mediaType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.CurrentMaterial == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected current material but got empty")
	}

	req := stepState.Request.(*vpb.ModifyLiveRoomStateRequest)
	if res.CurrentMaterial.MediaId != req.GetShareAMaterial().MediaId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected media %s but got %s", req.GetShareAMaterial().MediaId, res.CurrentMaterial.MediaId)
	}

	switch mediaType {
	case "audio":
		if res.CurrentMaterial.Data == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected media's data but got empty")
		}

		if res.CurrentMaterial.GetAudioState() == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected audio state but got empty")
		}

		actualCurrentT := res.CurrentMaterial.GetAudioState().CurrentTime.AsDuration()
		expectedCurrentT := req.GetShareAMaterial().GetAudioState().CurrentTime.AsDuration()
		if actualCurrentT != expectedCurrentT {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected audio's current time %v but got %v", expectedCurrentT, actualCurrentT)
		}

		actualPlayerSt := res.CurrentMaterial.GetAudioState().PlayerState.String()
		expectedPlayerSt := req.GetShareAMaterial().GetAudioState().PlayerState.String()
		if actualPlayerSt != expectedPlayerSt {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected audio's player state %v but got %v", expectedPlayerSt, actualPlayerSt)
		}
	case "pdf":
		if res.CurrentMaterial.State != nil {
			if _, ok := res.CurrentMaterial.State.(*vpb.VirtualClassroomState_CurrentMaterial_PdfState); !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf("current meterial is not pdf type")
			}
		}
	case "video":
		if res.CurrentMaterial.Data == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected media's data but got empty")
		}

		if res.CurrentMaterial.GetVideoState() == nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected video state but got empty")
		}

		actualCurrentT := res.CurrentMaterial.GetVideoState().CurrentTime.AsDuration()
		expectedCurrentT := req.GetShareAMaterial().GetVideoState().CurrentTime.AsDuration()
		if actualCurrentT != expectedCurrentT {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected video's current time %v but got %v", expectedCurrentT, actualCurrentT)
		}

		actualPlayerSt := res.CurrentMaterial.GetVideoState().PlayerState.String()
		expectedPlayerSt := req.GetShareAMaterial().GetVideoState().PlayerState.String()
		if actualPlayerSt != expectedPlayerSt {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected video's player state %v but got %v", expectedPlayerSt, actualPlayerSt)
		}
	default:
		return nil, fmt.Errorf("media type is not supported in the validate material step")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsEmptyCurrentMaterialStateOfLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	res, err := s.GetCurrentStateOfLiveRoom(ctx, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.CurrentMaterial != nil &&
		(len(res.CurrentMaterial.MediaId) != 0 || res.CurrentMaterial.State != nil || res.CurrentMaterial.Data != nil) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected current material be empty but got %v", res.CurrentMaterial)
	}

	return StepStateToContext(ctx, stepState), nil
}
