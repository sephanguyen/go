package lessonmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"google.golang.org/protobuf/types/known/durationpb"
)

func (s *Suite) UserShareAMaterialWithTypeIsVideoInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var videoMediaID string
	for _, media := range stepState.Medias {
		if media.Type == pb.MEDIA_TYPE_VIDEO {
			videoMediaID = media.MediaId
			break
		}
	}
	if len(videoMediaID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("there are not any media which have type is video")
	}

	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
			ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
				MediaId: videoMediaID,
				State: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_VideoState{
					VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
						CurrentTime: durationpb.New(2 * time.Minute),
						PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
					},
				},
			},
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Connections.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserShareAMaterialWithTypeIsPdfInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var pdfMediaID string
	for _, media := range stepState.Medias {
		if media.Type == pb.MEDIA_TYPE_PDF {
			pdfMediaID = media.MediaId
		}
	}
	if len(pdfMediaID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("there are not any media which have type is pdf")
	}

	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
			ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
				MediaId: pdfMediaID,
			},
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Connections.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserStopSharingMaterialInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_StopSharingMaterial{},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserPauseVideoInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var videoMediaID string
	for _, media := range stepState.Medias {
		if media.Type == pb.MEDIA_TYPE_VIDEO {
			videoMediaID = media.MediaId
			break
		}
	}
	if len(videoMediaID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("there are not any media which have type is video")
	}

	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
			ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
				MediaId: videoMediaID,
				State: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_VideoState{
					VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
						CurrentTime: durationpb.New(10 * time.Minute),
						PlayerState: bpb.PlayerState_PLAYER_STATE_PAUSE,
					},
				},
			},
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Connections.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserResumeVideoInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	return s.UserShareAMaterialWithTypeIsVideoInLiveLessonRoom(ctx)
}

func (s *Suite) UserStopVideoInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var videoMediaID string
	for _, media := range stepState.Medias {
		if media.Type == pb.MEDIA_TYPE_VIDEO {
			videoMediaID = media.MediaId
			break
		}
	}
	if len(videoMediaID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("there are not any media which have type is video")
	}

	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
			ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
				MediaId: videoMediaID,
				State: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_VideoState{
					VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
						CurrentTime: durationpb.New(10 * time.Minute),
						PlayerState: bpb.PlayerState_PLAYER_STATE_ENDED,
					},
				},
			},
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Connections.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) validateCurrentMaterialState(ctx context.Context) (*bpb.LiveLessonStateResponse, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return nil, stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return nil, err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.Id); err != nil {
		return nil, err
	}

	if res.CurrentTime.AsTime().IsZero() {
		return nil, fmt.Errorf("expected lesson's current time but got empty")
	}

	if res.CurrentMaterial == nil {
		return nil, fmt.Errorf("expected current material but got empty")
	}

	req := stepState.Request.(*bpb.ModifyLiveLessonStateRequest)
	if res.CurrentMaterial.MediaId != req.GetShareAMaterial().MediaId {
		return nil, fmt.Errorf("expected media %s but got %s", req.GetShareAMaterial().MediaId, res.CurrentMaterial.MediaId)
	}

	if res.CurrentMaterial.Data == nil {
		return nil, fmt.Errorf("expected media's data but got empty")
	}
	return res, nil
}

func (s *Suite) userGetCurrentMaterialStateOfLiveLessonRoomIsPdf(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res, err := s.validateCurrentMaterialState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.CurrentMaterial.State != nil {
		if _, ok := res.CurrentMaterial.State.(*bpb.LiveLessonState_CurrentMaterial_PdfState); !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("current meterial is not pdf type")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetCurrentMaterialStateOfLiveLessonRoomIsVideo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res, err := s.validateCurrentMaterialState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.CurrentMaterial.GetVideoState() == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected video state but got empty")
	}

	req := stepState.Request.(*bpb.ModifyLiveLessonStateRequest)
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

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetCurrentMaterialStateOfLiveLessonRoomIsEmpty(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.Id); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.CurrentTime.AsTime().IsZero() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson's current time but got empty")
	}

	if res.CurrentMaterial != nil &&
		(len(res.CurrentMaterial.MediaId) != 0 || res.CurrentMaterial.State != nil || res.CurrentMaterial.Data != nil) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected current material be empty but got %v", res.CurrentMaterial)
	}

	return StepStateToContext(ctx, stepState), nil
}
