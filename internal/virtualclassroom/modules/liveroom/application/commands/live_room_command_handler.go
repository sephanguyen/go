package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/queries/payloads"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
	vc_infrastructure "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
)

type LiveRoomCommand struct {
	LessonmgmtDB             database.Ext
	WrapperDBConnection      *support.WrapperDBConnection
	VideoTokenSuffix         string
	WhiteboardAppID          string
	MaximumLearnerStreamings int

	WhiteboardSvc vc_infrastructure.WhiteboardPort
	AgoraTokenSvc vc_infrastructure.AgoraTokenPort

	LiveRoomRepo            infrastructure.LiveRoomRepo
	LiveRoomStateRepo       infrastructure.LiveRoomStateRepo
	LiveRoomActivityLogRepo infrastructure.LiveRoomActivityLogRepo
	StudentsRepo            vc_infrastructure.StudentsRepo
	LessonRepo              vc_infrastructure.VirtualLessonRepo
}

type JoinLiveRoomResponse struct {
	StreamToken          string
	WhiteboardToken      string
	RoomID               string
	StmToken             string
	VideoToken           string
	ScreenRecordingToken string
	UserGroup            string
	ChannelID            string
}

func (l *LiveRoomCommand) JoinLiveRoom(ctx context.Context, channelName, rtmUserID string) (*JoinLiveRoomResponse, error) {
	liveRoom, err := l.createChannel(ctx, channelName)
	if err != nil {
		return nil, err
	}

	whiteboardToken, roomID, err := l.getWhiteboardToken(ctx, liveRoom.ChannelID, liveRoom.WhiteboardRoomID)
	if err != nil {
		return nil, err
	}
	liveRoom.WhiteboardRoomID = roomID

	userID := interceptors.UserIDFromContext(ctx)
	conn, err := l.WrapperDBConnection.GetDB(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	isUserAStudent, err := l.StudentsRepo.IsUserIDAStudent(ctx, conn, userID)
	if err != nil {
		return nil, fmt.Errorf("error in StudentsRepo.IsUserIDAStudent, user %s: %w", userID, err)
	}
	userGroup := constant.UserGroupTeacher
	if isUserAStudent {
		userGroup = constant.UserGroupStudent
	}

	var streamToken, videoToken, shareForRecordingToken, rtmToken string
	channelID := liveRoom.ChannelID
	if userGroup == constant.UserGroupStudent {
		streamToken, err = l.AgoraTokenSvc.GenerateAgoraStreamToken(channelID, userID, vc_domain.RoleSubscriber)
		if err != nil {
			return nil, fmt.Errorf("error retrieve subscribe token (user:%s, channel:%s): AgoraTokenSrv.GenerateAgoraStreamToken: %w", userID, channelID, err)
		}
	} else {
		streamToken, err = l.AgoraTokenSvc.GenerateAgoraStreamToken(channelID, userID, vc_domain.RolePublisher)
		if err != nil {
			return nil, fmt.Errorf("error retrieve broadcast token (user:%s, channel:%s): AgoraTokenSrv.GenerateAgoraStreamToken: %w", userID, channelID, err)
		}
		if streamToken == "" {
			return nil, fmt.Errorf("cannot get token for room uuid: %q", liveRoom.WhiteboardRoomID)
		}

		videoToken, err = l.AgoraTokenSvc.GenerateAgoraStreamToken(channelID, userID+l.VideoTokenSuffix, vc_domain.RolePublisher)
		if err != nil {
			return nil, fmt.Errorf("error retrieve video broadcast token (user:%s, channel:%s): AgoraTokenSrv.GenerateAgoraStreamToken: %w", userID+l.VideoTokenSuffix, channelID, err)
		}

		shareForRecordingToken, err = l.AgoraTokenSvc.GenerateAgoraStreamToken(channelID, userID+"-streamforcloudrecording", vc_domain.RolePublisher)
		if err != nil {
			return nil, fmt.Errorf("could not generate token streamforcloudrecording (user:%s, channel:%s): AgoraTokenSrv.GenerateAgoraStreamToken: %w", userID, channelID, err)
		}
	}

	rtmToken, err = l.AgoraTokenSvc.BuildRTMTokenByUserID(rtmUserID)
	if err != nil {
		return nil, fmt.Errorf("could not generate RTM token (user:%s, rtm_user_id: %s): AgoraTokenSrv.BuildRTMToken: %w", userID, rtmUserID, err)
	}

	return &JoinLiveRoomResponse{
		ChannelID:            liveRoom.ChannelID,
		RoomID:               liveRoom.WhiteboardRoomID,
		StreamToken:          streamToken,
		WhiteboardToken:      whiteboardToken,
		VideoToken:           videoToken,
		StmToken:             rtmToken,
		ScreenRecordingToken: shareForRecordingToken,
		UserGroup:            userGroup, // user group return is only either teacher or student
	}, nil
}

func (l *LiveRoomCommand) createChannel(ctx context.Context, channelName string) (*domain.LiveRoom, error) {
	liveRoom, err := l.LiveRoomRepo.GetLiveRoomByChannelName(ctx, l.LessonmgmtDB, channelName)
	if err != nil && err != domain.ErrChannelNotFound {
		return nil, fmt.Errorf("error in LiveRoomRepo.GetLiveRoomByChannelName, channel %s: %w", channelName, err)
	} else if err == domain.ErrChannelNotFound {
		channelID := idutil.ULIDNow()

		room, err := l.WhiteboardSvc.CreateRoom(ctx, &whiteboard.CreateRoomRequest{
			Name:     channelID,
			IsRecord: false,
		})
		if err != nil {
			return nil, fmt.Errorf("could not create a new whiteboard room for channel %s using new ID %s: %v", channelName, channelID, err)
		}

		err = l.LiveRoomRepo.CreateLiveRoom(ctx, l.LessonmgmtDB, channelID, channelName, room.UUID)
		if err != nil && err != domain.ErrNoChannelCreated {
			return nil, fmt.Errorf("error in LiveRoomRepo.CreateLiveRoom, channel name %s: %w", channelName, err)
		}

		liveRoom, err = l.LiveRoomRepo.GetLiveRoomByChannelName(ctx, l.LessonmgmtDB, channelName)
		if err != nil {
			return nil, fmt.Errorf("error in LiveRoomRepo.GetLiveRoomByChannelName, second attempt to fetch channel %s: %w", channelName, err)
		}
	}

	return liveRoom, nil
}

func (l *LiveRoomCommand) getWhiteboardToken(ctx context.Context, channelID, roomID string) (whiteboardToken, roomUUID string, err error) {
	// generate room ID
	roomUUID = roomID
	if len(roomUUID) == 0 {
		room, err := l.WhiteboardSvc.CreateRoom(ctx, &whiteboard.CreateRoomRequest{
			Name:     channelID,
			IsRecord: false,
		})
		if err != nil {
			return whiteboardToken, roomUUID, fmt.Errorf("could not create a new whiteboard room for channel ID %s: %v", channelID, err)
		}
		roomUUID = room.UUID

		if err = l.LiveRoomRepo.UpdateChannelRoomID(ctx, l.LessonmgmtDB, channelID, roomUUID); err != nil {
			return whiteboardToken, roomUUID, fmt.Errorf("error in LiveRoomRepo.UpdateChannelRoomID, channel %s: %w", channelID, err)
		}
	}

	// get whiteboard token
	retryCount := 0
	for {
		retryCount++
		whiteboardToken, err = l.WhiteboardSvc.FetchRoomToken(ctx, roomUUID)
		if err == nil || retryCount > 5 {
			break
		}
		ctxzap.Extract(ctx).Warn("cannot fetch whiteboard room token", zap.Error(err))

		time.Sleep(time.Duration(200*retryCount) * time.Millisecond)
		ctxzap.Extract(ctx).Warn(fmt.Sprintf("retry fetch whiteboard room token %d time", retryCount))
	}
	if err != nil {
		return whiteboardToken, roomUUID, fmt.Errorf("cannot fetch whiteboard room token from room ID %s: %q", roomUUID, err)
	}

	return whiteboardToken, roomUUID, nil
}

func (l *LiveRoomCommand) EndLiveRoom(ctx context.Context, channelID, lessonID string) error {
	now := time.Now()
	if err := l.LiveRoomRepo.EndLiveRoom(ctx, l.LessonmgmtDB, channelID, now); err != nil {
		return fmt.Errorf("error in LiveRoomRepo.EndLiveRoom, channel %s: %w", channelID, err)
	}

	if len(strings.TrimSpace(lessonID)) > 0 {
		if err := l.LessonRepo.EndLiveLesson(ctx, l.LessonmgmtDB, lessonID, now); err != nil && !strings.Contains(err.Error(), "cannot update lesson") {
			return fmt.Errorf("error in LessonRepo.EndLiveLesson, lessonID %s: %w", lessonID, err)
		}
	}

	return nil
}

func (l *LiveRoomCommand) CreateAndGetChannelInfo(ctx context.Context, channelName string) (*payloads.CreateAndGetChannelInfoResponse, error) {
	liveRoom, err := l.createChannel(ctx, channelName)
	if err != nil {
		return nil, err
	}

	whiteboardToken, roomID, err := l.getWhiteboardToken(ctx, liveRoom.ChannelID, liveRoom.WhiteboardRoomID)
	if err != nil {
		return nil, err
	}
	liveRoom.WhiteboardRoomID = roomID

	return &payloads.CreateAndGetChannelInfoResponse{
		ChannelID:       liveRoom.ChannelID,
		RoomID:          liveRoom.WhiteboardRoomID,
		WhiteboardAppID: l.WhiteboardAppID,
		WhiteboardToken: whiteboardToken,
	}, nil
}

func (l *LiveRoomCommand) PreparePublishLiveRoom(ctx context.Context, channelID, learnerID string) (vc_domain.PrepareToPublishStatus, error) {
	publishStatus := vc_domain.PublishStatusNone

	err := database.ExecInTx(ctx, l.LessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
		learnerIDs, err := l.LiveRoomStateRepo.GetStreamingLearners(ctx, tx, channelID, true)
		if err != nil && err != domain.ErrChannelNotFound {
			return fmt.Errorf("error in LiveRoomStateRepo.GetStreamingLearners, channel %s: %w", channelID, err)
		}
		if err != domain.ErrChannelNotFound && sliceutils.Contains(learnerIDs, learnerID) {
			publishStatus = vc_domain.PublishStatusPreparedBefore
			return fmt.Errorf("prepared before")
		}

		if err := l.LiveRoomStateRepo.IncreaseNumberOfStreaming(ctx, tx, channelID, learnerID, l.MaximumLearnerStreamings); err != nil {
			if err == domain.ErrNoChannelUpdated {
				publishStatus = vc_domain.PublishStatusMaxLimit
				return err
			}
			return fmt.Errorf("error in LiveRoomStateRepo.IncreaseNumberOfStreaming, channel %s learner %s: %w", channelID, learnerID, err)
		}

		if err := l.LiveRoomActivityLogRepo.CreateLog(ctx, tx, channelID, learnerID, constant.LogActionTypePublish); err != nil {
			return fmt.Errorf("error in LiveRoomActivityLogRepo.CreateLog, channel %s user %s: %w", channelID, learnerID, err)
		}

		return nil
	})
	if err != nil && err != domain.ErrNoChannelUpdated && err.Error() != "prepared before" {
		return "", fmt.Errorf("ExecInTx: %w", err)
	}

	return publishStatus, nil
}

func (l *LiveRoomCommand) UnpublishLiveRoom(ctx context.Context, channelID, learnerID string) (vc_domain.UnpublishStatus, error) {
	unpublishStatus := vc_domain.UnpublishStatsNone

	err := database.ExecInTx(ctx, l.LessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
		if err := l.LiveRoomStateRepo.DecreaseNumberOfStreaming(ctx, tx, channelID, learnerID); err != nil {
			if err == domain.ErrNoChannelUpdated {
				unpublishStatus = vc_domain.UnpublishStatsUnpublishedBefore
				return err
			}
			return fmt.Errorf("error in LiveRoomStateRepo.DecreaseNumberOfStreaming, channel %s learner %s: %w", channelID, learnerID, err)
		}

		if err := l.LiveRoomActivityLogRepo.CreateLog(ctx, tx, channelID, learnerID, constant.LogActionTypeUnpublish); err != nil {
			return fmt.Errorf("error in LiveRoomActivityLogRepo.CreateLog, channel %s user %s: %w", channelID, learnerID, err)
		}

		return nil
	})
	if err != nil && err != domain.ErrNoChannelUpdated {
		return "", fmt.Errorf("ExecInTx: %w", err)
	}

	return unpublishStatus, nil
}
