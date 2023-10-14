package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/objectutils"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/queries"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/controller"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	vc_infrastructure "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type LiveRoomModifierService struct {
	LessonmgmtDB        database.Ext
	WrapperDBConnection *support.WrapperDBConnection
	JSM                 nats.JetStreamManagement
	Cfg                 configurations.Config

	LiveRoomLogService *controller.LiveRoomLogService
	LiveRoomCommand    *commands.LiveRoomCommand
	LiveRoomStateQuery queries.LiveRoomStateQuery

	LiveRoomStateRepo       infrastructure.LiveRoomStateRepo
	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
	LiveRoomPoll            infrastructure.LiveRoomPoll
	StudentsRepo            vc_infrastructure.StudentsRepo
}

func (l *LiveRoomModifierService) JoinLiveRoom(ctx context.Context, req *vpb.JoinLiveRoomRequest) (*vpb.JoinLiveRoomResponse, error) {
	if len(strings.TrimSpace(req.ChannelName)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "channel name can't empty")
	}
	if len(strings.TrimSpace(req.RtmUserId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "rtm user ID can't empty")
	}

	response, err := l.LiveRoomCommand.JoinLiveRoom(ctx, req.ChannelName, req.RtmUserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	userID := interceptors.UserIDFromContext(ctx)
	if err := l.PublishLiveRoomEvent(ctx, &vpb.LiveRoomEvent{
		Message: &vpb.LiveRoomEvent_JoinLiveRoom_{
			JoinLiveRoom: &vpb.LiveRoomEvent_JoinLiveRoom{
				ChannelId:   response.ChannelID,
				ChannelName: req.ChannelName,
				UserGroup:   cpb.UserGroup(cpb.UserGroup_value[response.UserGroup]),
				UserId:      userID,
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("PublishLiveRoomEvent: error joining live room %s: %w", req.ChannelName, err)
	}

	_, err = l.LiveRoomLogService.LogWhenAttendeeJoin(ctx, response.ChannelID, userID)
	if err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"LiveRoomLogService.LogWhenAttendeeJoin: could not log this activity",
			zap.String("channel_id", response.ChannelID),
			zap.String("user_ID", userID),
			zap.Error(err),
		)
	}

	return &vpb.JoinLiveRoomResponse{
		ChannelId:            response.ChannelID,
		RoomId:               response.RoomID,
		StreamToken:          response.StreamToken,
		WhiteboardToken:      response.WhiteboardToken,
		VideoToken:           response.VideoToken,
		StmToken:             response.StmToken,
		AgoraAppId:           l.Cfg.Agora.AppID,
		WhiteboardAppId:      l.Cfg.Whiteboard.AppID,
		ScreenRecordingToken: response.ScreenRecordingToken,
	}, nil
}

func (l *LiveRoomModifierService) PublishLiveRoomEvent(ctx context.Context, msg *vpb.LiveRoomEvent) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal proto: %w", err)
	}

	msgID, err := l.JSM.PublishAsyncContext(ctx, constants.SubjectLiveRoomUpdated, data)
	if err != nil {
		return fmt.Errorf("publish live room event with subject %s failed, message id: %s, %w", constants.SubjectLiveRoomUpdated, msgID, err)
	}

	return nil
}

