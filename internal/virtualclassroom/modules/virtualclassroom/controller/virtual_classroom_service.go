package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/metrics"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/objectutils"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/controller"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	virtual_lesson_controller "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtuallesson/controller"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type VirtualClassroomModifierService struct {
	WrapperDBConnection *support.WrapperDBConnection
	JSM                 nats.JetStreamManagement
	Cfg                 configurations.Config

	VirtualClassRoomLogService    *controller.VirtualClassRoomLogService
	VirtualClassroomHandledMetric *controller.VirtualClassroomHandledMetric

	LiveLessonCommand *commands.LiveLessonCommand

	VirtualLessonRepo        infrastructure.VirtualLessonRepo
	LessonGroupRepo          infrastructure.LessonGroupRepo
	LessonMemberRepo         infrastructure.LessonMemberRepo
	LessonRoomStateRepo      infrastructure.LessonRoomStateRepo
	VirtualLessonPollingRepo infrastructure.VirtualLessonPollingRepo
	StudentsRepo             infrastructure.StudentsRepo
}

func getCommand(req *vpb.ModifyVirtualClassroomStateRequest, userID string) (commands.StateModifyCommand, error) {
	vClassCommand := &commands.VirtualClassroomCommand{}
	var command commands.StateModifyCommand
	switch req.Command.(type) {
	case *vpb.ModifyVirtualClassroomStateRequest_ShareAMaterial:
		t := &commands.ShareMaterialCommand{
			VirtualClassroomCommand: vClassCommand,
			State: &domain.CurrentMaterial{
				MediaID: req.GetShareAMaterial().MediaId,
			},
		}

		switch req.GetShareAMaterial().State.(type) {
		case *vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand_VideoState:
			tplState := req.GetShareAMaterial().GetVideoState()
			t.State.VideoState = &domain.VideoState{
				PlayerState: domain.PlayerState(tplState.PlayerState.String()),
			}
			if tplState.CurrentTime != nil {
				t.State.VideoState.CurrentTime = domain.Duration(tplState.CurrentTime.AsDuration())
			}
		case *vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand_PdfState:
			break
		case *vpb.ModifyVirtualClassroomStateRequest_CurrentMaterialCommand_AudioState:
			audioState := req.GetShareAMaterial().GetAudioState()
			t.State.AudioState = &domain.AudioState{
				PlayerState: domain.PlayerState(audioState.PlayerState.String()),
			}
			if audioState.CurrentTime != nil {
				t.State.AudioState.CurrentTime = domain.Duration(audioState.CurrentTime.AsDuration())
			}
		}

		command = t

	case *vpb.ModifyVirtualClassroomStateRequest_StopSharingMaterial:
		command = &commands.StopSharingMaterialCommand{
			VirtualClassroomCommand: &commands.VirtualClassroomCommand{
				CommanderID: userID,
				LessonID:    req.Id},
		}
	case *vpb.ModifyVirtualClassroomStateRequest_FoldHandAll:
		command = &commands.FoldHandAllCommand{
			VirtualClassroomCommand: vClassCommand,
		}
	case *vpb.ModifyVirtualClassroomStateRequest_FoldUserHand:
		command = &commands.UpdateHandsUpCommand{
			VirtualClassroomCommand: vClassCommand,
			UserID:                  req.GetFoldUserHand(),
			State: &domain.UserHandsUp{
				Value: false,
			},
		}
	case *vpb.ModifyVirtualClassroomStateRequest_RaiseHand:
		command = &commands.UpdateHandsUpCommand{
			VirtualClassroomCommand: vClassCommand,
			UserID:                  userID,
			State: &domain.UserHandsUp{
				Value: true,
			},
		}
	case *vpb.ModifyVirtualClassroomStateRequest_HandOff:
		command = &commands.UpdateHandsUpCommand{
			VirtualClassroomCommand: vClassCommand,
			UserID:                  userID,
			State: &domain.UserHandsUp{
				Value: false,
			},
		}
	case *vpb.ModifyVirtualClassroomStateRequest_AnnotationEnable:
		Learners := req.GetAnnotationEnable().Learners
		command = &commands.UpdateAnnotationCommand{
			VirtualClassroomCommand: vClassCommand,
			UserIDs:                 Learners,
			State: &domain.UserAnnotation{
				Value: true,
			},
		}
	case *vpb.ModifyVirtualClassroomStateRequest_AnnotationDisable:
		Learners := req.GetAnnotationDisable().Learners
		command = &commands.UpdateAnnotationCommand{
			VirtualClassroomCommand: vClassCommand,
			UserIDs:                 Learners,
			State: &domain.UserAnnotation{
				Value: false,
			},
		}
	case *vpb.ModifyVirtualClassroomStateRequest_AnnotationDisableAll:
		command = &commands.DisableAllAnnotationCommand{
			VirtualClassroomCommand: vClassCommand,
		}
	case *vpb.ModifyVirtualClassroomStateRequest_StartPolling:
		Options := domain.CurrentPollingOptions{}
		for _, option := range objectutils.SafeGetObject(req.GetStartPolling).Options {
			Options = append(Options, &domain.CurrentPollingOption{
				Answer:    option.Answer,
				IsCorrect: option.IsCorrect,
				Content:   option.Content,
			})
		}
		command = &commands.StartPollingCommand{
			VirtualClassroomCommand: vClassCommand,
			Options:                 Options,
			Question:                req.GetStartPolling().GetQuestion(),
		}
	case *vpb.ModifyVirtualClassroomStateRequest_StopPolling:
		command = &commands.StopPollingCommand{
			VirtualClassroomCommand: vClassCommand,
		}
	case *vpb.ModifyVirtualClassroomStateRequest_EndPolling:
		command = &commands.EndPollingCommand{
			VirtualClassroomCommand: vClassCommand,
		}
	case *vpb.ModifyVirtualClassroomStateRequest_SubmitPollingAnswer:
		command = &commands.SubmitPollingAnswerCommand{
			VirtualClassroomCommand: vClassCommand,
			UserID:                  userID,
			Answers:                 req.GetSubmitPollingAnswer().StringArrayValue,
		}
	case *vpb.ModifyVirtualClassroomStateRequest_SharePolling:
		command = &commands.SharePollingCommand{
			VirtualClassroomCommand: vClassCommand,
			IsShared:                req.GetSharePolling(),
		}
	case *vpb.ModifyVirtualClassroomStateRequest_Spotlight_:
		command = &commands.SpotlightCommand{
			VirtualClassroomCommand: vClassCommand,
			SpotlightedUser:         req.GetSpotlight().GetUserId(),
			IsEnable:                req.GetSpotlight().GetIsSpotlight(),
		}
	case *vpb.ModifyVirtualClassroomStateRequest_WhiteboardZoomState_:
		zoomStatePb := req.GetWhiteboardZoomState()
		zoomState := &domain.WhiteboardZoomState{
			PdfScaleRatio: zoomStatePb.GetPdfScaleRatio(),
			CenterX:       zoomStatePb.GetCenterX(),
			CenterY:       zoomStatePb.GetCenterY(),
			PdfWidth:      zoomStatePb.GetPdfWidth(),
			PdfHeight:     zoomStatePb.GetPdfHeight(),
		}

		command = &commands.WhiteboardZoomStateCommand{
			VirtualClassroomCommand: vClassCommand,
			WhiteboardZoomState:     zoomState,
		}
	case *vpb.ModifyVirtualClassroomStateRequest_ChatEnable:
		Learners := req.GetChatEnable().Learners
		command = &commands.UpdateChatCommand{
			VirtualClassroomCommand: vClassCommand,
			UserIDs:                 Learners,
			State: &domain.UserChat{
				Value: true,
			},
		}
	case *vpb.ModifyVirtualClassroomStateRequest_ChatDisable:
		Learners := req.GetChatDisable().Learners
		command = &commands.UpdateChatCommand{
			VirtualClassroomCommand: vClassCommand,
			UserIDs:                 Learners,
			State: &domain.UserChat{
				Value: false,
			},
		}
	case *vpb.ModifyVirtualClassroomStateRequest_UpsertSessionTime:
		command = &commands.UpsertSessionTimeCommand{
			VirtualClassroomCommand: vClassCommand,
		}
	default:
		return nil, status.Error(codes.Internal, fmt.Sprintf("unhandled state type %T", req.Command))
	}
	command.InitBasicData(userID, req.Id)
	return command, nil
}

