package virtualclassroom

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/helper"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (s *suite) returnsStatusCode(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	if stt.Code().String() != arg1 {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", arg1, stt.Code().String(), stt.Message())
	}
	return ctx, nil
}

func (s *suite) userJoinVirtualClassRoomInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = bpb.NewClassModifierServiceClient(s.BobConn).
		JoinLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), &bpb.JoinLessonRequest{
			LessonId: stepState.CurrentLessonID,
		})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserShareAMaterialWithTypeIsVideoInVirtualClassroom(ctx context.Context) (context.Context, error) {
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
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand{
				MediaId: videoMediaID,
				State: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand_VideoState{
					VideoState: &vpb.VirtualClassroomState_CurrentMaterial_VideoState{
						CurrentTime: durationpb.New(2 * time.Minute),
						PlayerState: vpb.PlayerState_PLAYER_STATE_PLAYING,
					},
				},
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(s.CommonSuite.SignedCtx(ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserShareAMaterialWithTypeIsPdfInLiveLessonRoom(ctx context.Context) (context.Context, error) {
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

	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand{
				MediaId: pdfMediaID,
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(s.CommonSuite.SignedCtx(ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserStopSharingMaterialInVirtualClassRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id:      stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_StopSharingMaterial{},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(s.CommonSuite.SignedCtx(ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserStopSharingMaterialInVirtualClassRoomInBob(ctx context.Context) (context.Context, error) {
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

func (s *suite) UserPauseVideoInLiveLessonRoom(ctx context.Context) (context.Context, error) {
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
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand{
				MediaId: videoMediaID,
				State: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand_VideoState{
					VideoState: &vpb.VirtualClassroomState_CurrentMaterial_VideoState{
						CurrentTime: durationpb.New(10 * time.Minute),
						PlayerState: vpb.PlayerState_PLAYER_STATE_PAUSE,
					},
				},
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(s.CommonSuite.SignedCtx(ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) UserResumeVideoInLiveLessonRoom(ctx context.Context) (context.Context, error) {
	return s.UserShareAMaterialWithTypeIsVideoInVirtualClassroom(ctx)
}

func (s *suite) UserStopVideoInVirtualClassroom(ctx context.Context) (context.Context, error) {
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

	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand{
				MediaId: videoMediaID,
				State: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand_VideoState{
					VideoState: &vpb.VirtualClassroomState_CurrentMaterial_VideoState{
						CurrentTime: durationpb.New(10 * time.Minute),
						PlayerState: vpb.PlayerState_PLAYER_STATE_ENDED,
					},
				},
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(s.CommonSuite.SignedCtx(ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateCurrentMaterialState(ctx context.Context) (*vpb.GetLiveLessonStateResponse, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return nil, stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return nil, err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.LessonId); err != nil {
		return nil, err
	}

	if res.CurrentTime.AsTime().IsZero() {
		return nil, fmt.Errorf("expected lesson's current time but got empty")
	}

	if res.CurrentMaterial == nil {
		return nil, fmt.Errorf("expected current material but got empty")
	}

	req := stepState.Request.(*vpb.ModifyVirtualClassroomStateRequest)
	if res.CurrentMaterial.MediaId != req.GetShareAMaterial().MediaId {
		return nil, fmt.Errorf("expected media %s but got %s", req.GetShareAMaterial().MediaId, res.CurrentMaterial.MediaId)
	}

	if res.CurrentMaterial.Data == nil {
		return nil, fmt.Errorf("expected media's data but got empty")
	}
	return res, nil
}

func (s *suite) userGetCurrentMaterialStateOfVirtualClassRoomIsPdf(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res, err := s.validateCurrentMaterialState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.CurrentMaterial.State != nil {
		if _, ok := res.CurrentMaterial.State.(*vpb.VirtualClassroomState_CurrentMaterial_PdfState); !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("current meterial is not pdf type")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentMaterialStateOfVirtualClassRoomIsVideo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res, err := s.validateCurrentMaterialState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.CurrentMaterial.GetVideoState() == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected video state but got empty")
	}

	req := stepState.Request.(*vpb.ModifyVirtualClassroomStateRequest)
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

func (s *suite) userGetCurrentMaterialStateOfVirtualClassRoomIsVideoInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res, err := s.validateCurrentMaterialStateInBob(ctx)
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

func (s *suite) userGetCurrentMaterialStateOfLiveLessonRoomIsEmpty(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoom(ctx, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err = isMatchLessonID(stepState.CurrentLessonID, res.LessonId); err != nil {
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

func (s *suite) userGetCurrentMaterialStateOfLiveLessonRoomIsEmptyInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoomInBob(ctx, stepState.CurrentLessonID)
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

func isMatchLessonID(expectedID, actualID string) error {
	if expectedID != actualID {
		return fmt.Errorf("expected lesson %s but got %s", expectedID, actualID)
	}
	return nil
}

func (s *suite) userShareAMaterialWithTypeIsVideoInVirtualClassRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand{
				MediaId: stepState.MediaIDs[0],
				State: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand_VideoState{
					VideoState: &vpb.VirtualClassroomState_CurrentMaterial_VideoState{
						CurrentTime: durationpb.New(2 * time.Minute),
						PlayerState: vpb.PlayerState_PLAYER_STATE_PLAYING,
					},
				},
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(s.CommonSuite.SignedCtx(ctx), req)
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userShareAMaterialWithTypeIsVideoInVirtualClassRoomInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
			ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
				MediaId: stepState.MediaIDs[0],
				State: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_VideoState{
					VideoState: &bpb.LiveLessonState_CurrentMaterial_VideoState{
						CurrentTime: durationpb.New(2 * time.Minute),
						PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
					},
				},
			},
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userShareAMaterialWithTypeIsPdfInVirtualClassRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand{
				MediaId: stepState.MediaIDs[0],
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(s.CommonSuite.SignedCtx(ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userSignedAsStudentWhoBelongToLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	stepState.AuthToken, err = s.CommonSuite.GenerateExchangeToken(stepState.StudentIds[0], constant.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = stepState.StudentIds[0]
	stepState.CurrentStudentID = stepState.StudentIds[0]
	stepState.CurrentUserGroup = constant.UserGroupStudent

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) haveAUncompletedVirtualClassRoomLog(ctx context.Context, arg1, arg2, arg3, arg4 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	expectedJoinedAttendees, err := strconv.Atoi(arg1)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected joined attendees need a number")
	}
	expectedNumberOfTimesGettingRoomState, err := strconv.Atoi(arg2)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of times getting room state need a number")
	}
	expectedNumberOfTimesUpdatingRoomState, err := strconv.Atoi(arg3)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of times updating room state need a number")
	}
	expectedNumberOfTimesReconnection, err := strconv.Atoi(arg4)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of times reconnection room state need a number")
	}

	return s.haveAVirtualClassRoomLog(ctx, expectedJoinedAttendees, expectedNumberOfTimesGettingRoomState, expectedNumberOfTimesUpdatingRoomState, expectedNumberOfTimesReconnection, false)
}

func (s *suite) haveAVirtualClassRoomLog(ctx context.Context, expectedJoinedAttendees, expectedNumberOfTimesGettingRoomState, expectedNumberOfTimesUpdatingRoomState, expectedNumberOfTimesReconnection int, isCompleted bool) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	logRepo := bob_repo.VirtualClassroomLogRepo{}
	actual, err := logRepo.GetLatestByLessonID(ctx, s.CommonSuite.LessonmgmtDBTrace, database.Text(stepState.CurrentLessonID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("logRepo.GetLatestByLessonID: %w", err)
	}

	expectedLogID := stepState.CurrentVirtualClassroomLogID
	if len(expectedLogID) != 0 {
		if expectedLogID != actual.LogID.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected log id %s but got %s", expectedLogID, actual.LogID.String)
		}
	}
	if actual.IsCompleted.Bool != isCompleted {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected this log have is_completed field: %v but got %v", isCompleted, actual.IsCompleted.Bool)
	}
	if expectedJoinedAttendees != len(actual.AttendeeIDs.Elements) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected number of joined attendees is %d but got %d", expectedJoinedAttendees, len(actual.AttendeeIDs.Elements))
	}
	if int32(expectedNumberOfTimesGettingRoomState) != actual.TotalTimesGettingRoomState.Int {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected number of times getting room state is %d but got %d", expectedNumberOfTimesGettingRoomState, actual.TotalTimesGettingRoomState.Int)
	}
	if int32(expectedNumberOfTimesUpdatingRoomState) != actual.TotalTimesUpdatingRoomState.Int {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected number of times updating room state is %d but got %d", expectedNumberOfTimesUpdatingRoomState, actual.TotalTimesUpdatingRoomState.Int)
	}
	if int32(expectedNumberOfTimesReconnection) != actual.TotalTimesReconnection.Int {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected number of times reconnection is %d but got %d", expectedNumberOfTimesReconnection, actual.TotalTimesReconnection.Int)
	}

	stepState.CurrentVirtualClassroomLogID = actual.LogID.String
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userShareAMaterialWithTypeIsAudioInVirtualClassRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.ModifyVirtualClassroomStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &vpb.ModifyVirtualClassroomStateRequest_ShareAMaterial{
			ShareAMaterial: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand{
				MediaId: stepState.MediaIDs[0],
				State: &vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand_AudioState{
					AudioState: &vpb.VirtualClassroomState_CurrentMaterial_AudioState{
						CurrentTime: durationpb.New(13 * time.Second),
						PlayerState: vpb.PlayerState_PLAYER_STATE_PLAYING,
					},
				},
			},
		},
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		ModifyVirtualClassroomState(s.CommonSuite.SignedCtx(ctx), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentMaterialStateOfVirtualClassRoomIsAudio(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res, err := s.validateCurrentMaterialState(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.CurrentMaterial.GetAudioState() == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected audio state but got empty")
	}

	req := stepState.Request.(*vpb.ModifyVirtualClassroomStateRequest)
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

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validateCurrentMaterialStateInBob(ctx context.Context) (*bpb.LiveLessonStateResponse, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return nil, stepState.ResponseErr
	}

	res, err := s.GetCurrentStateOfLiveLessonRoomInBob(ctx, stepState.CurrentLessonID)
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

func (s *suite) userShareAMaterialWithTypeIsAudioInVirtualClassRoomInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &bpb.ModifyLiveLessonStateRequest{
		Id: stepState.CurrentLessonID,
		Command: &bpb.ModifyLiveLessonStateRequest_ShareAMaterial{
			ShareAMaterial: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand{
				MediaId: stepState.MediaIDs[0],
				State: &bpb.ModifyLiveLessonStateRequest_CurrentMaterialCommand_AudioState{
					AudioState: &bpb.LiveLessonState_CurrentMaterial_AudioState{
						CurrentTime: durationpb.New(13 * time.Second),
						PlayerState: bpb.PlayerState_PLAYER_STATE_PLAYING,
					},
				},
			},
		},
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.BobConn).
		ModifyLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetCurrentMaterialStateOfVirtualClassRoomIsAudioInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	res, err := s.validateCurrentMaterialStateInBob(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.CurrentMaterial.GetAudioState() == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected audio state but got empty")
	}

	req := stepState.Request.(*bpb.ModifyLiveLessonStateRequest)
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

	return StepStateToContext(ctx, stepState), nil
}