func getCommand(req *vpb.ModifyLiveRoomStateRequest, userID string) (commands.ModifyStateCommand, error) {
	var command commands.ModifyStateCommand
	modifyLiveRoomCommand := &commands.ModifyLiveRoomCommand{}

	switch req.Command.(type) {
	case *vpb.ModifyLiveRoomStateRequest_AnnotationEnable:
		Learners := req.GetAnnotationEnable().Learners
		command = &commands.UpdateAnnotationCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			UserIDs:               Learners,
			State: &vc_domain.UserAnnotation{
				Value: true,
			},
		}
	case *vpb.ModifyLiveRoomStateRequest_AnnotationDisable:
		Learners := req.GetAnnotationDisable().Learners
		command = &commands.UpdateAnnotationCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			UserIDs:               Learners,
			State: &vc_domain.UserAnnotation{
				Value: false,
			},
		}
	case *vpb.ModifyLiveRoomStateRequest_AnnotationDisableAll:
		command = &commands.DisableAllAnnotationCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
		}
	case *vpb.ModifyLiveRoomStateRequest_ChatEnable:
		Learners := req.GetChatEnable().Learners
		command = &commands.UpdateChatCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			UserIDs:               Learners,
			State: &vc_domain.UserChat{
				Value: true,
			},
		}
	case *vpb.ModifyLiveRoomStateRequest_ChatDisable:
		Learners := req.GetChatDisable().Learners
		command = &commands.UpdateChatCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			UserIDs:               Learners,
			State: &vc_domain.UserChat{
				Value: false,
			},
		}
	case *vpb.ModifyLiveRoomStateRequest_StartPolling:
		Options := vc_domain.CurrentPollingOptions{}
		for _, option := range objectutils.SafeGetObject(req.GetStartPolling).Options {
			Options = append(Options, &vc_domain.CurrentPollingOption{
				Answer:    option.Answer,
				IsCorrect: option.IsCorrect,
				Content:   option.Content,
			})
		}

		command = &commands.StartPollingCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			Options:               Options,
			Question:              req.GetStartPolling().GetQuestion(),
		}
	case *vpb.ModifyLiveRoomStateRequest_StopPolling:
		command = &commands.StopPollingCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
		}
	case *vpb.ModifyLiveRoomStateRequest_EndPolling:
		command = &commands.EndPollingCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
		}
	case *vpb.ModifyLiveRoomStateRequest_SubmitPollingAnswer:
		command = &commands.SubmitPollingAnswerCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			UserID:                userID,
			Answers:               req.GetSubmitPollingAnswer().StringArrayValue,
		}
	case *vpb.ModifyLiveRoomStateRequest_SharePolling:
		command = &commands.SharePollingCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			IsShared:              req.GetSharePolling(),
		}
	case *vpb.ModifyLiveRoomStateRequest_RaiseHand:
		command = &commands.UpdateHandsUpCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			UserID:                userID,
			State: &vc_domain.UserHandsUp{
				Value: true,
			},
		}
	case *vpb.ModifyLiveRoomStateRequest_HandOff:
		command = &commands.UpdateHandsUpCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			UserID:                userID,
			State: &vc_domain.UserHandsUp{
				Value: false,
			},
		}
	case *vpb.ModifyLiveRoomStateRequest_FoldUserHand:
		command = &commands.UpdateHandsUpCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			UserID:                req.GetFoldUserHand(),
			State: &vc_domain.UserHandsUp{
				Value: false,
			},
		}
	case *vpb.ModifyLiveRoomStateRequest_FoldHandAll:
		command = &commands.FoldHandAllCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
		}
	case *vpb.ModifyLiveRoomStateRequest_Spotlight_:
		command = &commands.SpotlightCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			SpotlightedUser:       req.GetSpotlight().GetUserId(),
			IsEnable:              req.GetSpotlight().GetIsSpotlight(),
		}
	case *vpb.ModifyLiveRoomStateRequest_WhiteboardZoomState_:
		zoomStatePb := req.GetWhiteboardZoomState()
		zoomState := &vc_domain.WhiteboardZoomState{
			PdfScaleRatio: zoomStatePb.GetPdfScaleRatio(),
			CenterX:       zoomStatePb.GetCenterX(),
			CenterY:       zoomStatePb.GetCenterY(),
			PdfWidth:      zoomStatePb.GetPdfWidth(),
			PdfHeight:     zoomStatePb.GetPdfHeight(),
		}

		command = &commands.WhiteboardZoomStateCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			WhiteboardZoomState:   zoomState,
		}
	case *vpb.ModifyLiveRoomStateRequest_ShareAMaterial:
		t := &commands.ShareMaterialCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
			State: &vc_domain.CurrentMaterial{
				MediaID: req.GetShareAMaterial().MediaId,
			},
		}

		switch req.GetShareAMaterial().State.(type) {
		case *vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand_VideoState:
			tplState := req.GetShareAMaterial().GetVideoState()
			t.State.VideoState = &vc_domain.VideoState{
				PlayerState: vc_domain.PlayerState(tplState.PlayerState.String()),
			}
			if tplState.CurrentTime != nil {
				t.State.VideoState.CurrentTime = vc_domain.Duration(tplState.CurrentTime.AsDuration())
			}
		case *vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand_PdfState:
			break
		case *vpb.ModifyLiveRoomStateRequest_CurrentMaterialCommand_AudioState:
			audioState := req.GetShareAMaterial().GetAudioState()
			t.State.AudioState = &vc_domain.AudioState{
				PlayerState: vc_domain.PlayerState(audioState.PlayerState.String()),
			}
			if audioState.CurrentTime != nil {
				t.State.AudioState.CurrentTime = vc_domain.Duration(audioState.CurrentTime.AsDuration())
			}
		}

		command = t
	case *vpb.ModifyLiveRoomStateRequest_StopSharingMaterial:
		command = &commands.StopSharingMaterialCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
		}
	case *vpb.ModifyLiveRoomStateRequest_UpsertSessionTime:
		command = &commands.UpsertSessionTimeCommand{
			ModifyLiveRoomCommand: modifyLiveRoomCommand,
		}
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf("unhandled state type %T", req.Command))
	}

	command.InitBasicData(userID, req.ChannelId)
	return command, nil
}