func (v *VirtualClassroomModifierService) ModifyVirtualClassroomState(ctx context.Context, req *vpb.ModifyVirtualClassroomStateRequest) (*vpb.ModifyVirtualClassroomStateResponse, error) {
	conn, err := v.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	if len(strings.TrimSpace(req.Id)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Lesson ID can't empty")
	}
	// get lesson
	reader := &virtual_lesson_controller.VirtualLessonReaderService{
		WrapperDBConnection: v.WrapperDBConnection,
		VirtualLessonRepo:   v.VirtualLessonRepo,
	}

	lesson, err := reader.GetVirtualLessonByID(ctx, req.Id, virtual_lesson_controller.IncludeLearnerIDs(), virtual_lesson_controller.IncludeTeacherIDs())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	userID := interceptors.UserIDFromContext(ctx)
	permissionCommandChecker := commands.Create(&commands.ConfigPermissionCommandChecker{
		Lesson:              lesson,
		WrapperDBConnection: v.WrapperDBConnection,
		StudentsRepo:        v.StudentsRepo,
		Ctx:                 ctx,
	})

	virtualClassRoomDispatcher := commands.NewDispatcher(&commands.VirtualClassroomDispatcherConfig{Ctx: ctx,
		LessonGroupRepo:          v.LessonGroupRepo,
		VirtualLessonRepo:        v.VirtualLessonRepo,
		LessonMemberRepo:         v.LessonMemberRepo,
		VirtualLessonPollingRepo: v.VirtualLessonPollingRepo,
		LessonRoomStateRepo:      v.LessonRoomStateRepo,
		DB:                       conn,
		PermissionChecker:        permissionCommandChecker,
	})
	command, errGetCommand := getCommand(req, userID)
	if errGetCommand != nil {
		return nil, err
	}
	if err := virtualClassRoomDispatcher.CheckPermissionAndDispatch(command); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// log for virtual classroom
	if err = v.VirtualClassRoomLogService.LogWhenUpdateRoomState(ctx, req.Id); err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"VirtualClassRoomLogService.LogWhenUpdateRoomState: could not log this activity",
			zap.String("lesson_id", req.Id),
			zap.String("user_ID", userID),
			zap.Error(err),
		)
	}

	return &vpb.ModifyVirtualClassroomStateResponse{}, nil
}