func (l *LiveRoomModifierService) ModifyLiveRoomState(ctx context.Context, req *vpb.ModifyLiveRoomStateRequest) (*vpb.ModifyLiveRoomStateResponse, error) {
	if len(strings.TrimSpace(req.ChannelId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Channel ID can't empty")
	}

	userID := interceptors.UserIDFromContext(ctx)
	permissionCommandChecker := commands.Create(&commands.ConfigPermissionCommandChecker{
		Ctx:                 ctx,
		WrapperDBConnection: l.WrapperDBConnection,
		StudentsRepo:        l.StudentsRepo,
	})

	commandDispatcher := commands.NewDispatcher(&commands.ModifyLiveRoomCommandDispatcherConfig{
		Ctx:                     ctx,
		LessonmgmtDB:            l.LessonmgmtDB,
		PermissionChecker:       permissionCommandChecker,
		LiveRoomStateRepo:       l.LiveRoomStateRepo,
		LiveRoomMemberStateRepo: l.LiveRoomMemberStateRepo,
		LiveRoomPoll:            l.LiveRoomPoll,
	})

	command, err := getCommand(req, userID)
	if err != nil {
		return nil, err
	}

	if err := commandDispatcher.CheckPermissionAndDispatch(command); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = l.LiveRoomLogService.LogWhenUpdateRoomState(ctx, req.ChannelId); err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"LiveRoomLogService.LogWhenUpdateRoomState: could not log this activity",
			zap.String("channel_id", req.ChannelId),
			zap.String("user_ID", userID),
			zap.Error(err),
		)
	}

	return &vpb.ModifyLiveRoomStateResponse{}, nil
}