func (v *VirtualClassroomModifierService) JoinLiveLesson(ctx context.Context, req *vpb.JoinLiveLessonRequest) (*vpb.JoinLiveLessonResponse, error) {
	if len(strings.TrimSpace(req.LessonId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Lesson ID can't empty")
	}

	userID := interceptors.UserIDFromContext(ctx)

	response, err := v.LiveLessonCommand.JoinLiveLesson(ctx, req.LessonId, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = v.PublishLessonEvt(ctx, &bpb.EvtLesson{
		Message: &bpb.EvtLesson_JoinLesson_{
			JoinLesson: &bpb.EvtLesson_JoinLesson{
				LessonId:  req.LessonId,
				UserGroup: cpb.UserGroup(cpb.UserGroup_value[response.UserGroup]),
				UserId:    userID,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("PublishLessonEvt: error joining lesson %s: %w", req.LessonId, err)
	}

	// log for virtual classroom
	createdNewLog, err := v.VirtualClassRoomLogService.LogWhenAttendeeJoin(ctx, req.LessonId, userID)
	if err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"VirtualClassRoomLogService.LogWhenAttendeeJoin: could not log this activity",
			zap.String("lesson_id", req.LessonId),
			zap.String("user_ID", userID),
			zap.Error(err),
		)
	}
	if createdNewLog {
		v.VirtualClassroomHandledMetric.ObserveWhenStartRoom(ctx)
	}

	return &vpb.JoinLiveLessonResponse{
		RoomId:               response.RoomID,
		StreamToken:          response.StreamToken,
		WhiteboardToken:      response.WhiteboardToken,
		VideoToken:           response.VideoToken,
		StmToken:             response.StmToken,
		AgoraAppId:           v.Cfg.Agora.AppID,
		WhiteboardAppId:      v.Cfg.Whiteboard.AppID,
		ScreenRecordingToken: response.ScreenRecordingToken,
	}, nil
}

func (v *VirtualClassroomModifierService) LeaveLiveLesson(ctx context.Context, req *vpb.LeaveLiveLessonRequest) (*vpb.LeaveLiveLessonResponse, error) {
	if len(strings.TrimSpace(req.LessonId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Lesson ID can't empty")
	}
	if len(strings.TrimSpace(req.UserId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "User ID can't empty")
	}

	canLeaveLiveLesson, err := v.LiveLessonCommand.CanLeaveLiveLesson(ctx, req.LessonId, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !canLeaveLiveLesson {
		return nil, status.Errorf(codes.PermissionDenied, "user %s can only leave by him/her self from lesson %s", req.UserId, req.LessonId)
	}

	err = v.PublishLessonEvt(ctx, &bpb.EvtLesson{
		Message: &bpb.EvtLesson_LeaveLesson_{
			LeaveLesson: &bpb.EvtLesson_LeaveLesson{
				LessonId: req.LessonId,
				UserId:   req.UserId,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("PublishLessonEvt: error leave lesson %s: %w", req.LessonId, err)
	}

	return &vpb.LeaveLiveLessonResponse{}, nil
}

func (v *VirtualClassroomModifierService) EndLiveLesson(ctx context.Context, req *vpb.EndLiveLessonRequest) (*vpb.EndLiveLessonResponse, error) {
	if len(strings.TrimSpace(req.LessonId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Lesson ID can't empty")
	}

	userID := interceptors.UserIDFromContext(ctx)
	canEndLiveLesson, err := v.LiveLessonCommand.CanEndLiveLesson(ctx, req.LessonId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !canEndLiveLesson {
		return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("user %s could not end live lesson %s due to course is not valid", userID, req.LessonId))
	}

	if err := v.ResetAllLiveLessonStatesInternal(ctx, req.LessonId, userID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := v.LiveLessonCommand.EndLiveLesson(ctx, req.LessonId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// logging and observe metrics
	// make sure not observe one log twice
	exceptedID := ""
	log, _ := v.VirtualClassRoomLogService.GetCompletedLogByLesson(ctx, req.LessonId)
	if log != nil {
		exceptedID = log.LogID.String
	}

	// log for virtual classroom
	if err := v.VirtualClassRoomLogService.LogWhenEndRoom(ctx, req.LessonId); err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"VirtualClassRoomLogService.LogWhenEndRoom: could not log this activity",
			zap.String("lesson_id", req.LessonId),
			zap.String("user_ID", userID),
			zap.Error(err),
		)
	}

	log, err = v.VirtualClassRoomLogService.GetCompletedLogByLesson(ctx, req.LessonId)
	if err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Error(
			"VirtualClassRoomLogService.GetCompletedLogByLesson: could not get log",
			zap.String("lesson_id", req.LessonId),
			zap.String("user_ID", userID),
			zap.Error(err),
		)
	}
	if log != nil && exceptedID != log.LogID.String {
		v.VirtualClassroomHandledMetric.ObserveWhenEndRoom(ctx, log)
	}

	err = v.PublishLessonEvt(ctx, &bpb.EvtLesson{
		Message: &bpb.EvtLesson_EndLiveLesson_{
			EndLiveLesson: &bpb.EvtLesson_EndLiveLesson{
				LessonId: req.LessonId,
				UserId:   userID,
			},
		},
	})
	if err != nil {
		logger := ctxzap.Extract(ctx)
		logger.Warn(
			"PublishLessonEvt: error end live lesson",
			zap.String("lesson_id", req.LessonId),
			zap.String("user_ID", userID),
			zap.Error(err),
		)
	}

	return &vpb.EndLiveLessonResponse{}, nil
}

func (v *VirtualClassroomModifierService) PreparePublish(ctx context.Context, req *vpb.PreparePublishRequest) (*vpb.PreparePublishResponse, error) {
	if len(strings.TrimSpace(req.LessonId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Lesson ID can't empty")
	}
	if len(strings.TrimSpace(req.LearnerId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Learner ID can't empty")
	}

	publishStatus, err := v.LiveLessonCommand.PreparePublish(ctx, req.LessonId, req.LearnerId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vpb.PreparePublishResponse{
		Status: vpb.PrepareToPublishStatus(vpb.PrepareToPublishStatus_value[string(publishStatus)]),
	}, nil
}

func (v *VirtualClassroomModifierService) Unpublish(ctx context.Context, req *vpb.UnpublishRequest) (*vpb.UnpublishResponse, error) {
	if len(strings.TrimSpace(req.LessonId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Lesson ID can't empty")
	}
	if len(strings.TrimSpace(req.LearnerId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Learner ID can't empty")
	}

	unpublishStatus, err := v.LiveLessonCommand.Unpublish(ctx, req.LessonId, req.LearnerId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vpb.UnpublishResponse{
		Status: vpb.UnpublishStatus(vpb.UnpublishStatus_value[string(unpublishStatus)]),
	}, nil
}

func (v *VirtualClassroomModifierService) PublishLessonEvt(ctx context.Context, msg *bpb.EvtLesson) error {
	var msgID string
	data, _ := proto.Marshal(msg)

	msgID, err := v.JSM.PublishAsyncContext(ctx, constants.SubjectLessonUpdated, data)
	if err != nil {
		return fmt.Errorf("PublishLessonEvt s.JSM.PublishAsyncContext subject Lesson.Updated failed, msgID: %s, %w", msgID, err)
	}
	return nil
}

func (v *VirtualClassroomModifierService) RegisterMetric(collector metrics.MetricCollector) {
	v.VirtualClassroomHandledMetric = controller.RegisterVirtualClassroomHandledMetric(collector)
}

func (v *VirtualClassroomModifierService) ResetAllLiveLessonStatesInternal(ctx context.Context, lessonID, userID string) error {
	conn, err := v.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return err
	}

	reader := &virtual_lesson_controller.VirtualLessonReaderService{
		WrapperDBConnection: v.WrapperDBConnection,
		VirtualLessonRepo:   v.VirtualLessonRepo,
	}

	lesson, err := reader.GetVirtualLessonByID(ctx, lessonID, virtual_lesson_controller.IncludeLearnerIDs(), virtual_lesson_controller.IncludeTeacherIDs())
	if err != nil {
		return fmt.Errorf("error in reader.GetVirtualLessonByID %s: %w", lessonID, err)
	}

	permissionCommandChecker := commands.Create(&commands.ConfigPermissionCommandChecker{Lesson: lesson,
		WrapperDBConnection: v.WrapperDBConnection,
		StudentsRepo:        v.StudentsRepo,
		Ctx:                 ctx})

	virtualClassRoomDispatcher := commands.NewDispatcher(&commands.VirtualClassroomDispatcherConfig{Ctx: ctx,
		LessonGroupRepo:          v.LessonGroupRepo,
		VirtualLessonRepo:        v.VirtualLessonRepo,
		LessonMemberRepo:         v.LessonMemberRepo,
		VirtualLessonPollingRepo: v.VirtualLessonPollingRepo,
		LessonRoomStateRepo:      v.LessonRoomStateRepo,
		DB:                       conn,
		PermissionChecker:        permissionCommandChecker})

	command := &commands.ResetAllStatesCommand{
		VirtualClassroomCommand: &commands.VirtualClassroomCommand{
			CommanderID: userID,
			LessonID:    lessonID,
		},
	}
	if err := virtualClassRoomDispatcher.CheckPermissionAndDispatch(command); err != nil {
		return fmt.Errorf("error in reset all states for lesson %s: %w", lessonID, err)
	}

	return nil
}