func (l *LiveRoomModifierService) LeaveLiveRoom(ctx context.Context, req *vpb.LeaveLiveRoomRequest) (*vpb.LeaveLiveRoomResponse, error) {
	if len(strings.TrimSpace(req.ChannelId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Channel ID can't empty")
	}
	if len(strings.TrimSpace(req.UserId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "User ID can't empty")
	}

	liveRoom, err := l.LiveRoomStateQuery.GetLiveRoom(ctx, req.ChannelId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := l.PublishLiveRoomEvent(ctx, &vpb.LiveRoomEvent{
		Message: &vpb.LiveRoomEvent_LeaveLiveRoom_{
			LeaveLiveRoom: &vpb.LiveRoomEvent_LeaveLiveRoom{
				ChannelId:   liveRoom.ChannelID,
				ChannelName: liveRoom.ChannelName,
				UserId:      req.UserId,
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("PublishLiveRoomEvent: error leaving live room %s: %w", req.ChannelId, err)
	}

	return &vpb.LeaveLiveRoomResponse{}, nil
}

func (l *LiveRoomModifierService) ResetAllLiveRoomStatesInternal(ctx context.Context, channelID, userID string) error {
	permissionCommandChecker := commands.Create(&commands.ConfigPermissionCommandChecker{
		Ctx:                 ctx,
		WrapperDBConnection: l.WrapperDBConnection,
		StudentsRepo:        l.StudentsRepo,
	})

	commandDispatcher := commands.NewDispatcher(&commands.ModifyLiveRoomCommandDispatcherConfig{
		Ctx:                     ctx,
		LessonmgmtDB:            l.LessonmgmtDB,
		PermissionChecker:       permissionCommandChecker,
		LiveRoomStateRepo:       l.LiveRoomStateRepo,
		LiveRoomMemberStateRepo: l.LiveRoomMemberStateRepo,
		LiveRoomPoll:            l.LiveRoomPoll,
	})

	command := &commands.ResetAllStatesCommand{
		ModifyLiveRoomCommand: &commands.ModifyLiveRoomCommand{
			CommanderID: userID,
			ChannelID:   channelID,
		},
	}
	if err := commandDispatcher.CheckPermissionAndDispatch(command); err != nil {
		return fmt.Errorf("error in reset all states for channel %s: %w", channelID, err)
	}

	return nil
}

func (l *LiveRoomModifierService) EndLiveRoom(ctx context.Context, req *vpb.EndLiveRoomRequest) (*vpb.EndLiveRoomResponse, error) {
	if len(strings.TrimSpace(req.ChannelId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Channel ID can't empty")
	}

	userID := interceptors.UserIDFromContext(ctx)
	liveRoom, err := l.LiveRoomStateQuery.GetLiveRoom(ctx, req.ChannelId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := l.ResetAllLiveRoomStatesInternal(ctx, req.ChannelId, userID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := l.LiveRoomCommand.EndLiveRoom(ctx, req.ChannelId, req.LessonId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := l.LiveRoomLogService.LogWhenEndRoom(ctx, req.ChannelId); err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"LiveRoomLogService.LogWhenEndRoom: could not log this activity",
			zap.String("channel_id", req.ChannelId),
			zap.String("user_id", userID),
			zap.Error(err),
		)
	}

	if err := l.PublishLiveRoomEvent(ctx, &vpb.LiveRoomEvent{
		Message: &vpb.LiveRoomEvent_EndLiveRoom_{
			EndLiveRoom: &vpb.LiveRoomEvent_EndLiveRoom{
				ChannelId:   liveRoom.ChannelID,
				ChannelName: liveRoom.ChannelName,
				UserId:      userID,
			},
		},
	}); err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Warn(
			"PublishLiveRoomEvent: error ending live room",
			zap.String("channel_id", liveRoom.ChannelID),
			zap.String("user_id", userID),
			zap.Error(err),
		)
	}

	return &vpb.EndLiveRoomResponse{}, nil
}

func (l *LiveRoomModifierService) PreparePublishLiveRoom(ctx context.Context, req *vpb.PreparePublishLiveRoomRequest) (*vpb.PreparePublishLiveRoomResponse, error) {
	if len(strings.TrimSpace(req.ChannelId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Channel ID can't empty")
	}
	if len(strings.TrimSpace(req.LearnerId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Learner ID can't empty")
	}

	publishStatus, err := l.LiveRoomCommand.PreparePublishLiveRoom(ctx, req.ChannelId, req.LearnerId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vpb.PreparePublishLiveRoomResponse{
		Status: vpb.PrepareToPublishStatus(vpb.PrepareToPublishStatus_value[string(publishStatus)]),
	}, nil
}

func (l *LiveRoomModifierService) UnpublishLiveRoom(ctx context.Context, req *vpb.UnpublishLiveRoomRequest) (*vpb.UnpublishLiveRoomResponse, error) {
	if len(strings.TrimSpace(req.ChannelId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Channel ID can't empty")
	}
	if len(strings.TrimSpace(req.LearnerId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Learner ID can't empty")
	}

	unpublishStatus, err := l.LiveRoomCommand.UnpublishLiveRoom(ctx, req.ChannelId, req.LearnerId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vpb.UnpublishLiveRoomResponse{
		Status: vpb.UnpublishStatus(vpb.UnpublishStatus_value[string(unpublishStatus)]),
	}, nil
}
